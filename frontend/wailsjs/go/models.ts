export namespace config {
	
	export class Config {
	    startOnBoot: boolean;
	    minimizeToTray: boolean;
	    showNotifications: boolean;
	    scanReminder: boolean;
	    reminderDays: number;
	    scanNodeModules: boolean;
	    scanGoCache: boolean;
	    scanPythonCache: boolean;
	    scanRustTarget: boolean;
	    scanTempFiles: boolean;
	    scanBrowserCache: boolean;
	    scanIDECache: boolean;
	    scanBuildOutput: boolean;
	    unusedDaysThreshold: number;
	    unusedMinSizeMB: number;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startOnBoot = source["startOnBoot"];
	        this.minimizeToTray = source["minimizeToTray"];
	        this.showNotifications = source["showNotifications"];
	        this.scanReminder = source["scanReminder"];
	        this.reminderDays = source["reminderDays"];
	        this.scanNodeModules = source["scanNodeModules"];
	        this.scanGoCache = source["scanGoCache"];
	        this.scanPythonCache = source["scanPythonCache"];
	        this.scanRustTarget = source["scanRustTarget"];
	        this.scanTempFiles = source["scanTempFiles"];
	        this.scanBrowserCache = source["scanBrowserCache"];
	        this.scanIDECache = source["scanIDECache"];
	        this.scanBuildOutput = source["scanBuildOutput"];
	        this.unusedDaysThreshold = source["unusedDaysThreshold"];
	        this.unusedMinSizeMB = source["unusedMinSizeMB"];
	        this.theme = source["theme"];
	    }
	}

}

export namespace main {
	
	export class CategoryResult {
	    id: string;
	    name: string;
	    description: string;
	    size: number;
	    fileCount: number;
	    color: string;
	    selected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CategoryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.size = source["size"];
	        this.fileCount = source["fileCount"];
	        this.color = source["color"];
	        this.selected = source["selected"];
	    }
	}
	export class CleanResult {
	    cleanedSize: number;
	    cleanedFiles: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new CleanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cleanedSize = source["cleanedSize"];
	        this.cleanedFiles = source["cleanedFiles"];
	        this.errors = source["errors"];
	    }
	}
	export class DiskUsage {
	    total: number;
	    used: number;
	    free: number;
	    usedPercent: number;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.free = source["free"];
	        this.usedPercent = source["usedPercent"];
	        this.path = source["path"];
	    }
	}
	export class FileNode {
	    name: string;
	    path: string;
	    size: number;
	    isDir: boolean;
	    children?: FileNode[];
	    isProtected: boolean;
	    fileCount: number;
	    dirType: string;
	    recommendation: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], FileNode);
	        this.isProtected = source["isProtected"];
	        this.fileCount = source["fileCount"];
	        this.dirType = source["dirType"];
	        this.recommendation = source["recommendation"];
	        this.description = source["description"];
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
	export class ScanResult {
	    categories: CategoryResult[];
	    totalSize: number;
	    totalFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.categories = this.convertValues(source["categories"], CategoryResult);
	        this.totalSize = source["totalSize"];
	        this.totalFiles = source["totalFiles"];
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
	export class UnusedFile {
	    path: string;
	    name: string;
	    size: number;
	    lastAccessed: number;
	    daysUnused: number;
	    fileType: string;
	
	    static createFrom(source: any = {}) {
	        return new UnusedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.lastAccessed = source["lastAccessed"];
	        this.daysUnused = source["daysUnused"];
	        this.fileType = source["fileType"];
	    }
	}

}

