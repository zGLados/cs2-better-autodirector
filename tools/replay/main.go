// Command replay runs the autodirector's camera logic offline on CS2 demos and
// measures how often the camera was on the right player.
//
// Usage (from this folder):
//
//	go run . path\to\demo1.dem path\to\demo2.dem ...
//
// It prints a summary per demo and writes missed_kills.csv with every kill the
// camera missed (tick, players, what the camera was doing) so you can check it
// in CS2 with demo_gototick.
package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cs2-autodirector-gui/director"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

var (
	aimCheck   bool
	aimCosines []float64
	datasetOut string
	modelDir   string
	predFlag   string
	legacy     bool
)

// featureRow is one player at one moment, for training the prediction model
type featureRow struct {
	demo  string
	round int
	t     time.Duration
	id    string
	f     []float64
}

// Timing model of the live app
var (
	gsiInterval    = 100 * time.Millisecond // GSI update rate (throttle 0.1)
	gsiLag         = 100 * time.Millisecond // GSI buffer: director sees slightly old data
	directorTick   = 500 * time.Millisecond // AutoDirector loop interval
	switchDelay    = 250 * time.Millisecond // key press + GSI confirmation
	stuckRoundSecs = 180.0                  // a "live" round longer than this is a demo bug
)

type killRec struct {
	Demo        string
	Round       int
	Tick        int
	T           time.Duration
	Killer      string
	KillerID    string
	Victim      string
	VictimID    string
	Weapon      string
	Wallbang    bool
	CamAtKill   string
	CamOnKiller map[int]bool // seconds before kill (0..3) -> camera on killer
	CamOnVictim bool
	LastInfo    string
}

type camSample struct {
	t  time.Duration
	id string
}

type demoResult struct {
	Name         string
	Rows         []featureRow
	BrokenRounds map[int]bool
	Kills        []killRec
	Switches     int
	LiveSeconds  float64
	SkippedRound int
	Err          error
}

func main() {
	csvPath := flag.String("csv", "missed_kills.csv", "where to write the missed kills")
	verbose := flag.Bool("v", false, "print the director's decisions")
	flag.StringVar(&modelDir, "modeldir", "", "folder with <demo>.json models (for testing on demos the model never saw)")
	flag.StringVar(&predFlag, "pred", "", "predictive settings, e.g. dwell=3,margin=0.1,fastp=0.35,fastdwell=1.5,smooth=0.5,events=0")
	flag.BoolVar(&legacy, "legacy", false, "use the old encounter-based logic (no prediction)")
	flag.StringVar(&datasetOut, "dataset", "", "write per-player features + labels (who is in a kill in the next seconds) to this csv")
	flag.BoolVar(&aimCheck, "aimcheck", false, "sanity-check view angles: killer's aim vs victim at each kill")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("usage: replay [-csv file] demo1.dem [demo2.dem ...]")
		os.Exit(2)
	}

	var results []demoResult
	for _, path := range flag.Args() {
		fmt.Printf("▶ %s ...\n", filepath.Base(path))
		start := time.Now()
		r := replayDemo(path, *verbose)
		fmt.Printf("  parsed in %s\n", time.Since(start).Round(time.Second))
		if r.Err != nil {
			fmt.Printf("  ⚠️  %v (partial results kept)\n", r.Err)
		}
		printSummary(r)
		results = append(results, r)
	}

	if aimCheck && len(aimCosines) > 0 {
		sort.Float64s(aimCosines)
		on := 0
		for _, c := range aimCosines {
			if c >= 0.94 {
				on++
			}
		}
		fmt.Printf("aim check: median cos %.3f, %d/%d kills with killer aiming within 20°\n",
			aimCosines[len(aimCosines)/2], on, len(aimCosines))
	}
	if len(results) > 1 {
		var all demoResult
		all.Name = "TOTAL"
		for _, r := range results {
			all.Kills = append(all.Kills, r.Kills...)
			all.Switches += r.Switches
			all.LiveSeconds += r.LiveSeconds
			all.SkippedRound += r.SkippedRound
		}
		printSummary(all)
	}
	if datasetOut != "" {
		if err := writeDataset(datasetOut, results); err != nil {
			fmt.Println("could not write dataset:", err)
		} else {
			fmt.Println("dataset written to", datasetOut)
		}
	}
	if err := writeCSV(*csvPath, results); err != nil {
		fmt.Println("could not write csv:", err)
	} else {
		fmt.Println("missed kills written to", *csvPath)
	}
}

