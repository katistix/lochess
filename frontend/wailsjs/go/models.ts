export namespace analysis {
	
	export class MoveGrade {
	    class: string;
	    loss: number;
	    bestUci: string;
	    bestSan: string;
	
	    static createFrom(source: any = {}) {
	        return new MoveGrade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.class = source["class"];
	        this.loss = source["loss"];
	        this.bestUci = source["bestUci"];
	        this.bestSan = source["bestSan"];
	    }
	}
	export class Eval {
	    win: number;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Eval(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.win = source["win"];
	        this.text = source["text"];
	    }
	}
	export class Analysis {
	    evals: Eval[];
	    moves: MoveGrade[];
	    whiteAccuracy: number;
	    blackAccuracy: number;
	    keyMoments: number[];
	
	    static createFrom(source: any = {}) {
	        return new Analysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.evals = this.convertValues(source["evals"], Eval);
	        this.moves = this.convertValues(source["moves"], MoveGrade);
	        this.whiteAccuracy = source["whiteAccuracy"];
	        this.blackAccuracy = source["blackAccuracy"];
	        this.keyMoments = source["keyMoments"];
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

export namespace game {
	
	export class Ply {
	    uci: string;
	    san: string;
	    fen: string;
	    check: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Ply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uci = source["uci"];
	        this.san = source["san"];
	        this.fen = source["fen"];
	        this.check = source["check"];
	    }
	}
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

export namespace history {
	
	export class Game {
	    id: string;
	    // Go type: time
	    playedAt: any;
	    mode: string;
	    level: number;
	    humanColor: string;
	    result: string;
	    method: string;
	    plies: game.Ply[];
	    analysis?: analysis.Analysis;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.playedAt = this.convertValues(source["playedAt"], null);
	        this.mode = source["mode"];
	        this.level = source["level"];
	        this.humanColor = source["humanColor"];
	        this.result = source["result"];
	        this.method = source["method"];
	        this.plies = this.convertValues(source["plies"], game.Ply);
	        this.analysis = this.convertValues(source["analysis"], analysis.Analysis);
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
	export class Summary {
	    id: string;
	    // Go type: time
	    playedAt: any;
	    mode: string;
	    level: number;
	    humanColor: string;
	    result: string;
	    method: string;
	    plies: number;
	    analyzed: boolean;
	    accuracy: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.playedAt = this.convertValues(source["playedAt"], null);
	        this.mode = source["mode"];
	        this.level = source["level"];
	        this.humanColor = source["humanColor"];
	        this.result = source["result"];
	        this.method = source["method"];
	        this.plies = source["plies"];
	        this.analyzed = source["analyzed"];
	        this.accuracy = source["accuracy"];
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
	    gameId: string;
	
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
	        this.gameId = source["gameId"];
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

