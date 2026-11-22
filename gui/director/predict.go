package director

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"time"
)

// Model is a logistic regression that estimates, for each living player, the
// probability of getting a kill in the next 3 seconds. It was trained offline
// on real demos with tools/replay (see tools/replay/README.md).
type Model struct {
	Features  []string  `json:"features"`
	Coef      []float64 `json:"coef"`
	Intercept float64   `json:"intercept"`
}

// Prob returns P(kill in next 3s) for one feature vector
func (m *Model) Prob(f []float64) float64 {
	z := m.Intercept
	for i := range f {
		if i < len(m.Coef) {
			z += m.Coef[i] * f[i]
		}
	}
	return 1 / (1 + math.Exp(-z))
}

// LoadModel reads a model JSON file written by tools/replay/train.py
func LoadModel(path string) (*Model, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Model
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// PredictConfig controls the predictive camera during live rounds
type PredictConfig struct {
	Enabled       bool
	MinDwell      time.Duration // min time on a player before a normal switch
	Margin        float64       // new player's probability must exceed current's by this
	FastP         float64       // a player this likely to kill can take the camera sooner...
	FastDwell     time.Duration // ...after this much time on the current player
	Smooth        float64       // EMA factor for probabilities (0 = no smoothing)
	EventSwitches bool          // keep the old immediate switches (kill, damage, upset, AWP)
}

// DefaultPredictConfig is tuned with tools/replay on real demos
func DefaultPredictConfig() PredictConfig {
	return PredictConfig{
		Enabled:       !PredictionDisabled,
		MinDwell:      3 * time.Second,
		Margin:        0.05,
		FastP:         1.0, // fast switch off: made no difference in the demos
		FastDwell:     1500 * time.Millisecond,
		Smooth:        0.5,
		EventSwitches: false,
	}
}

// predictiveChoice returns the player to switch to ("" = stay) using the model.
func (pa *PlayerAnalyzer) predictiveChoice(players []PlayerInfo) string {
	if pa.featureState == nil {
		pa.featureState = NewFeatureState()
		pa.probs = map[string]float64{}
	}
	feats := pa.featureState.Compute(players, Now())
	cfg := pa.Predict

	ids := make([]string, 0, len(feats))
	for id := range feats {
		ids = append(ids, id)
	}
	sort.Strings(ids) // deterministic tie-breaks
	newProbs := map[string]float64{}
	for _, id := range ids {
		p := pa.Model.Prob(feats[id])
		if old, ok := pa.probs[id]; ok && cfg.Smooth > 0 {
			p = cfg.Smooth*old + (1-cfg.Smooth)*p
		}
		newProbs[id] = p
	}
	pa.probs = newProbs

	best, bestP := "", -1.0
	for _, id := range ids {
		if newProbs[id] > bestP {
			best, bestP = id, newProbs[id]
		}
	}
	if best == "" {
		return ""
	}
	cur := pa.currentSpectatedID
	curP, curAlive := newProbs[cur]
	if cur == "" || !curAlive {
		return best
	}
	if best == cur {
		return ""
	}
	dwell := since(pa.currentPlayerSwitchTime)
	switch {
	case dwell >= cfg.MinDwell && bestP >= curP+cfg.Margin:
		LogVerbose("[PREDICT] %s %.2f > current %.2f", getPlayerNameByID(best, players), bestP, curP)
		return best
	case dwell >= cfg.FastDwell && bestP >= cfg.FastP && curP < cfg.FastP/2:
		LogVerbose("[PREDICT] fast: %s %.2f vs current %.2f", getPlayerNameByID(best, players), bestP, curP)
		return best
	case dwell >= pa.maxStickyTime && bestP > curP:
		return best
	}
	return ""
}

// Probabilities returns the last predicted kill probability per player (for UI/debug)
func (pa *PlayerAnalyzer) Probabilities() map[string]float64 { return pa.probs }
