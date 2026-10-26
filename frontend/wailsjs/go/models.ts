export namespace database {
	
	export class CleanupResult {
	    relinked: number;
	    removed: number;
	
	    static createFrom(source: any = {}) {
	        return new CleanupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relinked = source["relinked"];
	        this.removed = source["removed"];
	    }
	}
	export class Photo {
	    id: number;
	    fileName: string;
	    sourcePath: string;
	    destinationPath: string;
	    fileSize: number;
	    fileHash: string;
	    backupStatus: string;
	    // Go type: time
	    backedUpAt: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Photo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileName = source["fileName"];
	        this.sourcePath = source["sourcePath"];
	        this.destinationPath = source["destinationPath"];
	        this.fileSize = source["fileSize"];
	        this.fileHash = source["fileHash"];
	        this.backupStatus = source["backupStatus"];
	        this.backedUpAt = this.convertValues(source["backedUpAt"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class Tag {
	    id: number;
	    name: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
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
	export class TagCloudItem {
	    id: number;
	    name: string;
	    photoCount: number;
	
	    static createFrom(source: any = {}) {
	        return new TagCloudItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.photoCount = source["photoCount"];
	    }
	}

}

export namespace service {
	
	export class FileResult {
	    source: string;
	    destination: string;
	    size: number;
	    hash: string;
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
	        this.hash = source["hash"];
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

