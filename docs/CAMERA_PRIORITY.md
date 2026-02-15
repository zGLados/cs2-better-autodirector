# 🎥 Camera Prioritization System

This document explains how the **Better Auto Observer** intelligently selects which player to spectate based on a sophisticated priority system.

---

## 🎯 Core Concept

The system analyzes all living players and their positions to detect **encounters** (situations where players from opposing teams are close to each other). Each encounter receives a **priority score**, and the camera switches to the player in the highest-priority encounter.

---

## 📊 Priority Calculation

### 1. **Distance-Based Priority** (Most Important!)

Distance is the **primary factor** because it indicates how likely a fight is to happen:

| Distance Range | Priority | Description |
|---------------|----------|-------------|
| **< 300 units** | +150 | 🔴 **Imminent fight!** Close combat, grenades, shotguns |
| **300-600 units** | +120 | 🟠 **High action** Close-range rifles, SMGs |
| **600-1000 units** | +80 | 🟡 **Medium-close** Standard rifle combat |
| **1000-1500 units** | +50 | 🟢 **Medium range** Longer engagements |
| **1500-2000 units** | +20 | 🔵 **Long range** AWP, distant sniping |

> **Note:** Encounters beyond 2000 units are **ignored** as they're too far to result in immediate action.

---

### 2. **Equipment Value**

Players with better weapons create more exciting moments:

```
Priority += Average Equipment Value / 200
```

**Examples:**
- Pistol round ($800 avg) → +4 priority
- Full buy ($9000 avg) → +45 priority
- AWP + full utility ($9000+ avg) → +45+ priority

---

### 3. **Skill Level (Kills)**

Players with more kills tend to create better action:

```
Priority += Average Kills × 3.0
```

**Examples:**
- 0 kills → +0 priority
- 5 kills avg → +15 priority  
- 10 kills avg → +30 priority
- 20 kills avg → +60 priority

---

### 4. **Health Status**

Low HP adds excitement but is weighted carefully (players might die before action):

```
If either player < 50 HP  → +10 priority
If either player < 30 HP  → +10 priority (additional)
```

**Total bonus:** Up to +20 for very low HP encounters

---

### 5. **Defuser Bonus**

CT with defuse kit in encounter:

```
Priority += 20
```

---

## ⭐ Current Player Priority Bonus (Smart Sticky Camera)

**The most important feature:** The system prefers to **stay with the current player** if they're in meaningful action, but switches away if stuck too long or better action is happening.

### How it works:

1. All encounters are calculated normally
2. If the **currently spectated player** is in an encounter, check conditions:
   - ✅ **Base priority must be >80** (close enough for action)
   - ✅ **Time on current player <15 seconds** (prevent getting stuck)
   
   If both conditions are met:
   ```
   That encounter's priority += 30
   ```
3. After re-sorting, if the current player is still in the top encounter, **stay with them**

### Why these limits matter:

✅ **Camera stability** - Don't jump during active fights  
✅ **Prevents infinite sticking** - Max 15s on one player forces variety  
✅ **Action-focused** - Only bonus for close encounters (<1000 units typically)  
✅ **Balanced** - +30 bonus is enough to prefer current player, but not override much better action  

### Example Scenarios:

**Scenario A: Bonus Applied (Good Action)**
```
Encounter: Player3 vs Player4 (currently spectating Player3)
  Distance: 600 units  
  Base Priority: 120 (distance) + 25 (equipment) = 145
  Time on Player3: 8 seconds
  ✅ Priority >80 AND <15s → +30 Bonus Applied
  Total Priority: 175 ✅ Stay with Player3
```

**Scenario B: No Bonus (Too Long) - BUT Active Encounter Exception**
```
Encounter: Player3 vs Player5
  Base Priority: 125 (close fight!)
  Time on Player3: 18 seconds
  ❌ Exceeded 15s limit BUT...
  ✅ Active encounter (Priority >80) → Exception applies!
  Total Priority: 125 + 30 = 155 ✅ Stay with Player3 (fight in progress)
```

**Scenario C: No Bonus (Too Long, Low Priority)**
```
Encounter: Player3 vs Player6
  Base Priority: 50 (far away)
  Time on Player3: 18 seconds
  ❌ Exceeded 15s limit AND Priority <80
  Total Priority: 50 → Switch to better encounter
```

**Scenario D: Clutch Situation Exception**
```
Encounter: Player3 vs Player7
  Distance: 1200 units
  Base Priority: 70
  Time on Player3: 25 seconds
  Players alive: 3 (CT: 2, T: 1)
  ✅ Clutch situation (<4 players) → No time limit!
  Total Priority: 70 → Stay with Player3 (limited action available)
```