func replayDemo(path string, verbose bool) (res demoResult) {
	res.Name = filepath.Base(path)
	f, err := os.Open(path)
	if err != nil {
		res.Err = err
		return
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	// Director with the demo's clock
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	director.Now = func() time.Time { return base.Add(p.CurrentTime()) }
	lastInfo := ""
	director.LogInfo = func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		lastInfo = msg
		if verbose {
			fmt.Printf("  [%s] %s\n", fmtDur(p.CurrentTime()), msg)
		}
	}
	director.LogVerbose = func(string, ...interface{}) {}
	pa := director.NewPlayerAnalyzer()
	if legacy {
		pa.Predict.Enabled = false
	}
	if modelDir != "" {
		if m, err := director.LoadModel(filepath.Join(modelDir, res.Name+".json")); err == nil {
			pa.Model = m
		} else {
			fmt.Println("  (no held-out model for this demo, using built-in)")
		}
	}
	applyPredFlag(&pa.Predict, predFlag)
	fs := director.NewFeatureState()
	res.BrokenRounds = map[int]bool{}

	var (
		camera      string // who is on screen
		pendingID   string // switch in flight
		pendingAt   time.Duration
		lastSwitch  time.Duration
		snapshots   []gsiSnapshot
		nextGSI     time.Duration
		nextTick    time.Duration
		history     []camSample
		roundNum    int
		roundOver   bool
		liveStart   time.Duration
		roundKills  []killRec
		roundBroken bool
	)
	camAt := func(t time.Duration) string {
		i := sort.Search(len(history), func(i int) bool { return history[i].t > t })
		if i == 0 {
			return ""
		}
		return history[i-1].id
	}

	phase := func() string {
		gs := p.GameState()
		switch {
		case gs.IsWarmupPeriod():
			return "warmup"
		case gs.IsFreezetimePeriod():
			return "freezetime"
		case roundOver:
			return "over"
		default:
			return "live"
		}
	}

	finishRound := func() {
		if roundBroken {
			res.SkippedRound++
			res.BrokenRounds[roundNum] = true
		} else {
			res.Kills = append(res.Kills, roundKills...)
		}
		roundKills = nil
		roundBroken = false
	}

	p.RegisterEventHandler(func(e events.RoundStart) {
		finishRound()
		roundNum++
		roundOver = false
		liveStart = p.CurrentTime()
	})
	p.RegisterEventHandler(func(e events.RoundEnd) { roundOver = true })
	p.RegisterEventHandler(func(e events.Kill) {
		if phase() != "live" || e.Killer == nil || e.Victim == nil || e.Killer == e.Victim {
			return
		}
		if e.Killer.Team == e.Victim.Team {
			return
		}
		now := p.CurrentTime()
		k := killRec{
			Demo: res.Name, Round: roundNum, Tick: p.GameState().IngameTick(), T: now,
			Killer: e.Killer.Name, KillerID: strconv.FormatUint(e.Killer.SteamID64, 10),
			Victim: e.Victim.Name, VictimID: strconv.FormatUint(e.Victim.SteamID64, 10),
			Wallbang: e.PenetratedObjects > 0, CamOnKiller: map[int]bool{}, LastInfo: lastInfo,
		}
		if e.Weapon != nil {
			k.Weapon = e.Weapon.String()
		}
		if aimCheck {
			kp, vp := e.Killer.Position(), e.Victim.Position()
			yaw := float64(e.Killer.ViewDirectionX()) * math.Pi / 180
			pd := float64(e.Killer.ViewDirectionY())
			if pd > 180 {
				pd -= 360
			}
			pt := pd * math.Pi / 180
			fx, fy, fz := math.Cos(pt)*math.Cos(yaw), math.Cos(pt)*math.Sin(yaw), -math.Sin(pt)
			dx, dy, dz := vp.X-kp.X, vp.Y-kp.Y, (vp.Z+40)-(kp.Z+64)
			d := math.Sqrt(dx*dx + dy*dy + dz*dz)
			if d > 1 {
				aimCosines = append(aimCosines, (fx*dx+fy*dy+fz*dz)/d)
			}
		}
		k.CamAtKill = camera
		k.CamOnVictim = camera == k.VictimID
		for s := 0; s <= 3; s++ {
			k.CamOnKiller[s] = camAt(now-time.Duration(s)*time.Second) == k.KillerID
		}
		roundKills = append(roundKills, k)
	})

	defer func() {
		if r := recover(); r != nil {
			res.Err = fmt.Errorf("parser panic: %v", r)
		}
		finishRound()
	}()

	for {
		more, err := p.ParseNextFrame()
		now := p.CurrentTime()

		// Apply a switch that is "arriving" on screen
		if pendingID != "" && now >= pendingAt {
			camera = pendingID
			pendingID = ""
			history = append(history, camSample{now, camera})
		}

		if now >= nextGSI {
			nextGSI = now + gsiInterval
			snapshots = append(snapshots, gsiSnapshot{now, buildGSI(p, phase(), camera)})
			if len(snapshots) > 20 {
				snapshots = snapshots[len(snapshots)-20:]
			}
		}

		if now >= nextTick && len(snapshots) > 0 {
			nextTick = now + directorTick
			ph := phase()
			if ph == "live" {
				res.LiveSeconds += directorTick.Seconds()
				if (now-liveStart).Seconds() > stuckRoundSecs && !roundBroken {
					// Demo didn't advance the round (known bug in some demos):
					// ignore this round's numbers instead of polluting them.
					roundBroken = true
				}
			}
			// Director sees GSI data that is gsiLag old
			state := snapshots[len(snapshots)-1].state
			for i := len(snapshots) - 1; i >= 0; i-- {
				if now-snapshots[i].t >= gsiLag {
					state = snapshots[i].state
					break
				}
			}
			if ph == "live" && datasetOut != "" {
				feats := fs.Compute(director.GetPlayerInfo(state), director.Now())
				for id, fv := range feats {
					res.Rows = append(res.Rows, featureRow{res.Name, roundNum, now, id, fv})
				}
			}
			if ph == "live" || ph == "freezetime" || ph == "warmup" || ph == "timeout" {
				// Same sync rule as the app
				if camera != "" && camera != pa.CurrentSpectatedID() && !pa.RecentlySwitched(2*time.Second) {
					pa.SyncCurrentPlayer(camera)
				}
				if target := pa.GetBestPlayerToSpectate(state, ph); target != "" && target != camera {
					pendingID = target
					pendingAt = now + switchDelay
					if ph == "live" {
						res.Switches++
					}
					lastSwitch = now
				}
			}
		}

		if err != nil {
			if !errors.Is(err, dem.ErrUnexpectedEndOfDemo) {
				res.Err = err
			}
			break
		}
		if !more {
			break
		}
	}
	_ = lastSwitch
	return
}

