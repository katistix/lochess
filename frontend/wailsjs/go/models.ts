export namespace game {
	
	export class State {
	    fen: string;
	    turn: string;
	    dests: Record<string, Array<string>>;
	    lastMove: string[];
	    check: boolean;
	    outcome: string;
	    method: string;
	    moves: string[];
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fen = source["fen"];
	        this.turn = source["turn"];
	        this.dests = source["dests"];
	        this.lastMove = source["lastMove"];
	        this.check = source["check"];
	        this.outcome = source["outcome"];
	        this.method = source["method"];
	        this.moves = source["moves"];
	    }
	}

}

export namespace match {
	
	export class Options {
	    mode: string;
	    level: number;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.level = source["level"];
	        this.color = source["color"];
	    }
	}
	export class State {
	    game: game.State;
	    phase: string;
	    mode: string;
	    level: number;
	    humanColor: string;
	    botToMove: boolean;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game = this.convertValues(source["game"], game.State);
	        this.phase = source["phase"];
	        this.mode = source["mode"];
	        this.level = source["level"];
	        this.humanColor = source["humanColor"];
	        this.botToMove = source["botToMove"];
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

}

