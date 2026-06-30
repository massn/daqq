package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"sync"
	"time"

	"github.com/massn/daqq/circuit"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

// Protocol constants for the random_circuit case-A problem. Width (qubits) and
// Depth (layers) are fixed so every participant builds the identical circuit
// from a round's shared seed; only then do independently computed distributions
// agree and become cross-validatable on-chain.
const (
	CircuitWidth = 5
	CircuitDepth = 10
)

// PollInterval is how often the polling fallback re-checks the latest round.
// The event subscription is the fast path; polling is the liveness safety net
// in case the websocket subscription silently stops delivering events.
const PollInterval = 30 * time.Second

// config is the participant's connection + signing setup, read from the
// environment so the same binary can run as any participant against any node
// (the defaults target a local single-node dev chain with the test keyring).
type config struct {
	Bin            string // quantumchaind binary name/path
	RPC            string // CometBFT RPC endpoint
	API            string // REST/gui endpoint for the polling fallback (/gui/seeds)
	ChainID        string
	From           string // key name to sign submissions with
	KeyringBackend string
	Home           string // node home dir (empty = quantumchaind default)
}

func loadConfig() config {
	env := func(k, def string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return def
	}
	return config{
		Bin:            env("QC_BIN", "quantumchaind"),
		RPC:            env("QC_RPC", "tcp://localhost:26657"),
		API:            env("QC_API", "http://localhost:1317"),
		ChainID:        env("QC_CHAIN_ID", "quantum-chain"),
		From:           env("QC_FROM", "alice"),
		KeyringBackend: env("QC_KEYRING_BACKEND", "test"),
		Home:           env("QC_HOME", ""),
	}
}

var cfg config

// processed deduplicates rounds across the two ingestion paths (event
// subscription and polling fallback) so a round is only submitted once.
var (
	processedMu sync.Mutex
	processed   = map[string]bool{}
)

// claim marks roundID as handled and returns true the first time it is seen.
func claim(roundID string) bool {
	processedMu.Lock()
	defer processedMu.Unlock()
	if processed[roundID] {
		return false
	}
	processed[roundID] = true
	return true
}

func main() {
	cfg = loadConfig()

	// Polling fallback: independent of the event subscription, periodically
	// check the latest round and submit any we have not handled yet. This keeps
	// the participant live even if the websocket subscription dies silently.
	go pollLoop()

	// Event subscription: low-latency fast path. Re-subscribe on disconnect so a
	// dropped websocket does not permanently stop the fast path.
	subscribeLoop()
}

// subscribeLoop connects to the node and consumes new_round events, reconnecting
// with a short backoff if the connection or subscription channel drops.
func subscribeLoop() {
	for {
		if err := subscribeOnce(); err != nil {
			log.Printf("event subscription error: %v (retrying in %s; polling fallback still active)", err, PollInterval)
		} else {
			log.Printf("event subscription closed; resubscribing in %s (polling fallback still active)", PollInterval)
		}
		time.Sleep(PollInterval)
	}
}

func subscribeOnce() error {
	client, err := rpchttp.New(cfg.RPC, "/websocket")
	if err != nil {
		return err
	}
	if err := client.Start(); err != nil {
		return err
	}
	defer client.Stop()

	query := "new_round.round_id EXISTS"
	out, err := client.Subscribe(context.Background(), "qc-client-"+cfg.From, query)
	if err != nil {
		return err
	}

	log.Printf("Listening for new_round events as %q (chain %s, node %s)...", cfg.From, cfg.ChainID, cfg.RPC)

	for e := range out {
		processEvent(e)
	}
	// Channel closed: the subscription ended (e.g. websocket dropped). Returning
	// nil lets subscribeLoop resubscribe.
	return nil
}

// pollLoop periodically fetches the latest round from the gui endpoint and
// submits it if not already handled. Mirrors the Qiskit reference client.
func pollLoop() {
	for {
		time.Sleep(PollInterval)
		roundID, seedHex, err := latestRound()
		if err != nil {
			log.Printf("poll error: %v", err)
			continue
		}
		if roundID == "" || seedHex == "" {
			continue
		}
		handleRound(roundID, seedHex)
	}
}