type gsiSnapshot struct {
	t     time.Duration
	state map[string]interface{}
}

// buildGSI produces the same JSON structure CS2's GSI sends to the app.
func buildGSI(p dem.Parser, phase, camera string) map[string]interface{} {
	all := map[string]interface{}{}
	for _, pl := range p.GameState().Participants().Playing() {
		if pl == nil || pl.SteamID64 == 0 {
			continue
		}
		team := "Unknown"
		switch pl.Team {
		case common.TeamCounterTerrorists:
			team = "CT"
		case common.TeamTerrorists:
			team = "T"
		}
		health := 0
		if pl.IsAlive() {
			health = pl.Health()
		}
		pos := pl.Position()
		yaw := float64(pl.ViewDirectionX()) * math.Pi / 180
		pitchDeg := float64(pl.ViewDirectionY())
		if pitchDeg > 180 {
			pitchDeg -= 360
		}
		pitch := pitchDeg * math.Pi / 180
		fwd := [3]float64{math.Cos(pitch) * math.Cos(yaw), math.Cos(pitch) * math.Sin(yaw), -math.Sin(pitch)}

		weapons := map[string]interface{}{}
		if w := pl.ActiveWeapon(); w != nil {
			weapons["weapon_0"] = map[string]interface{}{"name": gsiWeaponName(w.Type), "state": "active"}
		}
		id := strconv.FormatUint(pl.SteamID64, 10)
		all[id] = map[string]interface{}{
			"name":     pl.Name,
			"team":     team,
			"position": fmt.Sprintf("%.2f, %.2f, %.2f", pos.X, pos.Y, pos.Z),
			"forward":  fmt.Sprintf("%.4f, %.4f, %.4f", fwd[0], fwd[1], fwd[2]),
			"state": map[string]interface{}{
				"health":      float64(health),
				"armor":       float64(pl.Armor()),
				"money":       float64(pl.Money()),
				"equip_value": float64(pl.EquipmentValueCurrent()),
				"defusekit":   pl.HasDefuseKit(),
			},
			"match_stats": map[string]interface{}{"kills": float64(pl.Kills())},
			"weapons":     weapons,
		}
	}
	gs := map[string]interface{}{
		"allplayers": all,
		"round":      map[string]interface{}{"phase": phase},
	}
	if camera != "" {
		gs["player"] = map[string]interface{}{"steamid": camera}
	}
	return gs
}

