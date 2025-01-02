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

}

