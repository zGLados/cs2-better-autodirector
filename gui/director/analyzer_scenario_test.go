package director

// Scenario tests built from real logs. Run with: go test -run . -v .

import (
	"testing"
	"time"
)

type pl struct {
	id, name, team, pos, weapon string
	kills                       int
}

func state(ps []pl) map[string]interface{} {
	all := map[string]interface{}{}
	for _, p := range ps {
		all[p.id] = map[string]interface{}{
			"name": p.name, "team": p.team, "position": p.pos,
			"state":       map[string]interface{}{"health": 100.0, "equip_value": 4500.0},
			"match_stats": map[string]interface{}{"kills": float64(p.kills)},
			"weapons":     map[string]interface{}{"weapon_0": map[string]interface{}{"name": p.weapon, "state": "active"}},
		}
	}
	return map[string]interface{}{"allplayers": all}
}

// Positions copied from the user's log (Inferno, 19:51:43 and 19:51:44)
func TestInfernoDoubleKill(t *testing.T) {
	LogLevel = 0
	pa := NewPlayerAnalyzer()
	pa.Settings = NewDefaultSettings()
	pa.Predict.Enabled = false // tests the legacy (fallback) logic
	pa.SyncCurrentPlayer("priest")
	pa.lastSwitchTime = time.Now().Add(-10 * time.Second)
	pa.currentPlayerSwitchTime = time.Now().Add(-9 * time.Second)

	tick1 := []pl{
		{"lzer4", "lzer4", "CT", "2150.60, 1.63, 122.34", "weapon_m4a1", 4},
		{"gabriel", "gabriel", "T", "1265.68, -352.75, 260.03", "weapon_ak47", 1},
		{"priest", "priest", "T", "1745.06, -363.03, 256.03", "weapon_ak47", 7},
		{"LR", "LR", "CT", "442.15, 2579.83, 161.53", "weapon_m4a1", 1},
		{"FF", "FF", "T", "1097.24, 532.25, 124.62", "weapon_ak47", 3},
		{"Bastos", "Bastos", "CT", "1376.51, 1227.97, 181.03", "weapon_m4a1", 3},
		{"Bllizzter", "Bllizzter", "CT", "1089.01, 2779.91, 126.84", "weapon_m4a1", 1},
		{"sX", "s.X", "T", "1477.23, 322.75, 128.03", "weapon_ak47", 2},
		{"brz", "brz", "CT", "1512.95, 244.06, 133.03", "weapon_m4a1", 7},
	}
	got := pa.GetBestPlayerToSpectate(state(tick1), "live")
	t.Logf("tick 19:51:43 -> switch to %q (current %s)", got, pa.currentSpectatedID)

	// 19:51:44: s.X killed brz (brz gone, s.X kills 2->3)
	tick2 := []pl{
		{"FF", "FF", "T", "1249.08, 482.93, 117.49", "weapon_ak47", 3},
		{"lzer4", "lzer4", "CT", "2186.87, 31.77, 122.05", "weapon_m4a1", 4},
		{"Bastos", "Bastos", "CT", "1434.11, 1228.13, 176.03", "weapon_m4a1", 3},
		{"Bllizzter", "Bllizzter", "CT", "1089.01, 2779.91, 126.84", "weapon_m4a1", 1},
		{"gabriel", "gabriel", "T", "1324.39, -343.06, 259.00", "weapon_ak47", 1},
		{"priest", "priest", "T", "1750.57, -368.69, 256.03", "weapon_ak47", 7},
		{"sX", "s.X", "T", "1434.65, 261.21, 128.03", "weapon_ak47", 3},
		{"LR", "LR", "CT", "442.15, 2579.83, 161.53", "weapon_m4a1", 1},
	}
	got = pa.GetBestPlayerToSpectate(state(tick2), "live")
	t.Logf("tick 19:51:44 (s.X kills brz) -> switch to %q", got)
	if got != "sX" {
		t.Fatalf("expected camera to follow s.X after his kill, got %q", got)
	}

	// 0.5s later, before his 2nd kill: camera must stay on s.X
	pa.lastSwitchTime = time.Now().Add(-500 * time.Millisecond)
	tick3 := []pl{
		{"sX", "s.X", "T", "1459.83, 196.21, 130.87", "weapon_ak47", 3},
		{"LR", "LR", "CT", "442.15, 2579.83, 161.53", "weapon_m4a1", 1},
		{"lzer4", "lzer4", "CT", "2212.73, 53.35, 122.24", "weapon_m4a1", 4},
		{"Bastos", "Bastos", "CT", "1434.11, 1228.13, 176.03", "weapon_m4a1", 3},
		{"gabriel", "gabriel", "T", "1378.86, -352.19, 256.03", "weapon_ak47", 1},
		{"priest", "priest", "T", "1750.20, -368.45, 256.03", "weapon_ak47", 7},
		{"Bllizzter", "Bllizzter", "CT", "1089.01, 2779.91, 126.84", "weapon_m4a1", 1},
		{"FF", "FF", "T", "1321.05, 451.99, 123.93", "weapon_ak47", 3},
	}
	got = pa.GetBestPlayerToSpectate(state(tick3), "live")
	t.Logf("tick 19:51:44.5 -> %q, current=%s", got, pa.currentSpectatedID)
	if pa.currentSpectatedID != "sX" {
		t.Fatalf("camera left s.X before his second kill")
	}
}

