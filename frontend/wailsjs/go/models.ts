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

