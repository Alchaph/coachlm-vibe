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

	export class SettingsData {
	    claudeApiKey: string;
	    openaiApiKey: string;
	    activeLlm: string;
	    ollamaEndpoint: string;
	    stravaClientId: string;
	    stravaClientSecret: string;
	
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
	    }
	}

}
