package director

import (
	"math"
	"sort"
	"strings"
	"time"
)

// FeatureNames lists the per-player features used to predict who is about to
// be in a fight. Everything here comes from data CS2's GSI sends live, so the
// same code runs in the app and in the offline demo replay.
var FeatureNames = []string{
	"near_dist",       // distance to nearest enemy / 1000 (capped at 4)
	"enemies_800",     // enemies within 800 units
	"enemies_1500",    // enemies within 1500 units
	"self_aim_max",    // best alignment of own aim toward an enemy (cos, -1..1)
	"self_aim_on",     // 1 if aiming within ~20° of an enemy (<3000u)
	"aimed_at_max",    // best alignment of an enemy's aim toward this player
	"aimed_at_on",     // 1 if an enemy aims within ~20° of this player
	"closing_speed",   // speed toward nearest enemy / 250 (positive = approaching)
	"own_speed",       // own speed / 250
	"health",          // health / 100
	"is_awp",          // AWP or Scout in hand
	"is_rifle",        // rifle in hand
	"is_pistol_smg",   // pistol, SMG, shotgun, MG
	"no_gun",          // knife, grenade, bomb, nothing
	"equip",           // equipment value / 5000
	"match_kills",     // kills this match / 20
	"recent_kill",     // 1 if killed someone in the last 5s
	"team_alive",      // alive teammates (incl. self) / 5
	"enemy_alive",     // alive enemies / 5
	"near_zdiff",      // |height difference| to nearest enemy / 200
	"is_ct",           // 1 for CT
	"front_rank",      // 0 = closest of his team to the enemy, 1 = furthest
	"near_aim_mutual", // own aim * enemy aim toward each other for nearest enemy (both >0)
}

// FeatureState keeps what's needed between updates (previous positions, kills)
type FeatureState struct {
	prevPos   map[string]Position
	prevTime  time.Time
	lastKills map[string]int
	killTime  map[string]time.Time
}

// NewFeatureState creates an empty feature state
func NewFeatureState() *FeatureState {
	return &FeatureState{
		prevPos:   map[string]Position{},
		lastKills: map[string]int{},
		killTime:  map[string]time.Time{},
	}
}

// Compute returns features for every living player (steamID -> vector).
func (fs *FeatureState) Compute(players []PlayerInfo, now time.Time) map[string][]float64 {
	dt := now.Sub(fs.prevTime).Seconds()
	if fs.prevTime.IsZero() || dt <= 0 || dt > 2 {
		dt = 0
	}
	for _, p := range players {
		if old, ok := fs.lastKills[p.SteamID]; ok && p.Kills > old {
			fs.killTime[p.SteamID] = now
		}
		fs.lastKills[p.SteamID] = p.Kills
	}

	alive := map[string]int{}
	for _, p := range players {
		alive[p.Team]++
	}

	// distance of every player to his nearest enemy (for front rank)
	nearest := map[string]float64{}
	nearestIdx := map[string]int{}
	for i, p := range players {
		best, bi := math.Inf(1), -1
		for j, o := range players {
			if o.Team == p.Team || o.Team == "Unknown" || p.Team == "Unknown" {
				continue
			}
			if d := CalculateDistance3D(p.Position, o.Position); d < best {
				best, bi = d, j
			}
		}
		nearest[p.SteamID] = best
		nearestIdx[p.SteamID] = bi
		_ = i
	}

	out := map[string][]float64{}
	for _, p := range players {
		f := make([]float64, len(FeatureNames))
		nd := nearest[p.SteamID]
		f[0] = math.Min(nd/1000, 4)
		selfAim, aimedAt := -1.0, -1.0
		for _, o := range players {
			if o.Team == p.Team || o.Team == "Unknown" {
				continue
			}
			d := CalculateDistance3D(p.Position, o.Position)
			if d <= 800 {
				f[1]++
			}
			if d <= 1500 {
				f[2]++
			}
			if d > 3000 {
				continue
			}
			if c, ok := aimCos(p, o); ok && c > selfAim {
				selfAim = c
			}
			if c, ok := aimCos(o, p); ok && c > aimedAt {
				aimedAt = c
			}
		}
		f[3] = selfAim
		if selfAim >= aimOnTargetCos {
			f[4] = 1
		}
		f[5] = aimedAt
		if aimedAt >= aimOnTargetCos {
			f[6] = 1
		}

		if prev, ok := fs.prevPos[p.SteamID]; ok && dt > 0 {
			vx, vy := (p.Position.X-prev.X)/dt, (p.Position.Y-prev.Y)/dt
			f[8] = math.Min(math.Hypot(vx, vy)/250, 2)
			if bi := nearestIdx[p.SteamID]; bi >= 0 {
				o := players[bi]
				dx, dy := o.Position.X-p.Position.X, o.Position.Y-p.Position.Y
				if l := math.Hypot(dx, dy); l > 1 {
					f[7] = math.Max(-2, math.Min(2, (vx*dx+vy*dy)/l/250))
				}
			}
		}

		f[9] = float64(p.Health) / 100
		w := p.ActiveWeapon
		switch {
		case isSniperRifle(w):
			f[10] = 1
		case isRifle(w):
			f[11] = 1
		case isShootingWeapon(w):
			f[12] = 1
		default:
			f[13] = 1
		}
		f[14] = float64(p.EquipmentValue) / 5000
		f[15] = float64(p.Kills) / 20
		if t, ok := fs.killTime[p.SteamID]; ok && now.Sub(t) < 5*time.Second {
			f[16] = 1
		}
		f[17] = float64(alive[p.Team]) / 5
		enemyTeam := "CT"
		if p.Team == "CT" {
			enemyTeam = "T"
			f[20] = 1
		}
		f[18] = float64(alive[enemyTeam]) / 5
		if bi := nearestIdx[p.SteamID]; bi >= 0 {
			o := players[bi]
			f[19] = math.Min(math.Abs(p.Position.Z-o.Position.Z)/200, 3)
			c1, ok1 := aimCos(p, o)
			c2, ok2 := aimCos(o, p)
			if ok1 && ok2 && c1 > 0 && c2 > 0 {
				f[22] = c1 * c2
			}
		}

		// front rank among teammates
		var team []float64
		for _, o := range players {
			if o.Team == p.Team {
				team = append(team, nearest[o.SteamID])
			}
		}
		sort.Float64s(team)
		if len(team) > 1 {
			rank := sort.SearchFloat64s(team, nd)
			f[21] = float64(rank) / float64(len(team)-1)
		}

		out[p.SteamID] = f
	}

	fs.prevPos = map[string]Position{}
	for _, p := range players {
		fs.prevPos[p.SteamID] = p.Position
	}
	fs.prevTime = now
	return out
}

func isRifle(w string) bool {
	for _, r := range []string{"ak47", "m4a1", "aug", "sg556", "famas", "galilar"} {
		if strings.Contains(w, r) {
			return true
		}
	}
	return false
}
