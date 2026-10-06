export namespace editor {
	
	export class City {
	    CityOwner: number;
	    CityName: string;
	    CityPopulation: number;
	    ProductionUnit: string;
	    ProductionBuilding: string;
	    ProductionProject: string;
	    ProductionProcess: string;
	    BuildingType: string[];
	    ReligionType: string[];
	    HolyCityReligionType: string[];
	    ScriptData: string;
	    PlayerCulture: Record<number, number>;
	
	    static createFrom(source: any = {}) {
	        return new City(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CityOwner = source["CityOwner"];
	        this.CityName = source["CityName"];
	        this.CityPopulation = source["CityPopulation"];
	        this.ProductionUnit = source["ProductionUnit"];
	        this.ProductionBuilding = source["ProductionBuilding"];
	        this.ProductionProject = source["ProductionProject"];
	        this.ProductionProcess = source["ProductionProcess"];
	        this.BuildingType = source["BuildingType"];
	        this.ReligionType = source["ReligionType"];
	        this.HolyCityReligionType = source["HolyCityReligionType"];
	        this.ScriptData = source["ScriptData"];
	        this.PlayerCulture = source["PlayerCulture"];
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
	        this.leaders = source["leaders"];
	        this.playable = source["playable"];
	    }
	}
	export class ClipboardInfo {
	    width: number;
	    height: number;
	    cities: number;
	    units: number;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.cities = source["cities"];
	        this.units = source["units"];
	    }
	}
	export class Config {
	    game_dir: string;
	    mod: string;
	    auto_save: boolean;
	    language: string;
	    ui_language: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_dir = source["game_dir"];
	        this.mod = source["mod"];
	        this.auto_save = source["auto_save"];
	        this.language = source["language"];
	        this.ui_language = source["ui_language"];
	    }
	}
	export class CountStat {
	    type: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new CountStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.count = source["count"];
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
	export class Game {
	    Era: string;
	    Speed: string;
	    Calendar: string;
	    Victory: string[];
	    GameTurn: number;
	    MaxCityElimination: number;
	    NumAdvancedStartPoints: number;
	    TargetScore: number;
	    StartYear: number;
	    Description: string;
	    ModPath: string;
	    Tutorial: boolean;
	    Option: string[];
	    MPOption: string[];
	    ForceControl: string[];
	    MaxTurns: number;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Era = source["Era"];
	        this.Speed = source["Speed"];
	        this.Calendar = source["Calendar"];
	        this.Victory = source["Victory"];
	        this.GameTurn = source["GameTurn"];
	        this.MaxCityElimination = source["MaxCityElimination"];
	        this.NumAdvancedStartPoints = source["NumAdvancedStartPoints"];
	        this.TargetScore = source["TargetScore"];
	        this.StartYear = source["StartYear"];
	        this.Description = source["Description"];
	        this.ModPath = source["ModPath"];
	        this.Tutorial = source["Tutorial"];
	        this.Option = source["Option"];
	        this.MPOption = source["MPOption"];
	        this.ForceControl = source["ForceControl"];
	        this.MaxTurns = source["MaxTurns"];
	    }
	}
	export class HistoryState {
	    undo: string;
	    redo: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.undo = source["undo"];
	        this.redo = source["redo"];
	    }
	}
	export class LanguageOption {
	    name: string;
	    texts: number;
	
	    static createFrom(source: any = {}) {
	        return new LanguageOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.texts = source["texts"];
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
	
	    static createFrom(source: any = {}) {
	        return new MapProps(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.GridWidth = source["GridWidth"];
	        this.GridHeight = source["GridHeight"];
	        this.TopLatitude = source["TopLatitude"];
	        this.BottomLatitude = source["BottomLatitude"];
	        this.WrapX = source["WrapX"];
	        this.WrapY = source["WrapY"];
	        this.WorldSize = source["WorldSize"];
	        this.Climate = source["Climate"];
	        this.SeaLevel = source["SeaLevel"];
	        this.NumPlotsWritten = source["NumPlotsWritten"];
	        this.NumSignsWritten = source["NumSignsWritten"];
	        this.RandomizeResources = source["RandomizeResources"];
	    }
	}
	export class Problem {
	    severity: string;
	    section: string;
	    code: string;
	    args: Record<string, string>;
	    message: string;
	    x: number;
	    y: number;
	    player: number;
	    team: number;
	
	    static createFrom(source: any = {}) {
	        return new Problem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.section = source["section"];
	        this.code = source["code"];
	        this.args = source["args"];
	        this.message = source["message"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.player = source["player"];
	        this.team = source["team"];
	    }
	}
	export class MapStats {
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
	    terrains: CountStat[];
	    problems: Problem[];
	
	    static createFrom(source: any = {}) {
	        return new MapStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plots = source["plots"];
	        this.expected_plots = source["expected_plots"];
	        this.peaks = source["peaks"];
	        this.hills = source["hills"];
	        this.flat = source["flat"];
	        this.water = source["water"];
	        this.cities = source["cities"];
	        this.units = source["units"];
	        this.signs = source["signs"];
	        this.starting_plots = source["starting_plots"];
	        this.bonuses = source["bonuses"];
	        this.terrains = this.convertValues(source["terrains"], CountStat);
	        this.problems = this.convertValues(source["problems"], Problem);
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
	export class MapView {
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
	
	    static createFrom(source: any = {}) {
	        return new MapView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.terrains = source["terrains"];
	        this.features = source["features"];
	        this.bonuses = source["bonuses"];
	        this.terrain = source["terrain"];
	        this.plot_type = source["plot_type"];
	        this.feature = source["feature"];
	        this.bonus = source["bonus"];
	        this.flags = source["flags"];
	        this.city_owner = source["city_owner"];
	        this.unit_owner = source["unit_owner"];
	        this.unit_count = source["unit_count"];
	    }
	}
	export class OptionList {
	    key: string;
	    options: EnumOption[];
	
	    static createFrom(source: any = {}) {
	        return new OptionList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.options = this.convertValues(source["options"], EnumOption);
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
	export class PlotXY {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new PlotXY(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class PaintOp {
	    cells: PlotXY[];
	    terrain?: string;
	    plot_type?: number;
	    feature?: string;
	    feature_variety: number;
	    bonus?: string;
	    improvement?: string;
	    route?: string;
	
	    static createFrom(source: any = {}) {
	        return new PaintOp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cells = this.convertValues(source["cells"], PlotXY);
	        this.terrain = source["terrain"];
	        this.plot_type = source["plot_type"];
	        this.feature = source["feature"];
	        this.feature_variety = source["feature_variety"];
	        this.bonus = source["bonus"];
	        this.improvement = source["improvement"];
	        this.route = source["route"];
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
	    CityList: string[];
	    CivicOption: string[];
	    Civic: string[];
	    AttitudePlayer: number[];
	    AttitudeExtra: number[];
	
	    static createFrom(source: any = {}) {
	        return new Player(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CivDesc = source["CivDesc"];
	        this.CivShortDesc = source["CivShortDesc"];
	        this.LeaderName = source["LeaderName"];
	        this.CivAdjective = source["CivAdjective"];
	        this.FlagDecal = source["FlagDecal"];
	        this.WhiteFlag = source["WhiteFlag"];
	        this.LeaderType = source["LeaderType"];
	        this.CivType = source["CivType"];
	        this.Team = source["Team"];
	        this.Handicap = source["Handicap"];
	        this.Color = source["Color"];
	        this.ArtStyle = source["ArtStyle"];
	        this.PlayableCiv = source["PlayableCiv"];
	        this.MinorNationStatus = source["MinorNationStatus"];
	        this.StartingGold = source["StartingGold"];
	        this.RandomStartLocation = source["RandomStartLocation"];
	        this.StartingX = source["StartingX"];
	        this.StartingY = source["StartingY"];
	        this.StateReligion = source["StateReligion"];
	        this.StartingEra = source["StartingEra"];
	        this.CityList = source["CityList"];
	        this.CivicOption = source["CivicOption"];
	        this.Civic = source["Civic"];
	        this.AttitudePlayer = source["AttitudePlayer"];
	        this.AttitudeExtra = source["AttitudeExtra"];
	    }
	}
	export class Unit {
	    UnitType: string;
	    UnitOwner: number;
	    Level: number;
	    Experience: number;
	    PromotionType: string[];
	    UnitAIType: string;
	    Damage: number;
	    FacingDirection: number;
	
	    static createFrom(source: any = {}) {
	        return new Unit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UnitType = source["UnitType"];
	        this.UnitOwner = source["UnitOwner"];
	        this.Level = source["Level"];
	        this.Experience = source["Experience"];
	        this.PromotionType = source["PromotionType"];
	        this.UnitAIType = source["UnitAIType"];
	        this.Damage = source["Damage"];
	        this.FacingDirection = source["FacingDirection"];
	    }
	}
	export class Plot {
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
	    FeatureType: string[];
	    FeatureVariety: string[];
	    RouteType: string;
	    TerrainType: string;
	    PlotType: number;
	    Units: Unit[];
	    Cities: City[];
	    TeamReveal: number[];
	
	    static createFrom(source: any = {}) {
	        return new Plot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.Landmark = source["Landmark"];
	        this.ScriptData = source["ScriptData"];
	        this.IsNOfRiver = source["IsNOfRiver"];
	        this.IsWOfRiver = source["IsWOfRiver"];
	        this.RiverNSDirection = source["RiverNSDirection"];
	        this.RiverWEDirection = source["RiverWEDirection"];
	        this.StartingPlot = source["StartingPlot"];
	        this.BonusType = source["BonusType"];
	        this.ImprovementType = source["ImprovementType"];
	        this.FeatureType = source["FeatureType"];
	        this.FeatureVariety = source["FeatureVariety"];
	        this.RouteType = source["RouteType"];
	        this.TerrainType = source["TerrainType"];
	        this.PlotType = source["PlotType"];
	        this.Units = this.convertValues(source["Units"], Unit);
	        this.Cities = this.convertValues(source["Cities"], City);
	        this.TeamReveal = source["TeamReveal"];
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
	
	
	export class Region {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new Region(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class ResizeResult {
	    units: number;
	    cities: number;
	    signs: number;
	    starts: number[];
	
	    static createFrom(source: any = {}) {
	        return new ResizeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.units = source["units"];
	        this.cities = source["cities"];
	        this.signs = source["signs"];
	        this.starts = source["starts"];
	    }
	}
	export class SearchResult {
	    kind: string;
	    label: string;
	    x: number;
	    y: number;
	    player: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.player = source["player"];
	    }
	}
	export class Sign {
	    PlotX: number;
	    PlotY: number;
	    PlayerType: number;
	    Caption: string;
	
	    static createFrom(source: any = {}) {
	        return new Sign(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PlotX = source["PlotX"];
	        this.PlotY = source["PlotY"];
	        this.PlayerType = source["PlayerType"];
	        this.Caption = source["Caption"];
	    }
	}
	export class Team {
	    TeamID: number;
	    Tech: string[];
	    ContactWithTeam: number[];
	    AtWar: number[];
	    PermanentWarPeace: number[];
	    OpenBordersWithTeam: number[];
	    DefensivePactWithTeam: number[];
	    ProjectType: string[];
	    RevealMap: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Team(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TeamID = source["TeamID"];
	        this.Tech = source["Tech"];
	        this.ContactWithTeam = source["ContactWithTeam"];
	        this.AtWar = source["AtWar"];
	        this.PermanentWarPeace = source["PermanentWarPeace"];
	        this.OpenBordersWithTeam = source["OpenBordersWithTeam"];
	        this.DefensivePactWithTeam = source["DefensivePactWithTeam"];
	        this.ProjectType = source["ProjectType"];
	        this.RevealMap = source["RevealMap"];
	    }
	}
	
	export class WorldSizeOption {
	    type: string;
	    description: string;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new WorldSizeOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}

}