// latestRound reads the most recent (round_id, seed) from QC_API + /gui/seeds.
func latestRound() (string, string, error) {
	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(cfg.API + "/gui/seeds")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var payload struct {
		Seeds []struct {
			RoundID json.Number `json:"round_id"`
			Seed    string      `json:"seed"`
		} `json:"seeds"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", err
	}
	if len(payload.Seeds) == 0 {
		return "", "", nil
	}
	return payload.Seeds[0].RoundID.String(), payload.Seeds[0].Seed, nil
}

func processEvent(e coretypes.ResultEvent) {
	events := e.Events
	roundIDs := events["new_round.round_id"]
	seeds := events["new_round.seed"]

	if len(roundIDs) == 0 || len(seeds) == 0 {
		return
	}
	handleRound(roundIDs[0], seeds[0])
}

// handleRound builds the round's circuit and submits its distribution, exactly
// once per round_id regardless of which path (event or poll) discovered it.
func handleRound(roundIDStr, seedHex string) {
	if !claim(roundIDStr) {
		return
	}

	log.Printf("New Round: %s, Seed: %s", roundIDStr, seedHex)

	// 3. Generate Circuit
	seedBytes, _ := hex.DecodeString(seedHex) // Assuming strict hex
	if len(seedBytes) == 0 {
		log.Printf("Empty seed: %s", seedHex)
		return
	}
	// Fold the WHOLE shared seed (all 256 bits) into the PRNG state via FNV-1a,
	// so every byte of the beacon's random selects the circuit (not just the
	// first four bytes). circuit.SeedState mirrors the GUI's seedState() in JS.
	seed := circuit.SeedState(seedBytes)

	// Build this round's case-A circuit from the shared seed. Width/Depth are
	// fixed protocol constants, so every participant constructs the identical
	// circuit.
	qc := circuit.MakeRandomQC(seed, CircuitWidth, CircuitDepth)

	// 4. Compute the exact theoretical (case A) distribution and submit it.
	//
	// We submit the analytical distribution (not Monte-Carlo shot counts) so the
	// result is fully deterministic: every honest participant solving this
	// round's circuit produces a byte-identical distribution, which is what makes
	// the on-chain results cross-validatable. (Empirical shot counts — case B —
	// will live in a future `random_circuit_sampling` module.)
	probabilities, err := qc.Distribution()
	if err != nil {
		log.Printf("Failed to compute distribution: %v", err)
		return
	}

	// Encode as a list of (state, probability) entries sorted by state. The
	// Distribution message is a repeated list (not a proto map, which the Cosmos
	// SDK tx codec rejects), and sorting makes the encoded message byte-identical
	// across participants.
	states := make([]string, 0, len(probabilities))
	for s := range probabilities {
		states = append(states, s)
	}
	sort.Strings(states)
	entries := make([]map[string]string, 0, len(states))
	for _, s := range states {
		entries = append(entries, map[string]string{"state": s, "probability": probabilities[s]})
	}
	distribution := map[string]any{"entries": entries}
	distributionJSON, _ := json.Marshal(distribution)

	// 5. Submit Result via CLI.
	// quantum-chaind tx random_circuit submit-result [round-id] --distribution <json> ...
	args := []string{
		"tx", "random_circuit", "submit-result",
		roundIDStr,
		"--distribution", string(distributionJSON),
		"--from", cfg.From,
		"--chain-id", cfg.ChainID,
		"--keyring-backend", cfg.KeyringBackend,
		"--node", cfg.RPC,
		"-y",
	}
	if cfg.Home != "" {
		args = append(args, "--home", cfg.Home)
	}

	cmd := exec.Command(cfg.Bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Submitting results for round %s...", roundIDStr)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to submit result: %v", err)
		// A failed submission (e.g. transient RPC error) should be retryable, so
		// un-claim the round and let the next poll pick it up again.
		processedMu.Lock()
		delete(processed, roundIDStr)
		processedMu.Unlock()
	}
}
