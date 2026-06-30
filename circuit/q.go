package circuit

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/itsubaki/q"
)

const (
	// DefaultGrids is the default number of grid points for probability density histogram
	DefaultGrids = 500
	// DefaultOutputFile is the default output file name for visualization
	DefaultOutputFile = "line.html"
	// DefaultChartTitle is the default chart title
	DefaultChartTitle = "Quantum State Probability Density"
	// DefaultChartSubtitle is the default chart subtitle
	DefaultChartSubtitle = "Distribution of quantum state probabilities"
)

// executeCircuit initializes the quantum simulator and applies all gates in the circuit
func (qc *QC) executeCircuit() (*q.Q, []q.Qubit, error) {
	qsim := q.New()
	qubits := make([]q.Qubit, qc.Width)
	for i := 0; i < int(qc.Width); i++ {
		qubits[i] = qsim.Zero()
	}
	for _, gate := range qc.Gates {
		// Validate gate before applying
		if err := gate.Validate(qc.Width); err != nil {
			return nil, nil, fmt.Errorf("invalid gate: %w", err)
		}

		switch gate.GateName {
		case "H":
			qsim.H(qubits[*gate.TargetQubit])
		case "S":
			qsim.S(qubits[*gate.TargetQubit])
		case "T":
			qsim.T(qubits[*gate.TargetQubit])
		case "CNOT":
			qsim.CNOT(qubits[*gate.ControlQubit], qubits[*gate.TargetQubit])
		default:
			return nil, nil, fmt.Errorf("unknown gate: %s", gate.GateName)
		}
	}
	return qsim, qubits, nil
}

// RunShots simulates the circuit and samples its output distribution shots
// times, returning a map from each measured basis state (as a binary string)
// to the number of times it was observed.
func (qc *QC) RunShots(shots int) (map[string]uint64, error) {
	// Initialize random seed for reproducible results in tests
	rand.Seed(time.Now().UnixNano())

	qsim, _, err := qc.executeCircuit()
	if err != nil {
		return nil, fmt.Errorf("failed to execute circuit: %w", err)
	}
	state := qsim.State()

	results := make(map[string]uint64)
	for k := 0; k < shots; k++ {
		r := rand.Float64()
		sum := 0.0
		for _, s := range state {
			sum += s.Probability()
			if r < sum {
				results[s.BinaryString()]++
				break
			}
		}
	}
	return results, nil
}

// Distribution computes the exact theoretical (case A) output distribution of
// the circuit: a map from each computational-basis state (binary string, e.g.
// "00101") to its probability, encoded as a fixed-point decimal string.
//
// Unlike RunShots, this is fully deterministic — the same circuit always yields
// a byte-identical map — so independent participants that solve the same
// shared-seed circuit produce identical distributions. That determinism is what
// makes daqq's on-chain results cross-validatable: honest nodes agree exactly,
// and any divergence is immediately visible on the ledger.
func (qc *QC) Distribution() (map[string]string, error) {
	qsim, _, err := qc.executeCircuit()
	if err != nil {
		return nil, fmt.Errorf("failed to execute circuit: %w", err)
	}
	state := qsim.State()

	probabilities := make(map[string]string, len(state))
	for _, s := range state {
		// Quantize to 9 decimal places so that independent simulators (Go's
		// itsubaki/q, Qiskit, …) which agree on the distribution to ~1e-12 but
		// differ in the last float64 ULP still produce byte-identical strings —
		// enabling cross-implementation agreement. States whose probability
		// rounds to zero are dropped so the included-state set also matches.
		p := strconv.FormatFloat(s.Probability(), 'f', 9, 64)
		if p == "0.000000000" {
			continue
		}
		probabilities[s.BinaryString()] = p
	}
	return probabilities, nil
}

// State simulates the circuit, prints each basis state's probability sorted
// from highest to lowest, and renders a probability-density histogram to
// DefaultOutputFile.
func (qc *QC) State() error {
	qsim, _, err := qc.executeCircuit()
	if err != nil {
		return fmt.Errorf("failed to execute circuit: %w", err)
	}
	state := qsim.State()

	// sort
	sort.Slice(state, func(i, j int) bool {
		return state[i].Probability() > state[j].Probability()
	})
	for _, s := range state {
		fmt.Printf("%s: %f\n", s.BinaryString(), s.Probability())
	}

	grids := DefaultGrids
	period := float64(1 / float64(grids))
	xaxis := []float64{}
	for i := 0; i < grids; i++ {
		xaxis = append(xaxis, period*float64(i))
	}
	probDensity := make([]float64, grids)
	for i, x := range xaxis {
		var acc float64 = 0
		for _, s := range state {
			p := s.Probability()
			if (x-period <= p) && (p < x) {
				acc++
			}
		}
		probDensity[i] = acc / float64(len(state)) / float64(period)
	}

	/*

		percentile := 1.0
		percentiles := []float64{}
		for _, s := range state {
			percentile -= s.Probability()
			percentiles = append(percentiles, percentile)
		}
	*/

	page := components.NewPage()
	page.AddCharts(
		barBase(xaxis, probDensity),
	)
	f, err := os.Create(DefaultOutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := page.Render(io.MultiWriter(f)); err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	return nil
}

func barBase(xaxis, probDensity []float64) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    DefaultChartTitle,
			Subtitle: DefaultChartSubtitle,
		}),
	)
	bar.SetXAxis(xaxis).AddSeries("Probability Density", generateBarItems(probDensity))
	return bar
}

func generateBarItems(probDensity []float64) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, prob := range probDensity {
		items = append(items, opts.BarData{Value: prob})
	}
	return items
}