**Scenario E: No Bonus (Too Far)**
```
Encounter: Player3 vs Player6
  Distance: 1800 units  
  Base Priority: 35 (low because far away)
  ❌ Priority <80 → NO bonus
  Total Priority: 35 → Switch to closer action
```

---

## 🎮 Player Selection Within Encounter

Once the best encounter is identified, the system chooses which player to spectate:

### Priority Order:

1. **If currently spectating one player in this encounter:**
   ```
   → Stay with that player (camera stability)
   ```

2. **Otherwise, calculate player scores:**
   ```
   Score = (Equipment Value / 100) + (Kills × 10) + Health Bonus
   
   Health Bonus:
   - HP > 50 → +5 score
   - HP ≤ 50 → +0 score
   ```

3. **Choose player with higher score**

### Examples:

**Scenario 1: Equal equipment**
- Player A: $5000 equipment, 10 kills, 100 HP
  - Score: 50 + 100 + 5 = **155**
- Player B: $5000 equipment, 3 kills, 80 HP
  - Score: 50 + 30 + 5 = **85**
- **Result:** Spectate Player A (better fragger)

**Scenario 2: Better weapon vs better fragger**
- Player A: $9000 equipment (AWP), 2 kills, 100 HP
  - Score: 90 + 20 + 5 = **115**
- Player B: $4000 equipment (M4), 12 kills, 100 HP
  - Score: 40 + 120 + 5 = **165**
- **Result:** Spectate Player B (hot fragger)

---

## 🛡️ Safety Mechanisms

### 1. **Dead Player Prevention**

Before every switch, verify the target player is still alive:

```go
if target player is DEAD {
    Skip this switch
    Log warning
}
```

### 2. **Dead Player Detection**

Currently spectated player dies:

```go
if currently spectated player not in alive players {
    Bypass rate limiting
    Force immediate switch
    Log: "☠️ Currently spectated player DIED - forcing immediate switch!"
}
```

### 3. **Rate Limiting**

Prevent camera from switching too frequently:

```
Minimum time between switches: 2 seconds (except on player death)
```

**Exception:** When spectated player dies, switch **immediately** (0s delay)

### 4. **Sticky Time Limit**

Prevent camera from staying on one player indefinitely:

```
Maximum time on one player: 15 seconds
```

After 15 seconds, the current player bonus is **removed**, forcing the system to find better action.

**EXCEPTIONS - Sticky time limit is DISABLED in these situations:**

1. **Clutch Situation (< 4 players alive)**
   - When fewer than 4 players are alive, camera can stay on one player indefinitely
   - Reason: Limited action available, want to follow the remaining players
   - Log: `🎯 Clutch situation (X players alive) - sticky time limit disabled`

2. **Active Encounter (Priority >80)**
   - When current player is in an active, close-range encounter
   - Reason: Don't interrupt during an ongoing fight
   - Log: `🔥 Active encounter detected - sticky time limit disabled for this fight`

These exceptions ensure the camera stays stable during critical moments:

### 5. **Minimum Priority Threshold for Bonus**

Current player bonus only applied if encounter is worth watching:

```
Base Priority must be > 80
```

This typically means:
- Distance < 1000 units (medium-close range)
- OR very high equipment/skill values

**Why:** No point staying with player in boring, far-away encounter

---

## 🔄 Complete Decision Flow

```mermaid
graph TD
    A[Game State Update] --> B{Currently spectated player alive?}
    B -->|No| C[FORCE IMMEDIATE SWITCH]
    B -->|Yes| D{Rate limit satisfied?}
    D -->|No| E[Wait - No switch]
    D -->|Yes| F[Get all alive players]
    F --> G[Calculate all encounters < 2000 units]
    G --> H{Current player in any encounter?}
    H -->|Yes| I1{Less than 4 players alive?}
    I1 -->|Yes| EX1[🎯 Exception: Clutch - No time limit]
    I1 -->|No| I2{Time on player < 15s?}
    I2 -->|No| I3{Base Priority > 80?}
    I3 -->|Yes| EX2[🔥 Exception: Active encounter - No time limit]
    I3 -->|No| J1[❌ Sticky time exceeded - NO bonus]
    I2 -->|Yes| K{Base Priority > 80?}
    K -->|No| J2[❌ Priority too low - NO bonus]
    K -->|Yes| L[✅ Add +30 bonus]
    EX1 --> L
    EX2 --> L
    H -->|No| J[Continue]
    J1 --> J
    J2 --> J
    L --> J
    J --> M[Sort by priority]
    M --> N[Select top encounter]
    N --> O{Current player in this encounter?}
    O -->|Yes| P[Stay with current player]
    O -->|No| Q[Calculate player scores]
    Q --> R[Select player with higher score]
    R --> S{Target player alive?}
    S -->|No| T[Skip switch]
    S -->|Yes| U[Execute switch]
    P --> U
    C --> U
    U --> V[Update last switch time & sticky timer]
    V --> A
```