// Aim-based damage attribution: nearest enemy is behind a wall looking away,
// the real shooter is farther but aiming at the victim.
func TestDamageAttributionByAim(t *testing.T) {
	LogLevel = 0
	pa := NewPlayerAnalyzer()
	pa.Settings = NewDefaultSettings()
	pa.Predict.Enabled = false // tests the legacy (fallback) logic
	mk := func(hp float64, near, far string) map[string]interface{} {
		return map[string]interface{}{"allplayers": map[string]interface{}{
			"victim": map[string]interface{}{"name": "victim", "team": "CT", "position": "100, 100, 0",
				"state": map[string]interface{}{"health": hp}},
			"near": map[string]interface{}{"name": "near", "team": "T", "position": "300, 100, 0", "forward": near,
				"state":   map[string]interface{}{"health": 100.0},
				"weapons": map[string]interface{}{"w": map[string]interface{}{"name": "weapon_ak47", "state": "active"}}},
			"far": map[string]interface{}{"name": "far", "team": "T", "position": "100, 1000, 0", "forward": far,
				"state":   map[string]interface{}{"health": 100.0},
				"weapons": map[string]interface{}{"w": map[string]interface{}{"name": "weapon_ak47", "state": "active"}}},
		}}
	}
	pa.SyncCurrentPlayer("near")
	pa.lastSwitchTime = time.Now().Add(-5 * time.Second)
	pa.GetBestPlayerToSpectate(mk(100, "1, 0, 0", "0, -1, 0"), "live")
	pa.lastSwitchTime = time.Now().Add(-5 * time.Second)
	got := pa.GetBestPlayerToSpectate(mk(60, "1, 0, 0", "0, -1, 0"), "live")
	if got != "far" {
		t.Fatalf("expected damage credited to the aiming player 'far', got %q", got)
	}
}

// Anti ping-pong: a slightly better encounter elsewhere must not steal the
// camera (log 20:35:45-20:35:56: priest <-> F F every 2s at 127.8 vs 124.1),
// but a clearly better one must.
func TestNoPingPongBetweenSimilarEncounters(t *testing.T) {
	LogLevel = 0
	mk := func(t2pos string, t2kills int) map[string]interface{} {
		return state([]pl{
			{"ct1", "ct1", "CT", "1000, 1000, 0", "weapon_m4a1", 0},
			{"t1", "t1", "T", "1900, 1000, 0", "weapon_ak47", 0},
			{"ct2", "ct2", "CT", "5000, 5000, 0", "weapon_m4a1", 0},
			{"t2", "t2", "T", t2pos, "weapon_ak47", t2kills},
		})
	}
	newPA := func() *PlayerAnalyzer {
		pa := NewPlayerAnalyzer()
		pa.Settings = NewDefaultSettings()
		pa.Predict.Enabled = false // tests the legacy (fallback) logic
		pa.SyncCurrentPlayer("t1")
		pa.lastSwitchTime = time.Now().Add(-5 * time.Second)
		pa.currentPlayerSwitchTime = time.Now().Add(-5 * time.Second)
		return pa
	}

	// Other encounter only ~20 points better: stay on t1
	pa := newPA()
	if got := pa.GetBestPlayerToSpectate(mk("5850, 5000, 0", 14), "live"); got != "" {
		t.Fatalf("switched to %q for a marginally better encounter", got)
	}

	// Other encounter much better (players 250 units apart): switch
	pa = newPA()
	if got := pa.GetBestPlayerToSpectate(mk("5250, 5000, 0", 14), "live"); got == "" {
		t.Fatalf("did not switch to a clearly better encounter")
	}
}

// Predictive mode: the camera should go to the player who is close to an
// enemy and aiming at him, not to a player standing far away (e.g. an AWP
// holding an angle at the back while teammates entry).
func TestPredictPicksEntryNotBackAWP(t *testing.T) {
	gs := func() map[string]interface{} {
		p := func(name, team, pos, fwd, weapon string) map[string]interface{} {
			return map[string]interface{}{"name": name, "team": team, "position": pos, "forward": fwd,
				"state":       map[string]interface{}{"health": 100.0, "equip_value": 4700.0},
				"match_stats": map[string]interface{}{"kills": 10.0},
				"weapons":     map[string]interface{}{"w": map[string]interface{}{"name": weapon, "state": "active"}}}
		}
		return map[string]interface{}{"allplayers": map[string]interface{}{
			"awp":   p("awp", "T", "100, 100, 0", "1, 0, 0", "weapon_awp"),     // far back
			"entry": p("entry", "T", "2000, 100, 0", "1, 0, 0", "weapon_ak47"), // 400u from CT, aiming at him
			"ct":    p("ct", "CT", "2400, 100, 0", "-1, 0, 0", "weapon_m4a1"),  // aiming back
			"ct2":   p("ct2", "CT", "2400, 2500, 0", "0, 1, 0", "weapon_m4a1"), // elsewhere
		}}
	}
	pa := NewPlayerAnalyzer()
	pa.Settings = NewDefaultSettings()
	pa.SyncCurrentPlayer("awp")
	pa.lastSwitchTime = time.Now().Add(-10 * time.Second)
	pa.currentPlayerSwitchTime = time.Now().Add(-10 * time.Second)
	got := pa.GetBestPlayerToSpectate(gs(), "live")
	t.Logf("probabilities: %v -> %q", pa.Probabilities(), got)
	if got != "entry" && got != "ct" {
		t.Fatalf("expected the duel (entry or ct), got %q", got)
	}
}