func gsiWeaponName(t common.EquipmentType) string {
	switch t {
	case common.EqAWP:
		return "weapon_awp"
	case common.EqSSG08:
		return "weapon_ssg08"
	case common.EqAK47:
		return "weapon_ak47"
	case common.EqM4A4:
		return "weapon_m4a1"
	case common.EqM4A1:
		return "weapon_m4a1_silencer"
	case common.EqAUG:
		return "weapon_aug"
	case common.EqSG553:
		return "weapon_sg556"
	case common.EqDeagle:
		return "weapon_deagle"
	case common.EqKnife:
		return "weapon_knife"
	case common.EqBomb:
		return "weapon_c4"
	case common.EqZeus:
		return "weapon_taser"
	case common.EqHE:
		return "weapon_hegrenade"
	case common.EqFlash:
		return "weapon_flashbang"
	case common.EqSmoke:
		return "weapon_smokegrenade"
	case common.EqMolotov:
		return "weapon_molotov"
	case common.EqIncendiary:
		return "weapon_incgrenade"
	case common.EqDecoy:
		return "weapon_decoy"
	}
	name := strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(t.String()))
	if name == "" || name == "unknown" {
		return ""
	}
	return "weapon_" + name
}

func pct(n, d int) string {
	if d == 0 {
		return "  -  "
	}
	return fmt.Sprintf("%4.0f%%", 100*float64(n)/float64(d))
}

func printSummary(r demoResult) {
	n := len(r.Kills)
	var at0, at1, at2, at3, victim, seen int
	for _, k := range r.Kills {
		if k.CamOnKiller[0] {
			at0++
		}
		if k.CamOnKiller[1] {
			at1++
		}
		if k.CamOnKiller[2] {
			at2++
		}
		if k.CamOnKiller[3] {
			at3++
		}
		if k.CamOnVictim {
			victim++
		}
		if k.CamOnKiller[0] || k.CamOnVictim {
			seen++
		}
	}
	perMin := 0.0
	if r.LiveSeconds > 0 {
		perMin = float64(r.Switches) / (r.LiveSeconds / 60)
	}
	fmt.Printf("  ── %s: %d kills", r.Name, n)
	if r.SkippedRound > 0 {
		fmt.Printf(" (%d bugged rounds ignored)", r.SkippedRound)
	}
	fmt.Println()
	fmt.Printf("     camera on killer at the kill ........ %s\n", pct(at0, n))
	fmt.Printf("     camera on killer 1s / 2s / 3s before  %s / %s / %s\n", pct(at1, n), pct(at2, n), pct(at3, n))
	fmt.Printf("     camera on victim at the kill ........ %s\n", pct(victim, n))
	fmt.Printf("     kill on screen (killer or victim) ... %s\n", pct(seen, n))
	fmt.Printf("     switches per live minute ............ %.1f\n", perMin)
}

