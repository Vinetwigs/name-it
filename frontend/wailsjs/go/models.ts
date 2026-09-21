export namespace config {
	
	export class Config {
	    aggressiveness: string;
	    defaultReminderMinutes: number;
	    minimumReminderMinutes: number;
	    ignoredExtensions: string[];
	    ignoredDirectories: string[];
	    suspiciousWords: string[];
	    notificationsEnabled: boolean;
	    namingMode: string;
	    customPattern: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aggressiveness = source["aggressiveness"];
	        this.defaultReminderMinutes = source["defaultReminderMinutes"];
	        this.minimumReminderMinutes = source["minimumReminderMinutes"];
	        this.ignoredExtensions = source["ignoredExtensions"];
	        this.ignoredDirectories = source["ignoredDirectories"];
	        this.suspiciousWords = source["suspiciousWords"];
	        this.notificationsEnabled = source["notificationsEnabled"];
	        this.namingMode = source["namingMode"];
	        this.customPattern = source["customPattern"];
	    }
	}

}

export namespace history {
	
	export class Crime {
	    // Go type: time
	    timestamp: any;
	    filename: string;
	    score: number;
	    severity: string;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new Crime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.filename = source["filename"];
	        this.score = source["score"];
	        this.severity = source["severity"];
	        this.action = source["action"];
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

export namespace session {
	
	export class State {
	    active: boolean;
	    root: string;
	    filesWatched: number;
	    questionable: number;
	    // Go type: time
	    pausedUntil: any;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active = source["active"];
	        this.root = source["root"];
	        this.filesWatched = source["filesWatched"];
	        this.questionable = source["questionable"];
	        this.pausedUntil = this.convertValues(source["pausedUntil"], null);
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

