export namespace classifier {
	
	export class Category {
	    key: string;
	    name: string;
	    icon: string;
	    known: string;
	    cluster: string;
	    exts: string[];
	
	    static createFrom(source: any = {}) {
	        return new Category(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.known = source["known"];
	        this.cluster = source["cluster"];
	        this.exts = source["exts"];
	    }
	}

}

export namespace config {
	
	export class CustomRule {
	    ext: string;
	    target: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ext = source["ext"];
	        this.target = source["target"];
	    }
	}
	export class Config {
	    sources: string[];
	    customRules: CustomRule[];
	    times: string[];
	    enabled: boolean;
	    scheduleTime: string;
	    instant: boolean;
	    instantDelay: number;
	    theme: string;
	    minCluster: number;
	    minAgeSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sources = source["sources"];
	        this.customRules = this.convertValues(source["customRules"], CustomRule);
	        this.times = source["times"];
	        this.enabled = source["enabled"];
	        this.scheduleTime = source["scheduleTime"];
	        this.instant = source["instant"];
	        this.instantDelay = source["instantDelay"];
	        this.theme = source["theme"];
	        this.minCluster = source["minCluster"];
	        this.minAgeSeconds = source["minAgeSeconds"];
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

export namespace knownfolders {
	
	export class Folder {
	    key: string;
	    name: string;
	    path: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new Folder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.icon = source["icon"];
	    }
	}

}

export namespace main {
	
	export class About {
	    name: string;
	    version: string;
	    author: string;
	    configPath: string;
	    exePath: string;
	
	    static createFrom(source: any = {}) {
	        return new About(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.configPath = source["configPath"];
	        this.exePath = source["exePath"];
	    }
	}
	export class Dashboard {
	    sources: string[];
	    knownFolders: knownfolders.Folder[];
	    categories: classifier.Category[];
	    schedule: scheduler.Status;
	    lastRun?: organizer.RunRecord;
	    totalBytes: number;
	    runCount: number;
	    configPath: string;
	    downloadsCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sources = source["sources"];
	        this.knownFolders = this.convertValues(source["knownFolders"], knownfolders.Folder);
	        this.categories = this.convertValues(source["categories"], classifier.Category);
	        this.schedule = this.convertValues(source["schedule"], scheduler.Status);
	        this.lastRun = this.convertValues(source["lastRun"], organizer.RunRecord);
	        this.totalBytes = source["totalBytes"];
	        this.runCount = source["runCount"];
	        this.configPath = source["configPath"];
	        this.downloadsCount = source["downloadsCount"];
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

export namespace organizer {
	
	export class GroupStat {
	    category: string;
	    icon: string;
	    folder: string;
	    count: number;
	    bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new GroupStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.icon = source["icon"];
	        this.folder = source["folder"];
	        this.count = source["count"];
	        this.bytes = source["bytes"];
	    }
	}
	export class Move {
	    from: string;
	    to: string;
	    size: number;
	    category: string;
	    knownKey: string;
	
	    static createFrom(source: any = {}) {
	        return new Move(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.size = source["size"];
	        this.category = source["category"];
	        this.knownKey = source["knownKey"];
	    }
	}
	export class PlannedMove {
	    from: string;
	    to: string;
	    name: string;
	    size: number;
	    modified: number;
	    category: string;
	    icon: string;
	    knownKey: string;
	    kind: string;
	    err: string;
	
	    static createFrom(source: any = {}) {
	        return new PlannedMove(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	        this.category = source["category"];
	        this.icon = source["icon"];
	        this.knownKey = source["knownKey"];
	        this.kind = source["kind"];
	        this.err = source["err"];
	    }
	}
	export class Report {
	    source: string;
	    apply: boolean;
	    scanned: number;
	    skipped: number;
	    skippedYoung: number;
	    moved: number;
	    moves: PlannedMove[];
	    bytes: number;
	    groups: GroupStat[];
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.apply = source["apply"];
	        this.scanned = source["scanned"];
	        this.skipped = source["skipped"];
	        this.skippedYoung = source["skippedYoung"];
	        this.moved = source["moved"];
	        this.moves = this.convertValues(source["moves"], PlannedMove);
	        this.bytes = source["bytes"];
	        this.groups = this.convertValues(source["groups"], GroupStat);
	        this.errors = source["errors"];
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
	export class RunRecord {
	    id: string;
	    runId?: string;
	    // Go type: time
	    timestamp: any;
	    source: string;
	    moves: Move[];
	    bytes: number;
	    undone: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.runId = source["runId"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.source = source["source"];
	        this.moves = this.convertValues(source["moves"], Move);
	        this.bytes = source["bytes"];
	        this.undone = source["undone"];
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
	export class UndoResult {
	    runId: string;
	    records: number;
	    moves: number;
	    bytes: number;
	    missing: number;
	
	    static createFrom(source: any = {}) {
	        return new UndoResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.records = source["records"];
	        this.moves = source["moves"];
	        this.bytes = source["bytes"];
	        this.missing = source["missing"];
	    }
	}

}

export namespace scheduler {
	
	export class Status {
	    enabled: boolean;
	    times: string[];
	    registered: string[];
	    instant: boolean;
	    watchActive: boolean;
	    isAdmin: boolean;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.times = source["times"];
	        this.registered = source["registered"];
	        this.instant = source["instant"];
	        this.watchActive = source["watchActive"];
	        this.isAdmin = source["isAdmin"];
	        this.detail = source["detail"];
	    }
	}

}

