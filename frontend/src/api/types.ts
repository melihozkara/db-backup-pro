// TypeScript interfaces matching Go structs

export interface DatabaseConnection {
  id: number;
  name: string;
  type: string;
  host: string;
  port: number;
  username: string;
  password: string;
  database_name: string;
  ssl_enabled: boolean;
  auth_database?: string;
  created_at: string;
  updated_at: string;
}

export interface StorageTarget {
  id: number;
  name: string;
  type: string;
  config: string;
  created_at: string;
}

export interface BackupJob {
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
  folder_grouping?: string;
  created_at: string;
}

export interface BackupHistory {
  id: number;
  job_id: number;
  started_at: string;
  completed_at?: string;
  status: string;
  file_name: string;
  file_size: number;
  storage_path: string;
  error_message?: string;
  notification_sent: boolean;
}

export interface HistoryFilter {
  job_id: number;
  status: string;
  limit: number;
}

export interface DashboardStats {
  total_databases: number;
  total_storages: number;
  total_jobs: number;
  active_jobs: number;
  last_24h_total: number;
  last_24h_success: number;
  last_24h_failed: number;
}

export interface TelegramSettings {
  bot_token: string;
  chat_id: string;
  enabled: boolean;
}

export interface ToolPaths {
  pg_dump: string;
  mysqldump: string;
  mongodump: string;
}

export interface AppSettings {
  telegram: TelegramSettings;
  tool_paths: ToolPaths;
  default_retention: number;
  language: string;
}

export interface ServerConfig {
  port: number;
  host: string;
}
