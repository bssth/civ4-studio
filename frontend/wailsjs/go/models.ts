export namespace editor {

	export class Config {
	    game_dir: string;
	    mod: string;
	    auto_save: boolean;

	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_dir = source["game_dir"];
	        this.mod = source["mod"];
	        this.auto_save = source["auto_save"];
	    }
	}

	export class EnumOption {
	    type: string;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new EnumOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	    }
	}

	export class MapInfo {
	    path: string;
	    version: number;
	    teams_count: number;
	    player_count: number;
	    plot_count: number;

	    static createFrom(source: any = {}) {
	        return new MapInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.version = source["version"];
	        this.teams_count = source["teams_count"];
	        this.player_count = source["player_count"];
	        this.plot_count = source["plot_count"];
	    }
	}

	// Game maps the editor.Game struct. Field names match the Go struct
	// because Game uses default Go JSON marshalling (no json tags).
	export class Game {
	    Era: string;
	    Speed: string;
	    Calendar: string;
	    Victory: string[] | null;
	    GameTurn: number;
	    MaxCityElimination: number;
	    NumAdvancedStartPoints: number;
	    TargetScore: number;
	    StartYear: number;
	    Description: string;
	    ModPath: string;
	    Tutorial: boolean;
	    Option: string[] | null;
	    MPOption: string[] | null;
	    ForceControl: string[] | null;
	    MaxTurns: number;
	    Extra?: string[];

	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Era = source["Era"] ?? "";
	        this.Speed = source["Speed"] ?? "";
	        this.Calendar = source["Calendar"] ?? "";
	        this.Victory = source["Victory"] ?? [];
	        this.GameTurn = source["GameTurn"] ?? 0;
	        this.MaxCityElimination = source["MaxCityElimination"] ?? 0;
	        this.NumAdvancedStartPoints = source["NumAdvancedStartPoints"] ?? 0;
	        this.TargetScore = source["TargetScore"] ?? 0;
	        this.StartYear = source["StartYear"] ?? 0;
	        this.Description = source["Description"] ?? "";
	        this.ModPath = source["ModPath"] ?? "";
	        this.Tutorial = source["Tutorial"] ?? false;
	        this.Option = source["Option"] ?? [];
	        this.MPOption = source["MPOption"] ?? [];
	        this.ForceControl = source["ForceControl"] ?? [];
	        this.MaxTurns = source["MaxTurns"] ?? 0;
	        this.Extra = source["Extra"];
	    }
	}

	export class Team {
	    TeamID: number;
	    Tech: string[] | null;
	    ContactWithTeam: number[] | null;
	    AtWar: number[] | null;
	    PermanentWarPeace: number[] | null;
	    OpenBordersWithTeam: number[] | null;
	    DefensivePactWithTeam: number[] | null;
	    ProjectType: string[] | null;
	    RevealMap: boolean;
	    Extra?: string[];

	    static createFrom(source: any = {}) {
	        return new Team(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TeamID = source["TeamID"] ?? 0;
	        this.Tech = source["Tech"] ?? [];
	        this.ContactWithTeam = source["ContactWithTeam"] ?? [];
	        this.AtWar = source["AtWar"] ?? [];
	        this.PermanentWarPeace = source["PermanentWarPeace"] ?? [];
	        this.OpenBordersWithTeam = source["OpenBordersWithTeam"] ?? [];
	        this.DefensivePactWithTeam = source["DefensivePactWithTeam"] ?? [];
	        this.ProjectType = source["ProjectType"] ?? [];
	        this.RevealMap = source["RevealMap"] ?? false;
	        this.Extra = source["Extra"];
	    }
	}

	export class Player {
	    CivDesc: string;
	    CivShortDesc: string;
	    LeaderName: string;
	    CivAdjective: string;
	    FlagDecal: string;
	    WhiteFlag: boolean;
	    LeaderType: string;
	    CivType: string;
	    Team: number;
	    Handicap: string;
	    Color: string;
	    ArtStyle: string;
	    PlayableCiv: boolean;
	    MinorNationStatus: boolean;
	    StartingGold: number;
	    RandomStartLocation: boolean;
	    StartingX: number;
	    StartingY: number;
	    StateReligion: string;
	    StartingEra: string;
	    CityList: string[] | null;
	    CivicOption: string[] | null;
	    Civic: string[] | null;
	    AttitudePlayer: number[] | null;
	    AttitudeExtra: number[] | null;
	    Extra?: string[];

	    static createFrom(source: any = {}) {
	        return new Player(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CivDesc = source["CivDesc"] ?? "";
	        this.CivShortDesc = source["CivShortDesc"] ?? "";
	        this.LeaderName = source["LeaderName"] ?? "";
	        this.CivAdjective = source["CivAdjective"] ?? "";
	        this.FlagDecal = source["FlagDecal"] ?? "";
	        this.WhiteFlag = source["WhiteFlag"] ?? false;
	        this.LeaderType = source["LeaderType"] ?? "";
	        this.CivType = source["CivType"] ?? "";
	        this.Team = source["Team"] ?? 0;
	        this.Handicap = source["Handicap"] ?? "";
	        this.Color = source["Color"] ?? "";
	        this.ArtStyle = source["ArtStyle"] ?? "";
	        this.PlayableCiv = source["PlayableCiv"] ?? false;
	        this.MinorNationStatus = source["MinorNationStatus"] ?? false;
	        this.StartingGold = source["StartingGold"] ?? 0;
	        this.RandomStartLocation = source["RandomStartLocation"] ?? false;
	        this.StartingX = source["StartingX"] ?? 0;
	        this.StartingY = source["StartingY"] ?? 0;
	        this.StateReligion = source["StateReligion"] ?? "";
	        this.StartingEra = source["StartingEra"] ?? "";
	        this.CityList = source["CityList"] ?? [];
	        this.CivicOption = source["CivicOption"] ?? [];
	        this.Civic = source["Civic"] ?? [];
	        this.AttitudePlayer = source["AttitudePlayer"] ?? [];
	        this.AttitudeExtra = source["AttitudeExtra"] ?? [];
	        this.Extra = source["Extra"];
	    }
	}

}
