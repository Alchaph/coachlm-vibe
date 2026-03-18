export namespace cloudsync {
	
	export class SyncStatus {
	    enabled: boolean;
	    provider: string;
	    lastSyncedAt: string;
	    lastChatSyncAt: string;
	    syncing: boolean;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.provider = source["provider"];
	        this.lastSyncedAt = source["lastSyncedAt"];
	        this.lastChatSyncAt = source["lastChatSyncAt"];
	        this.syncing = source["syncing"];
	        this.lastError = source["lastError"];
	    }
	}

}

export namespace main {
	
	export class ActivityRecord {
	    name: string;
	    activityType: string;
	    startDate: string;
	    distance: number;
	    durationSecs: number;
	    avgPaceSecs: number;
	    avgHR: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivityRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.activityType = source["activityType"];
	        this.startDate = source["startDate"];
	        this.distance = source["distance"];
	        this.durationSecs = source["durationSecs"];
	        this.avgPaceSecs = source["avgPaceSecs"];
	        this.avgHR = source["avgHR"];
	    }
	}
	export class InsightData {
	    id: number;
	    content: string;
	    sourceSessionId: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.sourceSessionId = source["sourceSessionId"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class ProfileData {
	    age: number;
	    maxHR: number;
	    thresholdPaceSecs: number;
	    weeklyMileageTarget: number;
	    raceGoals: string;
	    injuryHistory: string;
	    experienceLevel: string;
	    trainingDaysPerWeek: number;
	    restingHR: number;
	    preferredTerrain: string;
	    heartRateZones: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.age = source["age"];
	        this.maxHR = source["maxHR"];
	        this.thresholdPaceSecs = source["thresholdPaceSecs"];
	        this.weeklyMileageTarget = source["weeklyMileageTarget"];
	        this.raceGoals = source["raceGoals"];
	        this.injuryHistory = source["injuryHistory"];
	        this.experienceLevel = source["experienceLevel"];
	        this.trainingDaysPerWeek = source["trainingDaysPerWeek"];
	        this.restingHR = source["restingHR"];
	        this.preferredTerrain = source["preferredTerrain"];
	        this.heartRateZones = source["heartRateZones"];
	    }
	}
	export class SettingsData {
	    ollamaEndpoint: string;
	    ollamaModel: string;
	    customSystemPrompt: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ollamaEndpoint = source["ollamaEndpoint"];
	        this.ollamaModel = source["ollamaModel"];
	        this.customSystemPrompt = source["customSystemPrompt"];
	    }
	}
	export class StatsData {
	    totalCount: number;
	    totalDistanceKm: number;
	    earliestDate: string;
	    latestDate: string;
	
	    static createFrom(source: any = {}) {
	        return new StatsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCount = source["totalCount"];
	        this.totalDistanceKm = source["totalDistanceKm"];
	        this.earliestDate = source["earliestDate"];
	        this.latestDate = source["latestDate"];
	    }
	}

}

export namespace plan {
	
	export class ActualMetrics {
	    durationMin?: number;
	    distanceKm?: number;
	
	    static createFrom(source: any = {}) {
	        return new ActualMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.durationMin = source["durationMin"];
	        this.distanceKm = source["distanceKm"];
	    }
	}
	export class Race {
	    id: string;
	    name: string;
	    distanceKm: number;
	    // Go type: time
	    raceDate: any;
	    terrain: string;
	    elevationM?: number;
	    goalTimeSec?: number;
	    priority: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Race(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.distanceKm = source["distanceKm"];
	        this.raceDate = this.convertValues(source["raceDate"], null);
	        this.terrain = source["terrain"];
	        this.elevationM = source["elevationM"];
	        this.goalTimeSec = source["goalTimeSec"];
	        this.priority = source["priority"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class Session {
	    id: string;
	    weekId: string;
	    dayOfWeek: number;
	    type: string;
	    durationMin: number;
	    distanceKm?: number;
	    hrZone?: number;
	    paceMinLow?: number;
	    paceMinHigh?: number;
	    notes: string;
	    status: string;
	    actualDurationMin?: number;
	    actualDistanceKm?: number;
	    // Go type: time
	    completedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.weekId = source["weekId"];
	        this.dayOfWeek = source["dayOfWeek"];
	        this.type = source["type"];
	        this.durationMin = source["durationMin"];
	        this.distanceKm = source["distanceKm"];
	        this.hrZone = source["hrZone"];
	        this.paceMinLow = source["paceMinLow"];
	        this.paceMinHigh = source["paceMinHigh"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.actualDurationMin = source["actualDurationMin"];
	        this.actualDistanceKm = source["actualDistanceKm"];
	        this.completedAt = this.convertValues(source["completedAt"], null);
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
	export class Week {
	    id: string;
	    planId: string;
	    weekNumber: number;
	    // Go type: time
	    weekStart: any;
	    sessions: Session[];
	
	    static createFrom(source: any = {}) {
	        return new Week(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.planId = source["planId"];
	        this.weekNumber = source["weekNumber"];
	        this.weekStart = this.convertValues(source["weekStart"], null);
	        this.sessions = this.convertValues(source["sessions"], Session);
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
	export class TrainingPlan {
	    id: string;
	    raceId: string;
	    // Go type: time
	    generatedAt: any;
	    llmBackend: string;
	    promptHash: string;
	    weeks?: Week[];
	
	    static createFrom(source: any = {}) {
	        return new TrainingPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.raceId = source["raceId"];
	        this.generatedAt = this.convertValues(source["generatedAt"], null);
	        this.llmBackend = source["llmBackend"];
	        this.promptHash = source["promptHash"];
	        this.weeks = this.convertValues(source["weeks"], Week);
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

