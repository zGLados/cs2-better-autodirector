import './style.css';
import './app.css';
import './dashboard.css';

import { StartAutoDirector, StopAutoDirector, GetStatus, GetPlayers, GetEncounters, GetStatistics, GetSettings, SaveSettings, ResetSettings, ExportSettings, ImportSettings } from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

// Global state
let isRunning = false;
let logs = [];
const MAX_LOGS = 100;
let currentPage = 'dashboard'; // Track current page

// Initialize dashboard
document.addEventListener('DOMContentLoaded', () => {
    showPage('dashboard');
    setupWailsEvents();
});

// Page navigation
function showPage(page) {
    currentPage = page;
    if (page === 'dashboard') {
        initDashboard();
        setupEventListeners();
        startDataPolling();
    } else if (page === 'settings') {
        initSettings();
        setupSettingsListeners();
    }
}

function initDashboard() {
    document.querySelector('#app').innerHTML = `
        <div class="dashboard">
            <!-- Header -->
            <header class="dashboard-header">
                <div class="header-content">
                    <h1>🎮 CS2 Auto Director</h1>
                    <div class="header-controls">
                        <button id="settingsBtn" class="btn btn-secondary" style="white-space: nowrap; flex-shrink: 0;">⚙️ Settings</button>
                        <button id="startBtn" class="btn btn-success" style="white-space: nowrap; flex-shrink: 0;">▶ Start</button>
                        <button id="stopBtn" class="btn btn-danger" style="white-space: nowrap; flex-shrink: 0;" disabled>⏹ STOP</button>
                    </div>
                </div>
            </header>

            <!-- Main Content -->
            <div class="dashboard-content">
                <!-- Left Column -->
                <div class="dashboard-column">
                    <!-- Status Widget -->
                    <div class="widget status-widget">
                        <h2>📊 Status</h2>
                        <div class="status-grid">
                            <div class="status-item">
                                <span class="status-label">State:</span>
                                <span id="runningStatus" class="status-value">Stopped</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Current Player:</span>
                                <span id="currentPlayer" class="status-value">-</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Round:</span>
                                <span id="roundNumber" class="status-value">-</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Score:</span>
                                <span id="score" class="status-value">- : -</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Phase:</span>
                                <span id="roundPhase" class="status-value">-</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Total Switches:</span>
                                <span id="totalSwitches" class="status-value">0</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Switches/Min:</span>
                                <span id="switchesPerMin" class="status-value">0.0</span>
                            </div>
                            <div class="status-item">
                                <span class="status-label">Alive:</span>
                                <span id="alivePlayers" class="status-value">0</span>
                            </div>
                        </div>
                    </div>

                    <!-- Statistics Widget -->
                    <div class="widget stats-widget">
                        <h2>📈 Statistics</h2>
                        <div class="stats-grid">
                            <div class="stat-item">
                                <div class="stat-label">Sniper Kills</div>
                                <div id="sniperKills" class="stat-value">0</div>
                            </div>
                            <div class="stat-item">
                                <div class="stat-label">Upset Victories</div>
                                <div id="upsetVictories" class="stat-value">0</div>
                            </div>
                            <div class="stat-item">
                                <div class="stat-label">Damage Detections</div>
                                <div id="damageDetections" class="stat-value">0</div>
                            </div>
                        </div>
                    </div>

                    <!-- Encounters Widget -->
                    <div class="widget encounters-widget">
                        <h2>⚔️ Top Encounters</h2>
                        <div id="encountersList" class="encounters-list">
                            <div class="no-data">No encounters detected</div>
                        </div>
                    </div>
                </div>

                <!-- Right Column -->
                <div class="dashboard-column">
                    <!-- Players Widget -->
                    <div class="widget players-widget">
                        <h2>🎯 Scoreboard</h2>
                        <div class="scoreboard-container">
                            <div id="scoreboard" class="scoreboard">
                                <div class="no-data">Waiting for game data...</div>
                            </div>
                        </div>
                    </div>

                    <!-- Logs Widget -->
                    <div class="widget logs-widget">
                        <h2>📝 Live Logs</h2>
                        <div id="logsContainer" class="logs-container">
                            <div class="log-entry">Waiting for events...</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

function setupEventListeners() {
    const settingsBtn = document.getElementById('settingsBtn');
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');

    if (settingsBtn) {
        settingsBtn.addEventListener('click', (e) => {
            e.preventDefault();
            console.log('Settings button clicked!');
            showPage('settings');
        });
        console.log('Settings button found and event listener attached');
    } else {
        console.error('Settings button not found!');
    }

    if (startBtn) {
        startBtn.addEventListener('click', async () => {
            try {
                await StartAutoDirector();
                isRunning = true;
                updateControlButtons();
                addLog('✅ AutoDirector started');
            } catch (err) {
                addLog(`❌ Error starting: ${err}`);
                console.error(err);
            }
        });
    }

    if (stopBtn) {
        stopBtn.addEventListener('click', async () => {
            try {
                await StopAutoDirector();
                isRunning = false;
                updateControlButtons();
                addLog('⏹ AutoDirector stopped');
            } catch (err) {
                addLog(`❌ Error stopping: ${err}`);            
                console.error(err);
            }
        });
    }
}

function setupWailsEvents() {
    // Listen for status changes
    EventsOn('status_changed', (status) => {
        isRunning = (status === 'running');
        updateControlButtons();
        addLog(`🔄 Status changed: ${status}`);
    });

    // Listen for player updates
    EventsOn('players_updated', (players) => {
        updatePlayersTable(players);
    });

    // Listen for encounter updates
    EventsOn('encounters_updated', (encounters) => {
        updateEncountersList(encounters);
    });

    // Listen for switches
    EventsOn('switch_occurred', (status) => {
        addLog(`➡️ Switched to: ${status.CurrentPlayerName || status.CurrentPlayer}`);
        updateStatusWidget(status);
    });

    // Listen for errors
    EventsOn('error', (error) => {
        addLog(`⚠️ Error: ${error}`);
    });
}

function updateControlButtons() {
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');
    const statusEl = document.getElementById('runningStatus');

    if (isRunning) {
        startBtn.disabled = true;
        stopBtn.disabled = false;
        statusEl.textContent = 'Running';
        statusEl.className = 'status-value status-running';
    } else {
        startBtn.disabled = false;
        stopBtn.disabled = true;
        statusEl.textContent = 'Stopped';
        statusEl.className = 'status-value status-stopped';
    }
}

async function updateStatus() {
    try {
        const status = await GetStatus();
        updateStatusWidget(status);
    } catch (err) {
        console.error('Error getting status:', err);
    }
}

async function updatePlayers() {
    try {
        const players = await GetPlayers();
        updatePlayersTable(players);
    } catch (err) {
        console.error('Error getting players:', err);
    }
}

async function updateEncounters() {
    try {
        const encounters = await GetEncounters();
        updateEncountersList(encounters);
    } catch (err) {
        console.error('Error getting encounters:', err);
    }
}

async function updateStats() {
    try {
        const stats = await GetStatistics();
        updateStatisticsWidget(stats);
    } catch (err) {
        console.error('Error getting statistics:', err);
    }
}

function updateStatusWidget(status) {
    document.getElementById('runningStatus').textContent = status.IsRunning ? 'Running' : 'Stopped';
    document.getElementById('runningStatus').className = status.IsRunning ? 'status-value status-running' : 'status-value status-stopped';
    document.getElementById('currentPlayer').textContent = status.CurrentPlayerName || status.CurrentPlayer || '-';
    document.getElementById('totalSwitches').textContent = status.TotalSwitches || 0;
    document.getElementById('alivePlayers').textContent = status.AlivePlayers || 0;
    
    // Round information
    document.getElementById('roundNumber').textContent = status.RoundNumber || '-';
    document.getElementById('score').textContent = `${status.ScoreCT || 0} : ${status.ScoreT || 0}`;
    document.getElementById('roundPhase').textContent = status.RoundPhase || '-';
}

function updateStatisticsWidget(stats) {
    document.getElementById('sniperKills').textContent = stats.SniperKills || 0;
    document.getElementById('upsetVictories').textContent = stats.UpsetVictories || 0;
    document.getElementById('damageDetections').textContent = stats.DamageDetections || 0;
    document.getElementById('switchesPerMin').textContent = (stats.SwitchesPerMin || 0).toFixed(1);
}

function updatePlayersTable(players) {
    const scoreboard = document.getElementById('scoreboard');
    
    if (!players || players.length === 0) {
        scoreboard.innerHTML = '<div class="no-data">No players found</div>';
        return;
    }

    // Separate players by team
    const ctPlayers = players.filter(p => p.Team === 'CT');
    const tPlayers = players.filter(p => p.Team === 'T');
    
    // Sort by observer slot (stable sorting) - if slot is 0 or undefined, sort by name as fallback
    ctPlayers.sort((a, b) => {
        if (a.Slot && b.Slot) return a.Slot - b.Slot;
        if (a.Slot) return -1;
        if (b.Slot) return 1;
        return a.Name.localeCompare(b.Name);
    });
    tPlayers.sort((a, b) => {
        if (a.Slot && b.Slot) return a.Slot - b.Slot;
        if (a.Slot) return -1;
        if (b.Slot) return 1;
        return a.Name.localeCompare(b.Name);
    });
    
    const createPlayerRow = (p) => {
        const weapon = formatWeapon(p.ActiveWeapon);
        const healthPercent = p.Health;
        const healthClass = healthPercent > 75 ? 'hp-high' : healthPercent > 40 ? 'hp-mid' : 'hp-low';
        
        return `
            <div class="player-row">
                <div class="player-name">${escapeHtml(p.Name)}</div>
                <div class="player-stats">
                    <span class="stat-hp ${healthClass}">${p.Health} HP</span>
                    <span class="stat-armor">${p.Armor > 0 ? p.Armor + ' ARM' : ''}</span>
                    <span class="stat-kd">${p.Kills}/${p.Deaths}</span>
                    <span class="stat-weapon">${weapon}</span>
                    <span class="stat-money">$${(p.Money / 1000).toFixed(1)}k</span>
                </div>
            </div>
        `;
    };
    
    scoreboard.innerHTML = `
        ${ctPlayers.length > 0 ? `
            <div class="team-section team-ct-section">
                <div class="team-header team-ct-header">
                    <span class="team-name">COUNTER-TERRORISTS</span>
                    <span class="team-count">${ctPlayers.length} alive</span>
                </div>
                <div class="team-players">
                    ${ctPlayers.map(createPlayerRow).join('')}
                </div>
            </div>
        ` : ''}
        
        ${tPlayers.length > 0 ? `
            <div class="team-section team-t-section">
                <div class="team-header team-t-header">
                    <span class="team-name">TERRORISTS</span>
                    <span class="team-count">${tPlayers.length} alive</span>
                </div>
                <div class="team-players">
                    ${tPlayers.map(createPlayerRow).join('')}
                </div>
            </div>
        ` : ''}
    `;
}

function updateEncountersList(encounters) {
    const container = document.getElementById('encountersList');
    
    if (!encounters || encounters.length === 0) {
        container.innerHTML = '<div class="no-data">No encounters detected</div>';
        return;
    }

    container.innerHTML = encounters.slice(0, 5).map((enc, idx) => `
        <div class="encounter-item ${enc.IsCurrentPlayer ? 'current-encounter' : ''}">
            <div class="encounter-rank">#${idx + 1}</div>
            <div class="encounter-details">
                <div class="encounter-players">
                    <span class="team-ct">${escapeHtml(enc.Player1)}</span>
                    <span class="vs">vs</span>
                    <span class="team-t">${escapeHtml(enc.Player2)}</span>
                </div>
                <div class="encounter-stats">
                    <span>Distance: ${enc.Distance.toFixed(0)} units</span>
                    <span class="priority-badge">Priority: ${enc.Priority.toFixed(1)}</span>
                </div>
            </div>
        </div>
    `).join('');
}

function addLog(message) {
    const timestamp = new Date().toLocaleTimeString();
    logs.unshift({ time: timestamp, message });
    
    // Keep only last MAX_LOGS entries
    if (logs.length > MAX_LOGS) {
        logs = logs.slice(0, MAX_LOGS);
    }
    
    updateLogsDisplay();
}

function updateLogsDisplay() {
    const container = document.getElementById('logsContainer');
    container.innerHTML = logs.map(log => `
        <div class="log-entry">
            <span class="log-time">${log.time}</span>
            <span class="log-message">${escapeHtml(log.message)}</span>
        </div>
    `).join('');
}

function formatWeapon(weapon) {
    if (!weapon) return '-';
    
    // Remove weapon_ prefix and convert to uppercase
    let name = weapon.replace('weapon_', '').toUpperCase();
    
    // Shorten common weapon names for compact display
    const shortNames = {
        'M4A1_SILENCER': 'M4A1-S',
        'USP_SILENCER': 'USP-S',
        'HEGRENADE': 'HE',
        'FLASHBANG': 'FLASH',
        'SMOKEGRENADE': 'SMOKE',
        'MOLOTOV': 'FIRE',
        'INCGRENADE': 'FIRE',
        'DECOY': 'DECOY',
        'KNIFE': 'KNIFE',
        'KNIFE_T': 'KNIFE'
    };
    
    return shortNames[name] || name;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ============================================================================
// SETTINGS PAGE
// ============================================================================

function initSettings() {
    document.querySelector('#app').innerHTML = `
        <div class="dashboard">
            <!-- Header -->
            <header class="dashboard-header">
                <div class="header-content">
                    <h1>⚙️ Settings</h1>
                    <div class="header-controls">
                        <button id="backBtn" class="btn btn-secondary">← Back to Dashboard</button>
                    </div>
                </div>
            </header>

            <!-- Settings Content -->
            <div class="settings-content">
                <div class="settings-container">
                    <form id="settingsForm" class="settings-form">
                        <!-- Camera Priority Bonuses -->
                        <div class="settings-section">
                            <h2>🎯 Camera Priority Bonuses</h2>
                            <div class="settings-grid">
                                <div class="setting-item">
                                    <label for="awpBonus">AWP Bonus</label>
                                    <input type="number" id="awpBonus" step="1" min="0" max="100">
                                    <span class="setting-help">Bonus points for AWP sniper rifle</span>
                                </div>
                                <div class="setting-item">
                                    <label for="scoutBonus">Scout Bonus</label>
                                    <input type="number" id="scoutBonus" step="1" min="0" max="100">
                                    <span class="setting-help">Bonus points for Scout sniper rifle</span>
                                </div>
                                <div class="setting-item">
                                    <label for="ak47Bonus">AK-47 Bonus</label>
                                    <input type="number" id="ak47Bonus" step="1" min="0" max="50">
                                    <span class="setting-help">Bonus points for AK-47 over M4s</span>
                                </div>
                                <div class="setting-item">
                                    <label for="damageDealtBonus">Damage Dealt Bonus</label>
                                    <input type="number" id="damageDealtBonus" step="1" min="0" max="200">
                                    <span class="setting-help">Bonus for dealing significant damage</span>
                                </div>
                                <div class="setting-item">
                                    <label for="upsetVictoryBonus">Upset Victory Bonus</label>
                                    <input type="number" id="upsetVictoryBonus" step="1" min="0" max="300">
                                    <span class="setting-help">Bonus for unexpected combat victories</span>
                                </div>
                                <div class="setting-item">
                                    <label for="sniperKillBonus">Sniper Kill Bonus</label>
                                    <input type="number" id="sniperKillBonus" step="1" min="0" max="300">
                                    <span class="setting-help">Bonus for sniper kills (AWP gets 1.5x)</span>
                                </div>
                            </div>
                        </div>

                        <!-- Duration Settings -->
                        <div class="settings-section">
                            <h2>⏱️ Bonus Durations (seconds)</h2>
                            <div class="settings-grid">
                                <div class="setting-item">
                                    <label for="damageDealtDuration">Damage Dealt Duration</label>
                                    <input type="number" id="damageDealtDuration" step="1" min="1" max="30">
                                    <span class="setting-help">How long damage dealt bonus lasts</span>
                                </div>
                                <div class="setting-item">
                                    <label for="upsetVictoryDuration">Upset Victory Duration</label>
                                    <input type="number" id="upsetVictoryDuration" step="1" min="1" max="30">
                                    <span class="setting-help">How long upset victory bonus lasts</span>
                                </div>
                                <div class="setting-item">
                                    <label for="sniperKillDuration">Sniper Kill Duration</label>
                                    <input type="number" id="sniperKillDuration" step="1" min="1" max="30">
                                    <span class="setting-help">How long sniper kill bonus lasts</span>
                                </div>
                            </div>
                        </div>

                        <!-- Distance Limits -->
                        <div class="settings-section">
                            <h2>📏 Distance Limits (game units)</h2>
                            <div class="settings-grid">
                                <div class="setting-item">
                                    <label for="maxEncounterDistance">Max Encounter Distance</label>
                                    <input type="number" id="maxEncounterDistance" step="100" min="500" max="5000">
                                    <span class="setting-help">Maximum distance to consider an encounter</span>
                                </div>
                                <div class="setting-item">
                                    <label for="maxDamageDistNormal">Max Damage Distance (Normal)</label>
                                    <input type="number" id="maxDamageDistNormal" step="100" min="500" max="5000">
                                    <span class="setting-help">Max distance for damage detection (rifles)</span>
                                </div>
                                <div class="setting-item">
                                    <label for="maxDamageDistSniper">Max Damage Distance (Sniper)</label>
                                    <input type="number" id="maxDamageDistSniper" step="100" min="1000" max="10000">
                                    <span class="setting-help">Max distance for damage detection (snipers)</span>
                                </div>
                            </div>
                        </div>

                        <!-- Advanced Settings -->
                        <div class="settings-section">
                            <h2>🔧 Advanced Settings</h2>
                            <div class="settings-grid">
                                <div class="setting-item">
                                    <label for="minHealthLoss">Minimum Health Loss</label>
                                    <input type="number" id="minHealthLoss" step="5" min="5" max="100">
                                    <span class="setting-help">Minimum HP loss to trigger damage detection</span>
                                </div>
                                <div class="setting-item">
                                    <label for="verticalDiffThreshold">Vertical Diff Threshold</label>
                                    <input type="number" id="verticalDiffThreshold" step="10" min="50" max="1000">
                                    <span class="setting-help">Max vertical distance for encounters (different floors)</span>
                                </div>
                            </div>
                        </div>

                        <!-- Action Buttons -->
                        <div class="settings-actions">
                            <div class="settings-actions-left">
                                <button type="button" id="exportBtn" class="btn btn-secondary">📤 Export Settings</button>
                                <button type="button" id="importBtn" class="btn btn-secondary">📥 Import Settings</button>
                            </div>
                            <div class="settings-actions-right">
                                <button type="button" id="resetBtn" class="btn btn-danger">🔄 Reset to Defaults</button>
                                <button type="submit" class="btn btn-success">💾 Save Settings</button>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;

    // Load current settings
    loadSettings();
}

