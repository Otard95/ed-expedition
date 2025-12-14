export namespace models {
	
	export class JumpHistoryEntry {
	    // Go type: time
	    timestamp: any;
	    system_name: string;
	    system_id: number;
	    on_route: boolean;
	    expected: boolean;
	    synthetic?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JumpHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.system_name = source["system_name"];
	        this.system_id = source["system_id"];
	        this.on_route = source["on_route"];
	        this.expected = source["expected"];
	        this.synthetic = source["synthetic"];
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
	export class Link {
	    id: string;
	    from: RoutePosition;
	    to: RoutePosition;
	
	    static createFrom(source: any = {}) {
	        return new Link(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.from = this.convertValues(source["from"], RoutePosition);
	        this.to = this.convertValues(source["to"], RoutePosition);
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
	export class RoutePosition {
	    route_id: string;
	    jump_index: number;
	
	    static createFrom(source: any = {}) {
	        return new RoutePosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.route_id = source["route_id"];
	        this.jump_index = source["jump_index"];
	    }
	}
	export class Expedition {
	    id: string;
	    name: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    last_updated: any;
	    status: string;
	    start?: RoutePosition;
	    routes: string[];
	    links: Link[];
	    baked_route_id?: string;
	    current_baked_index: number;
	    baked_loop_back_index?: number;
	    jump_history: JumpHistoryEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Expedition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.last_updated = this.convertValues(source["last_updated"], null);
	        this.status = source["status"];
	        this.start = this.convertValues(source["start"], RoutePosition);
	        this.routes = source["routes"];
	        this.links = this.convertValues(source["links"], Link);
	        this.baked_route_id = source["baked_route_id"];
	        this.current_baked_index = source["current_baked_index"];
	        this.baked_loop_back_index = source["baked_loop_back_index"];
	        this.jump_history = this.convertValues(source["jump_history"], JumpHistoryEntry);
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
	export class ExpeditionSummary {
	    id: string;
	    name: string;
	    status: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    last_updated: any;
	
	    static createFrom(source: any = {}) {
	        return new ExpeditionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.last_updated = this.convertValues(source["last_updated"], null);
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
	
	
	export class Position {
	    x: number;
	    y: number;
	    z: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.z = source["z"];
	    }
	}
	export class RouteJump {
	    system_name: string;
	    system_id: number;
	    scoopable: boolean;
	    distance: number;
	    fuel_in_tank?: number;
	    fuel_used?: number;
	    overcharge?: boolean;
	    position?: Position;
	
	    static createFrom(source: any = {}) {
	        return new RouteJump(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.system_name = source["system_name"];
	        this.system_id = source["system_id"];
	        this.scoopable = source["scoopable"];
	        this.distance = source["distance"];
	        this.fuel_in_tank = source["fuel_in_tank"];
	        this.fuel_used = source["fuel_used"];
	        this.overcharge = source["overcharge"];
	        this.position = this.convertValues(source["position"], Position);
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
	export class Route {
	    id: string;
	    name: string;
	    plotter: string;
	    plotter_parameters: Record<string, any>;
	    plotter_metadata?: Record<string, any>;
	    jumps: RouteJump[];
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Route(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.plotter = source["plotter"];
	        this.plotter_parameters = source["plotter_parameters"];
	        this.plotter_metadata = source["plotter_metadata"];
	        this.jumps = this.convertValues(source["jumps"], RouteJump);
	        this.created_at = this.convertValues(source["created_at"], null);
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

