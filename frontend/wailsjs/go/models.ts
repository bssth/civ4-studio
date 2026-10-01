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
	    group?: string;
	    color?: string;

	    static createFrom(source: any = {}) {
	        return new EnumOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	        this.group = source["group"];
	        this.color = source["color"];
	    }
	}

	export class CivilizationOption {
	    type: string;
	    description: string;
	    short_description: string;
	    adjective: string;
	    color: string;
	    art_style: string;
	    leaders: string[];
	    playable: boolean;

	    static createFrom(source: any = {}) {
	        return new CivilizationOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	        this.short_description = source["short_description"];
	        this.adjective = source["adjective"];
	        this.color = source["color"];
	        this.art_style = source["art_style"];
	        this.leaders = source["leaders"] ?? [];
	        this.playable = source["playable"];
	    }
	}

	export class MapInfo {
	    path: string;
	    version: number;
	    teams_count: number;
	    player_count: number;
	    plot_count: number;
	    dirty: boolean;

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
	        this.dirty = source["dirty"];
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

	// MapProps maps the editor.MapProps struct (BeginMap section), default Go JSON field names
	export class MapProps {
	    GridWidth: number;
	    GridHeight: number;
	    TopLatitude: number;
	    BottomLatitude: number;
	    WrapX: number;
	    WrapY: number;
	    WorldSize: string;
	    Climate: string;
	    SeaLevel: string;
	    NumPlotsWritten: number;
	    NumSignsWritten: number;
	    RandomizeResources: boolean;
	    Extra?: string[];

	    static createFrom(source: any = {}) {
	        return new MapProps(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.GridWidth = source["GridWidth"] ?? 0;
	        this.GridHeight = source["GridHeight"] ?? 0;
	        this.TopLatitude = source["TopLatitude"] ?? 90;
	        this.BottomLatitude = source["BottomLatitude"] ?? -90;
	        this.WrapX = source["WrapX"] ?? 0;
	        this.WrapY = source["WrapY"] ?? 0;
	        this.WorldSize = source["WorldSize"] ?? "";
	        this.Climate = source["Climate"] ?? "";
	        this.SeaLevel = source["SeaLevel"] ?? "";
	        this.NumPlotsWritten = source["NumPlotsWritten"] ?? 0;
	        this.NumSignsWritten = source["NumSignsWritten"] ?? 0;
	        this.RandomizeResources = source["RandomizeResources"] ?? false;
	        this.Extra = source["Extra"];
	    }
	}

	export interface CountStat {
	    type: string;
	    count: number;
	}

	export interface MapStats {
	    plots: number;
	    expected_plots: number;
	    peaks: number;
	    hills: number;
	    flat: number;
	    water: number;
	    cities: number;
	    units: number;
	    signs: number;
	    starting_plots: number;
	    bonuses: number;
	    terrains: CountStat[] | null;
	    problems: string[];
	}

	export interface WorldSizeOption {
	    type: string;
	    description: string;
	    width: number;
	    height: number;
	}

	export interface MapView {
	    width: number;
	    height: number;
	    terrains: string[];
	    features: string[];
	    bonuses: string[];
	    terrain: number[];
	    plot_type: number[];
	    feature: number[];
	    bonus: number[];
	    flags: number[];
	    city_owner: number[];
	    unit_owner: number[];
	    unit_count: number[];
	}

	// Plot, City and Unit map the Go structs with default JSON field names
	export interface Unit {
	    UnitType: string;
	    UnitOwner: number;
	    Level: number;
	    Experience: number;
	    PromotionType: string[] | null;
	    UnitAIType: string;
	    Damage: number;
	    FacingDirection: number;
	    Extra?: string[] | null;
	}

	export interface City {
	    CityOwner: number;
	    CityName: string;
	    CityPopulation: number;
	    ProductionUnit: string;
	    ProductionBuilding: string;
	    ProductionProject: string;
	    ProductionProcess: string;
	    BuildingType: string[] | null;
	    ReligionType: string[] | null;
	    HolyCityReligionType: string[] | null;
	    ScriptData: string;
	    PlayerCulture: Record<string, number> | null;
	    Extra?: string[] | null;
	}

	export interface Plot {
	    X: number;
	    Y: number;
	    Landmark: string;
	    ScriptData: string;
	    IsNOfRiver: boolean;
	    IsWOfRiver: boolean;
	    RiverNSDirection: number;
	    RiverWEDirection: number;
	    StartingPlot: boolean;
	    BonusType: string;
	    ImprovementType: string;
	    FeatureType: string[] | null;
	    FeatureVariety: string[] | null;
	    RouteType: string;
	    TerrainType: string;
	    PlotType: number;
	    Units: Unit[] | null;
	    Cities: City[] | null;
	    TeamReveal: number[] | null;
	    Extra?: string[] | null;
	}

}
