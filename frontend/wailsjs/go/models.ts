export namespace config {
	
	export class ServerConfig {
	    port: number;
	    host: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.host = source["host"];
	    }
	}

}

export namespace database {
	
	export class ToolPaths {
	    pg_dump: string;
	    mysqldump: string;
	    mongodump: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolPaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pg_dump = source["pg_dump"];
	        this.mysqldump = source["mysqldump"];
	        this.mongodump = source["mongodump"];
	    }
	}
	export class TelegramSettings {
	    bot_token: string;
	    chat_id: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TelegramSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bot_token = source["bot_token"];
	        this.chat_id = source["chat_id"];
	        this.enabled = source["enabled"];
	    }
	}
	export class AppSettings {
	    telegram: TelegramSettings;
	    tool_paths: ToolPaths;
	    default_retention: number;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.telegram = this.convertValues(source["telegram"], TelegramSettings);
	        this.tool_paths = this.convertValues(source["tool_paths"], ToolPaths);
	        this.default_retention = source["default_retention"];
	        this.language = source["language"];
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
	export class BackupHistory {
	    id: number;
	    job_id: number;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    completed_at?: any;
	    status: string;
	    file_name: string;
	    file_size: number;
	    storage_path: string;
	    error_message?: string;
	    notification_sent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BackupHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.job_id = source["job_id"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.status = source["status"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.storage_path = source["storage_path"];
	        this.error_message = source["error_message"];
	        this.notification_sent = source["notification_sent"];
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
	export class BackupJob {
	    id: number;
	    name: string;
	    database_id: number;
	    storage_id: number;
	    schedule_type: string;
	    schedule_config: string;
	    compression: string;
	    encryption: boolean;
	    encryption_key?: string;
	    retention_days: number;
	    is_active: boolean;
	    custom_prefix: string;
	    custom_folder: string;
	    folder_grouping: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new BackupJob(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.database_id = source["database_id"];
	        this.storage_id = source["storage_id"];
	        this.schedule_type = source["schedule_type"];
	        this.schedule_config = source["schedule_config"];
	        this.compression = source["compression"];
	        this.encryption = source["encryption"];
	        this.encryption_key = source["encryption_key"];
	        this.retention_days = source["retention_days"];
	        this.is_active = source["is_active"];
	        this.custom_prefix = source["custom_prefix"];
	        this.custom_folder = source["custom_folder"];
	        this.folder_grouping = source["folder_grouping"];
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
	export class DashboardStats {
	    total_databases: number;
	    total_storages: number;
	    total_jobs: number;
	    active_jobs: number;
	    last_24h_total: number;
	    last_24h_success: number;
	    last_24h_failed: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_databases = source["total_databases"];
	        this.total_storages = source["total_storages"];
	        this.total_jobs = source["total_jobs"];
	        this.active_jobs = source["active_jobs"];
	        this.last_24h_total = source["last_24h_total"];
	        this.last_24h_success = source["last_24h_success"];
	        this.last_24h_failed = source["last_24h_failed"];
	    }
	}
	export class DatabaseConnection {
	    id: number;
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    database_name: string;
	    ssl_enabled: boolean;
	    auth_database: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.database_name = source["database_name"];
	        this.ssl_enabled = source["ssl_enabled"];
	        this.auth_database = source["auth_database"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class HistoryFilter {
	    job_id: number;
	    status: string;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.job_id = source["job_id"];
	        this.status = source["status"];
	        this.limit = source["limit"];
	    }
	}
	export class StorageTarget {
	    id: number;
	    name: string;
	    type: string;
	    config: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new StorageTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.config = source["config"];
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

