export namespace firewall {
	
	export class FirewallRule {
	    name: string;
	    direction: string;
	    action: string;
	    program: string;
	    localAddr: string;
	    remoteAddr: string;
	    localPort: string;
	    remotePort: string;
	    protocol: string;
	    enabled: boolean;
	    profile: string;
	
	    static createFrom(source: any = {}) {
	        return new FirewallRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.direction = source["direction"];
	        this.action = source["action"];
	        this.program = source["program"];
	        this.localAddr = source["localAddr"];
	        this.remoteAddr = source["remoteAddr"];
	        this.localPort = source["localPort"];
	        this.remotePort = source["remotePort"];
	        this.protocol = source["protocol"];
	        this.enabled = source["enabled"];
	        this.profile = source["profile"];
	    }
	}

}

export namespace ports {
	
	export class ServicePort {
	    name: string;
	    serviceName: string;
	    defaultPort: number;
	    protocol: string;
	    description: string;
	    running: boolean;
	    listenPort: number;
	
	    static createFrom(source: any = {}) {
	        return new ServicePort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.serviceName = source["serviceName"];
	        this.defaultPort = source["defaultPort"];
	        this.protocol = source["protocol"];
	        this.description = source["description"];
	        this.running = source["running"];
	        this.listenPort = source["listenPort"];
	    }
	}

}

export namespace smb {
	
	export class CreateShareRequest {
	    name: string;
	    path: string;
	    description: string;
	    accessMode: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateShareRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.description = source["description"];
	        this.accessMode = source["accessMode"];
	    }
	}
	export class FirewallProfileStatus {
	    name: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FirewallProfileStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	    }
	}
	export class FirewallStatus {
	    available: boolean;
	    profiles: FirewallProfileStatus[];
	
	    static createFrom(source: any = {}) {
	        return new FirewallStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.profiles = this.convertValues(source["profiles"], FirewallProfileStatus);
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
	export class NetworkShare {
	    server: string;
	    name: string;
	    path: string;
	    comment: string;
	    provider: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkShare(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.comment = source["comment"];
	        this.provider = source["provider"];
	    }
	}
	export class Overview {
	    serviceRunning: boolean;
	    serviceStatus: string;
	    startupType: string;
	    smb1Enabled: boolean;
	    smb2Enabled: boolean;
	    requireSecuritySignature: boolean;
	    encryptData: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Overview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceRunning = source["serviceRunning"];
	        this.serviceStatus = source["serviceStatus"];
	        this.startupType = source["startupType"];
	        this.smb1Enabled = source["smb1Enabled"];
	        this.smb2Enabled = source["smb2Enabled"];
	        this.requireSecuritySignature = source["requireSecuritySignature"];
	        this.encryptData = source["encryptData"];
	    }
	}
	export class Session {
	    sessionId: string;
	    clientComputerName: string;
	    clientUserName: string;
	    numOpens: number;
	    secondsExists: number;
	    secondsIdle: number;
	    dialect: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.clientComputerName = source["clientComputerName"];
	        this.clientUserName = source["clientUserName"];
	        this.numOpens = source["numOpens"];
	        this.secondsExists = source["secondsExists"];
	        this.secondsIdle = source["secondsIdle"];
	        this.dialect = source["dialect"];
	    }
	}
	export class Share {
	    name: string;
	    path: string;
	    networkPath: string;
	    description: string;
	    special: boolean;
	    state: string;
	    concurrentUserLimit: number;
	    encryptData: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Share(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.networkPath = source["networkPath"];
	        this.description = source["description"];
	        this.special = source["special"];
	        this.state = source["state"];
	        this.concurrentUserLimit = source["concurrentUserLimit"];
	        this.encryptData = source["encryptData"];
	    }
	}

}