---

## 📈 Example Priority Scenarios

### Scenario 1: Close Combat vs Distant Encounter

**Encounter A:**
- Distance: 250 units (close!)
- Equipment: $4000 avg
- Kills: 5 avg
- HP: Both healthy

```
Priority = 150 (distance) + 20 (equipment) + 15 (kills) = 185
```

**Encounter B:**
- Distance: 1800 units (far)
- Equipment: $9000 avg (AWP duel!)
- Kills: 15 avg  
- HP: Both healthy

```
Priority = 20 (distance) + 45 (equipment) + 45 (kills) = 110
```

**Result:** Spectate **Encounter A** (close combat is more likely to result in kills)

---

### Scenario 2: Current Player Bonus in Action

**Encounter A:**
- Distance: 400 units
- Priority: 165

**Encounter B (contains current player):**
- Distance: 900 units
- Base Priority: 95
- + Current Bonus: +100
- **Total: 195** ✅

**Result:** Stay with current player (camera stability)

---

### Scenario 3: Player Death Override

```
Currently spectating: PlayerX
PlayerX: Just died

Action:
1. Detect PlayerX not in alive players list
2. Log: "☠️ Currently spectated player DIED"
3. BYPASS rate limiting
4. Find best encounter immediately
5. Switch in <100ms
```

---

## 🎛️ Configuration Parameters

Current values in the code:

| Parameter | Value | Purpose |
|-----------|-------|---------|
| `minSwitchInterval` | 2 seconds | Minimum time between normal switches |
| `maxEncounterDistance` | 2000 units | Maximum distance for encounter detection |
| `currentPlayerBonus` | +100 | Priority bonus for current player's encounters |
| `closeRangePriority` | +150 | Bonus for < 300 unit encounters |
| `mediumRangePriority` | +80 | Bonus for 600-1000 unit encounters |

---

## 🔍 Debugging with Verbose Mode

Run with `-v` flag to see detailed priority calculations:

```cmd
better-autoobserver.exe -v
```

**Example output:**

```
[ANALYZER] Found 8 alive players
[ANALYZER] Detected 12 potential encounters
[ANALYZER] ⭐ Current player in encounter: +100 priority (135.5 → 235.5)
[INFO] ⚔️  ENCOUNTER: ShadowSC (T) vs m1337zk3 (CT) | Distance: 450 units | Priority: 235.5
[ANALYZER] ✓ Staying with current player ShadowSC in this encounter
```

---

## 💡 Tips for Best Results

1. **Use in GOTV/Demo mode** - Position data only available there
2. **Keep CS in foreground** - Keyboard simulation requires focus
3. **Run as Administrator** - Improves keyboard input reliability
4. **Watch the logs** - Use `-v` to understand camera decisions

---

## 🚀 Future Improvements

Potential enhancements to the priority system:

- [ ] Detect bomb plant/defuse situations (+200 priority)
- [ ] Detect flashbang/smoke usage (temporary +50 priority)
- [ ] Track recent damage dealt (wounded players more likely to die)
- [ ] Machine learning to predict kill probability from player movements
- [ ] Custom priority profiles (aggressive vs conservative switching)

---

## 📝 Technical Notes

### Position Data Format

CS2 sends position as **string format:**

```json
"position": "-1520.06, 430.89, -63.97"
```

The system parses this into X, Y, Z coordinates for distance calculation using 2D Euclidean distance (ignoring Z for vertical differences).

### Distance Calculation

```go
distance = sqrt((x2-x1)² + (y2-y1)²)
```

Z-coordinate (height) is intentionally ignored to treat multi-level encounters at same priority.

---

## 📚 Related Documentation

- [Main README](../README.md) - Project overview and setup
- [Build Instructions](BUILD.md) - How to compile the project
- [GSI Configuration](../config/gamestate_integration_autoobserver.cfg) - Game State Integration settings

---

<div align="center">

**Made with ❤️ for the CS2 community**

[Report Issue](../../issues) • [Request Feature](../../issues)

</div>