func writeCSV(path string, results []demoResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"demo", "round", "tick", "time", "killer", "victim", "weapon", "wallbang", "camera_on", "on_killer_1s_before", "last_director_message"})
	for _, r := range results {
		names := map[string]string{}
		for _, k := range r.Kills {
			names[k.KillerID] = k.Killer
			names[k.VictimID] = k.Victim
		}
		for _, k := range r.Kills {
			if k.CamOnKiller[0] {
				continue
			}
			cam := names[k.CamAtKill]
			if cam == "" {
				cam = k.CamAtKill
			}
			if k.CamOnVictim {
				cam += " (victim)"
			}
			w.Write([]string{k.Demo, strconv.Itoa(k.Round), strconv.Itoa(k.Tick), fmtDur(k.T), k.Killer, k.Victim, k.Weapon,
				strconv.FormatBool(k.Wallbang), cam, strconv.FormatBool(k.CamOnKiller[1]), k.LastInfo})
		}
	}
	return nil
}

func fmtDur(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

// writeDataset labels each row with whether that player was the killer / was
// involved (killer or victim) in a kill within the next 1..3 seconds.
func writeDataset(path string, results []demoResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	head := []string{"demo", "round", "t", "id"}
	head = append(head, director.FeatureNames...)
	head = append(head, "kill1", "kill2", "kill3", "inv1", "inv2", "inv3")
	w.Write(head)
	b := func(x bool) string {
		if x {
			return "1"
		}
		return "0"
	}
	for _, r := range results {
		type ev struct {
			t      time.Duration
			killer bool
		}
		byID := map[string][]ev{}
		for _, k := range r.Kills {
			byID[k.KillerID] = append(byID[k.KillerID], ev{k.T, true})
			byID[k.VictimID] = append(byID[k.VictimID], ev{k.T, false})
		}
		for _, row := range r.Rows {
			if r.BrokenRounds[row.round] {
				continue
			}
			var kill, inv [4]bool
			for _, e := range byID[row.id] {
				dt := e.t - row.t
				for s := 1; s <= 3; s++ {
					if dt > 0 && dt <= time.Duration(s)*time.Second {
						inv[s] = true
						if e.killer {
							kill[s] = true
						}
					}
				}
			}
			rec := []string{row.demo, strconv.Itoa(row.round), fmt.Sprintf("%.2f", row.t.Seconds()), row.id}
			for _, v := range row.f {
				rec = append(rec, strconv.FormatFloat(v, 'f', 4, 64))
			}
			rec = append(rec, b(kill[1]), b(kill[2]), b(kill[3]), b(inv[1]), b(inv[2]), b(inv[3]))
			w.Write(rec)
		}
	}
	return nil
}

func applyPredFlag(c *director.PredictConfig, spec string) {
	for _, kv := range strings.Split(spec, ",") {
		parts := strings.SplitN(strings.TrimSpace(kv), "=", 2)
		if len(parts) != 2 {
			continue
		}
		v, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		sec := time.Duration(v * float64(time.Second))
		switch parts[0] {
		case "dwell":
			c.MinDwell = sec
		case "margin":
			c.Margin = v
		case "fastp":
			c.FastP = v
		case "fastdwell":
			c.FastDwell = sec
		case "smooth":
			c.Smooth = v
		case "events":
			c.EventSwitches = v != 0
		}
	}
}
