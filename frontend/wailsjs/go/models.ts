export namespace game {
	
	export class State {
	    fen: string;
	    turn: string;
	    dests: Record<string, Array<string>>;
	    lastMove: string[];
	    check: boolean;
	    outcome: string;
	    method: string;
	
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
	    }
	}

}

