// Dual-mode API client: Wails bindings in desktop mode, HTTP fetch in web mode

import { isWailsMode } from './mode';
import { http } from './http';
import type {
  DatabaseConnection,
  StorageTarget,
  BackupJob,
  BackupHistory,
  HistoryFilter,
  DashboardStats,
  AppSettings,
  ServerConfig,
} from './types';

// ==================== DATABASE CONNECTIONS ====================

export async function GetDatabases(): Promise<DatabaseConnection[]> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetDatabases();
  return http.get('/api/databases');
}

export async function GetDatabase(id: number): Promise<DatabaseConnection> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetDatabase(id);
  return http.get(`/api/databases/${id}`);
}

export async function AddDatabase(db: DatabaseConnection): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).AddDatabase(db as any);
  await http.post('/api/databases', db);
}

export async function UpdateDatabase(db: DatabaseConnection): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).UpdateDatabase(db as any);
  await http.put(`/api/databases/${db.id}`, db);
}

export async function DeleteDatabase(id: number): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).DeleteDatabase(id);
  await http.del(`/api/databases/${id}`);
}

export async function TestDatabaseConnection(db: DatabaseConnection): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).TestDatabaseConnection(db as any);
  await http.post('/api/databases/test', db);
}

// ==================== STORAGE TARGETS ====================

export async function GetStorageTargets(): Promise<StorageTarget[]> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetStorageTargets();
  return http.get('/api/storage-targets');
}

export async function GetStorageTarget(id: number): Promise<StorageTarget> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetStorageTarget(id);
  return http.get(`/api/storage-targets/${id}`);
}

export async function AddStorageTarget(st: StorageTarget): Promise<number> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).AddStorageTarget(st as any);
  const res = await http.post<{ id: number }>('/api/storage-targets', st);
  return res.id;
}

export async function UpdateStorageTarget(st: StorageTarget): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).UpdateStorageTarget(st as any);
  await http.put(`/api/storage-targets/${st.id}`, st);
}

export async function DeleteStorageTarget(id: number): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).DeleteStorageTarget(id);
  await http.del(`/api/storage-targets/${id}`);
}

export async function TestStorageConnection(st: StorageTarget): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).TestStorageConnection(st as any);
  await http.post('/api/storage-targets/test', st);
}

// ==================== BACKUP JOBS ====================

export async function GetBackupJobs(): Promise<BackupJob[]> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetBackupJobs();
  return http.get('/api/jobs');
}

export async function GetBackupJob(id: number): Promise<BackupJob> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetBackupJob(id);
  return http.get(`/api/jobs/${id}`);
}

export async function AddBackupJob(job: BackupJob): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).AddBackupJob(job as any);
  await http.post('/api/jobs', job);
}

export async function UpdateBackupJob(job: BackupJob): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).UpdateBackupJob(job as any);
  await http.put(`/api/jobs/${job.id}`, job);
}

export async function DeleteBackupJob(id: number): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).DeleteBackupJob(id);
  await http.del(`/api/jobs/${id}`);
}

export async function ToggleBackupJobActive(id: number, active: boolean): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).ToggleBackupJobActive(id, active);
  await http.put(`/api/jobs/${id}/toggle`, { active });
}

export async function RunBackupNow(jobID: number): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).RunBackupNow(jobID);
  await http.post(`/api/jobs/${jobID}/run`);
}

// ==================== HISTORY & DASHBOARD ====================

export async function GetBackupHistory(filter: HistoryFilter): Promise<BackupHistory[]> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetBackupHistory(filter as any);
  const params = new URLSearchParams();
  if (filter.job_id) params.set('job_id', String(filter.job_id));
  if (filter.status) params.set('status', filter.status);
  if (filter.limit) params.set('limit', String(filter.limit));
  return http.get(`/api/history?${params}`);
}

export async function GetDashboardStats(): Promise<DashboardStats> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetDashboardStats();
  return http.get('/api/dashboard/stats');
}

// ==================== SETTINGS ====================

export async function GetSettings(): Promise<AppSettings> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetSettings();
  return http.get('/api/settings');
}

export async function SaveSettings(settings: AppSettings): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).SaveSettings(settings as any);
  await http.put('/api/settings', settings);
}

export async function SetAppLanguage(lang: string): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).SetAppLanguage(lang);
  await http.put('/api/language', { language: lang });
}

export async function GetTranslations(lang: string): Promise<Record<string, any>> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).GetTranslations(lang);
  return http.get(`/api/translations/${lang}`);
}

// ==================== TELEGRAM ====================

export async function TestTelegramConnection(botToken: string): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).TestTelegramConnection(botToken);
  await http.post('/api/telegram/test-token', { bot_token: botToken });
}

export async function SendTestTelegramMessage(botToken: string, chatID: string): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).SendTestTelegramMessage(botToken, chatID);
  await http.post('/api/telegram/test-message', { bot_token: botToken, chat_id: chatID });
}

// ==================== UTILITY ====================

export async function SelectFolder(): Promise<string> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).SelectFolder();
  // Web mode doesn't support native folder dialogs
  throw new Error('Folder selection not available in web mode');
}

export async function ValidateFolder(path: string): Promise<void> {
  if (isWailsMode()) return (await import('../../wailsjs/go/main/App')).ValidateFolder(path);
  await http.post('/api/validate-folder', { path });
}

export async function GetServerConfig(): Promise<ServerConfig> {
  if (isWailsMode()) {
    // In wails mode, load from Go binding
    const mod = await import('../../wailsjs/go/main/App');
    if ('GetServerConfig' in mod) return (mod as any).GetServerConfig();
    return { port: 8090, host: '127.0.0.1' };
  }
  return http.get('/api/server-config');
}

export async function SaveServerConfig(cfg: ServerConfig): Promise<void> {
  if (isWailsMode()) {
    const mod = await import('../../wailsjs/go/main/App');
    if ('SaveServerConfig' in mod) return (mod as any).SaveServerConfig(cfg);
    return;
  }
  await http.put('/api/server-config', cfg);
}
