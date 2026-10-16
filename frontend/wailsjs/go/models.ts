export namespace service {
	
	export class FileResult {
	    source: string;
	    destination: string;
	    size: number;
	    status: string;
	    error: string;
	    copiedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.destination = source["destination"];
	        this.size = source["size"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.copiedAt = source["copiedAt"];
	    }
	}
	export class Result {
	    total: number;
	    copied: number;
	    skipped: number;
	    failed: number;
	    files: FileResult[];
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.copied = source["copied"];
	        this.skipped = source["skipped"];
	        this.failed = source["failed"];
	        this.files = this.convertValues(source["files"], FileResult);
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