async function loadSettings() {
    try {
        const settings = await GetSettings();
        document.getElementById('awpBonus').value = settings.awp_bonus;
        document.getElementById('scoutBonus').value = settings.scout_bonus;
        document.getElementById('ak47Bonus').value = settings.ak47_bonus;
        document.getElementById('damageDealtBonus').value = settings.damage_dealt_bonus;
        document.getElementById('upsetVictoryBonus').value = settings.upset_victory_bonus;
        document.getElementById('sniperKillBonus').value = settings.sniper_kill_bonus;
        document.getElementById('damageDealtDuration').value = settings.damage_dealt_duration;
        document.getElementById('upsetVictoryDuration').value = settings.upset_victory_duration;
        document.getElementById('sniperKillDuration').value = settings.sniper_kill_duration;
        document.getElementById('maxEncounterDistance').value = settings.max_encounter_distance;
        document.getElementById('maxDamageDistNormal').value = settings.max_damage_dist_normal;
        document.getElementById('maxDamageDistSniper').value = settings.max_damage_dist_sniper;
        document.getElementById('minHealthLoss').value = settings.min_health_loss;
        document.getElementById('verticalDiffThreshold').value = settings.vertical_diff_threshold;
    } catch (err) {
        console.error('Failed to load settings:', err);
        alert('Failed to load settings: ' + err);
    }
}

