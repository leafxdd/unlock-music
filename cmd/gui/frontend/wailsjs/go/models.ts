export namespace main {
	
	export class DropTarget {
	    dir: string;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DropTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.isDir = source["isDir"];
	    }
	}
	export class Settings {
	    inputDir: string;
	    outputDir: string;
	    skipNoop: boolean;
	    removeSource: boolean;
	    updateMetadata: boolean;
	    overwriteOutput: boolean;
	    qmcMmkvPath: string;
	    qmcMmkvKey: string;
	    kggDbPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputDir = source["inputDir"];
	        this.outputDir = source["outputDir"];
	        this.skipNoop = source["skipNoop"];
	        this.removeSource = source["removeSource"];
	        this.updateMetadata = source["updateMetadata"];
	        this.overwriteOutput = source["overwriteOutput"];
	        this.qmcMmkvPath = source["qmcMmkvPath"];
	        this.qmcMmkvKey = source["qmcMmkvKey"];
	        this.kggDbPath = source["kggDbPath"];
	    }
	}

}

