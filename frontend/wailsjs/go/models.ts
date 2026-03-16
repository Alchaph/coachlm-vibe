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
	    }
	}
	export class SettingsData {
	    claudeApiKey: string;
	    openaiApiKey: string;
	    activeLlm: string;
	    ollamaEndpoint: string;
	    stravaClientId: string;
	    stravaClientSecret: string;
	    claudeModel: string;
	    openaiModel: string;
	    ollamaModel: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.claudeApiKey = source["claudeApiKey"];
	        this.openaiApiKey = source["openaiApiKey"];
	        this.activeLlm = source["activeLlm"];
	        this.ollamaEndpoint = source["ollamaEndpoint"];
	        this.stravaClientId = source["stravaClientId"];
	        this.stravaClientSecret = source["stravaClientSecret"];
	        this.claudeModel = source["claudeModel"];
	        this.openaiModel = source["openaiModel"];
	        this.ollamaModel = source["ollamaModel"];
	    }
	}

}

