export namespace main {
	
	export class OutfitDetails {
	    figure: string;
	
	    static createFrom(source: any = {}) {
	        return new OutfitDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.figure = source["figure"];
	    }
	}
	export class OutfitConfig {
	    outfits: {[key: string]: OutfitDetails};
	
	    static createFrom(source: any = {}) {
	        return new OutfitConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outfits = this.convertValues(source["outfits"], OutfitDetails, true);
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