function setupSettingsListeners() {
    document.getElementById('backBtn').addEventListener('click', () => {
        showPage('dashboard');
    });

    document.getElementById('settingsForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const settings = {
            awp_bonus: parseFloat(document.getElementById('awpBonus').value),
            scout_bonus: parseFloat(document.getElementById('scoutBonus').value),
            ak47_bonus: parseFloat(document.getElementById('ak47Bonus').value),
            damage_dealt_bonus: parseFloat(document.getElementById('damageDealtBonus').value),
            upset_victory_bonus: parseFloat(document.getElementById('upsetVictoryBonus').value),
            sniper_kill_bonus: parseFloat(document.getElementById('sniperKillBonus').value),
            damage_dealt_duration: parseInt(document.getElementById('damageDealtDuration').value),
            upset_victory_duration: parseInt(document.getElementById('upsetVictoryDuration').value),
            sniper_kill_duration: parseInt(document.getElementById('sniperKillDuration').value),
            max_encounter_distance: parseFloat(document.getElementById('maxEncounterDistance').value),
            max_damage_dist_normal: parseFloat(document.getElementById('maxDamageDistNormal').value),
            max_damage_dist_sniper: parseFloat(document.getElementById('maxDamageDistSniper').value),
            min_health_loss: parseInt(document.getElementById('minHealthLoss').value),
            vertical_diff_threshold: parseFloat(document.getElementById('verticalDiffThreshold').value),
        };

        try {
            await SaveSettings(settings);
            alert('✅ Settings saved successfully!');
        } catch (err) {
            console.error('Failed to save settings:', err);
            alert('Failed to save settings: ' + err);
        }
    });

    // Export Button
    document.getElementById('exportBtn').addEventListener('click', async () => {
        try {
            await ExportSettings();
            alert('✅ Settings exported successfully!');
        } catch (err) {
            if (!err.toString().includes('cancelled')) {
                console.error('Failed to export settings:', err);
                alert('Failed to export settings: ' + err);
            }
        }
    });

    // Import Button
    document.getElementById('importBtn').addEventListener('click', async () => {
        try {
            await ImportSettings();
            await loadSettings(); // Reload to show imported settings
            alert('✅ Settings imported successfully!');
        } catch (err) {
            if (!err.toString().includes('cancelled')) {
                console.error('Failed to import settings:', err);
                alert('Failed to import settings: ' + err);
            }
        }
    });

    // Reset Button
    document.getElementById('resetBtn').addEventListener('click', async () => {
        if (confirm('Are you sure you want to reset all settings to defaults?')) {
            try {
                await ResetSettings();
                await loadSettings(); // Reload to show defaults
                alert('✅ Settings reset to defaults!');
            } catch (err) {
                console.error('Failed to reset settings:', err);
                alert('Failed to reset settings: ' + err);
            }
        }
    });
}

function startDataPolling() {
    // Poll data every 500ms
    setInterval(() => {
        if (isRunning) {
            updateStatus();
            updatePlayers();
            updateEncounters();
            updateStats();
        }
    }, 500);
}

