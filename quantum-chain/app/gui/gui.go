package gui

import (
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

//go:embed index.html
var content embed.FS

// netInfoClient is the subset of the CometBFT RPC client needed for peer info.
// The node's in-process client implements it, even though the narrower
// client.CometRPC interface exposed by client.Context does not declare NetInfo.
type netInfoClient interface {
	NetInfo(context.Context) (*coretypes.ResultNetInfo, error)
}

type nodeInfo struct {
	ID      string `json:"id"`
	Moniker string `json:"moniker"`
	Height  string `json:"height,omitempty"`
}

type peerInfo struct {
	ID       string `json:"id"`
	Moniker  string `json:"moniker"`
	IP       string `json:"ip"`
	Outbound bool   `json:"outbound"`
}

type networkResponse struct {
	Self   nodeInfo   `json:"self"`
	Peers  []peerInfo `json:"peers"`
	NPeers int        `json:"n_peers"`
}

// SeedEntry is a single round's network-shared random seed.
type SeedEntry struct {
	RoundID uint64 `json:"round_id"`
	Seed    string `json:"seed"`
}

// SeedProvider returns all finalized shared-random seeds. The app supplies an
// implementation that reads the beacon module state.
type SeedProvider func() ([]SeedEntry, error)

type seedsResponse struct {
	Seeds []SeedEntry `json:"seeds"`
}

// ProblemEntry is a registered daqq problem (a shared-data protocol the network
// can solve), exposed to the visualizer.
type ProblemEntry struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// ProblemProvider returns the registered problems, lowest ID first. The app
// supplies an implementation that reads the problems module registry.
type ProblemProvider func() ([]ProblemEntry, error)

type problemsResponse struct {
	Problems []ProblemEntry `json:"problems"`
}

// ResultSubmission is one participant's submitted solution for a round. The
// distribution itself is summarized by a hash so the payload stays small and
// agreement between participants is a simple hash comparison.
type ResultSubmission struct {
	Address          string `json:"address"`
	BlockHeight      int64  `json:"block_height"`
	NumStates        int    `json:"num_states"`
	DistributionHash string `json:"distribution_hash"`
}

// ResultRound groups every participant's submission for a single round and flags
// whether they all agree — the on-chain cross-validation signal: independent
// participants that solved the same shared-seed circuit must produce an
// identical distribution.
type ResultRound struct {
	RoundID     uint64             `json:"round_id"`
	Submissions []ResultSubmission `json:"submissions"`
	Agreement   bool               `json:"agreement"`
}

// ResultsProvider returns submitted results grouped by round. The app supplies
// an implementation that reads the random_circuit module state.
type ResultsProvider func() ([]ResultRound, error)

type resultsResponse struct {
	Results []ResultRound `json:"results"`
}

// RegisterGUIService serves the embedded web visualizer and same-origin data
// endpoints from the node's own API server, so the browser never makes a
// cross-origin request and no CORS configuration is required.
func RegisterGUIService(clientCtx client.Context, seeds SeedProvider, problems ProblemProvider, results ResultsProvider, rtr *mux.Router) {
	rtr.HandleFunc("/gui/net_info", netInfoHandler(clientCtx))
	rtr.HandleFunc("/gui/seeds", seedsHandler(seeds))
	rtr.HandleFunc("/gui/problems", problemsHandler(problems))
	rtr.HandleFunc("/gui/results", resultsHandler(results))
	// ラウンド進行状況カード(index.html の #progress-card)は専用エンドポイント
	// を増やさず、/gui/net_info が返すブロック高(self.height)からクライアント側
	// で算出している。50ブロック周期は固定なので round_id = height / 50,
	// offset = height % 50, phase: commit 0-30 / reveal 30-45 / final 45-50 を
	// JS 側で計算し、ブロック時間を実測して各ブロック間を補間して表示する。
	rtr.Handle("/gui", http.RedirectHandler("/gui/", http.StatusMovedPermanently))
	rtr.PathPrefix("/gui/").Handler(http.StripPrefix("/gui/", http.FileServer(http.FS(content))))
}

// seedsHandler reports the network-shared random seeds, newest round first.
func seedsHandler(seeds SeedProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := seedsResponse{Seeds: []SeedEntry{}}
		if seeds != nil {
			if list, err := seeds(); err == nil {
				resp.Seeds = list
				sort.Slice(resp.Seeds, func(i, j int) bool {
					return resp.Seeds[i].RoundID > resp.Seeds[j].RoundID
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// problemsHandler reports the registered daqq problems, lowest ID first.
func problemsHandler(problems ProblemProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := problemsResponse{Problems: []ProblemEntry{}}
		if problems != nil {
			if list, err := problems(); err == nil {
				resp.Problems = list
				sort.Slice(resp.Problems, func(i, j int) bool {
					return resp.Problems[i].ID < resp.Problems[j].ID
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// resultsHandler reports submitted results grouped by round, newest round
// first, so the visualizer can show per-round cross-validation (which
// participants submitted and whether their distributions agree).
func resultsHandler(results ResultsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := resultsResponse{Results: []ResultRound{}}
		if results != nil {
			if list, err := results(); err == nil {
				resp.Results = list
				sort.Slice(resp.Results, func(i, j int) bool {
					return resp.Results[i].RoundID > resp.Results[j].RoundID
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// netInfoHandler reports this node and its connected peers by calling the
// in-process CometBFT RPC client directly.
func netInfoHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp := networkResponse{Peers: []peerInfo{}}

		if status, err := clientCtx.Client.Status(ctx); err == nil {
			resp.Self = nodeInfo{
				ID:      string(status.NodeInfo.DefaultNodeID),
				Moniker: status.NodeInfo.Moniker,
				Height:  strconv.FormatInt(status.SyncInfo.LatestBlockHeight, 10),
			}
		}

		if nic, ok := clientCtx.Client.(netInfoClient); ok {
			if ni, err := nic.NetInfo(ctx); err == nil {
				resp.NPeers = ni.NPeers
				for _, p := range ni.Peers {
					resp.Peers = append(resp.Peers, peerInfo{
						ID:       string(p.NodeInfo.DefaultNodeID),
						Moniker:  p.NodeInfo.Moniker,
						IP:       p.RemoteIP,
						Outbound: p.IsOutbound,
					})
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
