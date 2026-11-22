export namespace director {
	
	export class Position {
	    X: number;
	    Y: number;
	    Z: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.Z = source["Z"];
	    }
	}
	export class PlayerInfo {
	    SteamID: string;
	    Name: string;
	    Team: string;
	    Slot: number;
	    Position: Position;
	    Health: number;
	    Armor: number;
	    Money: number;
	    Kills: number;
	    Deaths: number;
	    EquipmentValue: number;
	    HasDefuser: boolean;
	    ActiveWeapon: string;
	    Forward: Position;
	    HasForward: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlayerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SteamID = source["SteamID"];
	        this.Name = source["Name"];
	        this.Team = source["Team"];
	        this.Slot = source["Slot"];
	        this.Position = this.convertValues(source["Position"], Position);
	        this.Health = source["Health"];
	        this.Armor = source["Armor"];
	        this.Money = source["Money"];
	        this.Kills = source["Kills"];
	        this.Deaths = source["Deaths"];
	        this.EquipmentValue = source["EquipmentValue"];
	        this.HasDefuser = source["HasDefuser"];
	        this.ActiveWeapon = source["ActiveWeapon"];
	        this.Forward = this.convertValues(source["Forward"], Position);
	        this.HasForward = source["HasForward"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Settings {
	    awp_bonus: number;
	    scout_bonus: number;
	    ak47_bonus: number;
	    damage_dealt_bonus: number;
	    upset_victory_bonus: number;
	    sniper_kill_bonus: number;
	    damage_dealt_duration: number;
	    upset_victory_duration: number;
	    sniper_kill_duration: number;
	    max_encounter_distance: number;
	    max_damage_dist_normal: number;
	    max_damage_dist_sniper: number;
	    min_health_loss: number;
	    vertical_diff_threshold: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.awp_bonus = source["awp_bonus"];
	        this.scout_bonus = source["scout_bonus"];
	        this.ak47_bonus = source["ak47_bonus"];
	        this.damage_dealt_bonus = source["damage_dealt_bonus"];
	        this.upset_victory_bonus = source["upset_victory_bonus"];
	        this.sniper_kill_bonus = source["sniper_kill_bonus"];
	        this.damage_dealt_duration = source["damage_dealt_duration"];
	        this.upset_victory_duration = source["upset_victory_duration"];
	        this.sniper_kill_duration = source["sniper_kill_duration"];
	        this.max_encounter_distance = source["max_encounter_distance"];
	        this.max_damage_dist_normal = source["max_damage_dist_normal"];
	        this.max_damage_dist_sniper = source["max_damage_dist_sniper"];
	        this.min_health_loss = source["min_health_loss"];
	        this.vertical_diff_threshold = source["vertical_diff_threshold"];
	    }
	}

}

export namespace main {
	
	export class EncounterInfo {
	    Player1: string;
	    Player2: string;
	    Player1Team: string;
	    Player2Team: string;
	    Distance: number;
	    Priority: number;
	    IsCurrentPlayer: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EncounterInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Player1 = source["Player1"];
	        this.Player2 = source["Player2"];
	        this.Player1Team = source["Player1Team"];
	        this.Player2Team = source["Player2Team"];
	        this.Distance = source["Distance"];
	        this.Priority = source["Priority"];
	        this.IsCurrentPlayer = source["IsCurrentPlayer"];
	    }
	}
	export class Statistics {
	    TotalSwitches: number;
	    SwitchesPerMin: number;
	    Uptime: number;
	    // Go type: time
	    StartTime: any;
	    // Go type: time
	    LastSwitchTime: any;
	    SniperKills: number;
	    UpsetVictories: number;
	    DamageDetections: number;
	
	    static createFrom(source: any = {}) {
	        return new Statistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TotalSwitches = source["TotalSwitches"];
	        this.SwitchesPerMin = source["SwitchesPerMin"];
	        this.Uptime = source["Uptime"];
	        this.StartTime = this.convertValues(source["StartTime"], null);
	        this.LastSwitchTime = this.convertValues(source["LastSwitchTime"], null);
	        this.SniperKills = source["SniperKills"];
	        this.UpsetVictories = source["UpsetVictories"];
	        this.DamageDetections = source["DamageDetections"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StatusInfo {
	    IsRunning: boolean;
	    CurrentPlayer: string;
	    CurrentPlayerName: string;
	    CurrentEncounter: string;
	    LastSwitch: string;
	    Uptime: string;
	    TotalSwitches: number;
	    RoundPhase: string;
	    AlivePlayers: number;
	    RoundTime: string;
	    RoundNumber: number;
	    ScoreCT: number;
	    ScoreT: number;
	
	    static createFrom(source: any = {}) {
	        return new StatusInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IsRunning = source["IsRunning"];
	        this.CurrentPlayer = source["CurrentPlayer"];
	        this.CurrentPlayerName = source["CurrentPlayerName"];
	        this.CurrentEncounter = source["CurrentEncounter"];
	        this.LastSwitch = source["LastSwitch"];
	        this.Uptime = source["Uptime"];
	        this.TotalSwitches = source["TotalSwitches"];
	        this.RoundPhase = source["RoundPhase"];
	        this.AlivePlayers = source["AlivePlayers"];
	        this.RoundTime = source["RoundTime"];
	        this.RoundNumber = source["RoundNumber"];
	        this.ScoreCT = source["ScoreCT"];
	        this.ScoreT = source["ScoreT"];
	    }
	}

}

