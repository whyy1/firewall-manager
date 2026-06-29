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

