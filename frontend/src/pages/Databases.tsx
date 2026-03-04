import { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  GetDatabases, AddDatabase, UpdateDatabase, DeleteDatabase, TestDatabaseConnection,
  GetStorageTargets, AddStorageTarget, TestStorageConnection, SelectFolder,
  GetBackupJobs, AddBackupJob, UpdateBackupJob, DeleteBackupJob, ToggleBackupJobActive, RunBackupNow,
} from '../api/client';
import { EventsOn, EventsOff } from '../api/events';
import { isWebMode } from '../api/mode';
import { useToast } from '../components/Toast';
import { useConfirm } from '../components/ConfirmDialog';
import { PageLoader } from '../components/Spinner';
import PageHeader from '../components/PageHeader';
import { Card, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select } from '../components/ui/select';
import { Badge } from '../components/ui/badge';
import { Switch } from '../components/ui/switch';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../components/ui/dialog';
import {
  Plus, Database, Pencil, Trash2, Plug, Loader2, Play, HardDrive, Globe, Lock, Cloud, FolderOpen, Target,
} from 'lucide-react';

// ==================== CONSTANTS ====================

const dbTypes = [
  { value: 'postgres', label: 'PostgreSQL', port: 5432, color: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/20' },
  { value: 'mysql', label: 'MySQL', port: 3306, color: 'text-orange-400', bg: 'bg-orange-500/10', border: 'border-orange-500/20' },
  { value: 'mongodb', label: 'MongoDB', port: 27017, color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20' },
];

const storageTypesMeta = [
  { value: 'local', icon: HardDrive, color: 'text-zinc-400', bg: 'bg-zinc-500/10' },
  { value: 'ftp', icon: Globe, color: 'text-blue-400', bg: 'bg-blue-500/10' },
  { value: 'sftp', icon: Lock, color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
  { value: 's3', icon: Cloud, color: 'text-amber-400', bg: 'bg-amber-500/10' },
];

const now = () => new Date().toISOString();

const emptyDbForm: any = {
  name: '', type: 'postgres', host: 'localhost', port: 5432,
  username: '', password: '', database_name: '', ssl_enabled: false,
  auth_database: '', created_at: now(), updated_at: now(),
};

// Parse database connection URI into form fields
const parseConnectionURI = (uri: string): Partial<any> | null => {
  try {
    // MongoDB: mongodb://user:pass@host:port/db?authSource=admin
    // MongoDB+SRV: mongodb+srv://user:pass@host/db?authSource=admin
    // PostgreSQL: postgresql://user:pass@host:port/db?sslmode=require
    // MySQL: mysql://user:pass@host:port/db

    const mongoMatch = uri.match(/^mongodb(?:\+srv)?:\/\//);
    const pgMatch = uri.match(/^(?:postgresql|postgres):\/\//);
    const mysqlMatch = uri.match(/^mysql:\/\//);

    if (!mongoMatch && !pgMatch && !mysqlMatch) return null;

    let type = 'postgres';
    let defaultPort = 5432;
    const isSrv = uri.startsWith('mongodb+srv://');

    if (mongoMatch) { type = 'mongodb'; defaultPort = 27017; }
    else if (mysqlMatch) { type = 'mysql'; defaultPort = 3306; }

    // Remove scheme
    const withoutScheme = uri.replace(/^(?:mongodb(?:\+srv)?|postgresql|postgres|mysql):\/\//, '');

    // Split query params
    const [mainPart, queryString] = withoutScheme.split('?');

    // Parse auth@host/db
    let userInfo = '';
    let hostPart = mainPart;

    if (mainPart.includes('@')) {
      const atIdx = mainPart.lastIndexOf('@');
      userInfo = mainPart.substring(0, atIdx);
      hostPart = mainPart.substring(atIdx + 1);
    }

    // Parse username:password
    let username = '';
    let password = '';
    if (userInfo) {
      const colonIdx = userInfo.indexOf(':');
      if (colonIdx >= 0) {
        username = decodeURIComponent(userInfo.substring(0, colonIdx));
        password = decodeURIComponent(userInfo.substring(colonIdx + 1));
      } else {
        username = decodeURIComponent(userInfo);
      }
    }

    // Parse host:port/database
    const [hostPortPart, ...dbParts] = hostPart.split('/');
    const databaseName = dbParts.join('/');

    let host = hostPortPart;
    let port = defaultPort;
    const portMatch = hostPortPart.match(/^(.+):(\d+)$/);
    if (portMatch) {
      host = portMatch[1];
      port = parseInt(portMatch[2]);
    }

    // Parse query params
    let authDatabase = '';
    let sslEnabled = isSrv;
    if (queryString) {
      const params = new URLSearchParams(queryString);
      authDatabase = params.get('authSource') || '';
      if (params.get('ssl') === 'true' || params.get('tls') === 'true' || params.get('sslmode') === 'require') {
        sslEnabled = true;
      }
    }

    return {
      type, host, port, username, password,
      database_name: databaseName,
      ssl_enabled: sslEnabled,
      auth_database: authDatabase,
    };
  } catch {
    return null;
  }
};

// ==================== HELPERS ====================

const getDbStyle = (type: string) => dbTypes.find(t => t.value === type) || dbTypes[0];
const getStorageStyle = (type: string) => storageTypesMeta.find(t => t.value === type) || storageTypesMeta[0];

const getStorageDescription = (st: any) => {
  try {
    const cfg = JSON.parse(st.config || '{}');
    switch (st.type) {
      case 'local': return cfg.path || '-';
      case 's3': return `${cfg.bucket}${cfg.path || ''}`;
      default: return `${cfg.host}:${cfg.port}${cfg.path || ''}`;
    }
  } catch {
    return '-';
  }
};

// ==================== COMPONENT ====================

export default function Databases() {
  const { t } = useTranslation();
  const [databases, setDatabases] = useState<any[]>([]);
  const [storages, setStorages] = useState<any[]>([]);
  const [jobs, setJobs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const toast = useToast();
  const { confirm } = useConfirm();

  // DB Modal state
  const [showDbModal, setShowDbModal] = useState(false);
  const [editingDbId, setEditingDbId] = useState<number | null>(null);
  const [dbForm, setDbForm] = useState<any>(emptyDbForm);
  const [testingDb, setTestingDb] = useState(false);
  const [savingDb, setSavingDb] = useState(false);

  // Job (Backup Target) Modal state
  const [showJobModal, setShowJobModal] = useState(false);
  const [editingJobId, setEditingJobId] = useState<number | null>(null);
  const [jobDatabaseId, setJobDatabaseId] = useState<number>(0);
  const [storageMode, setStorageMode] = useState<'existing' | 'new'>('existing');
  const [selectedStorageId, setSelectedStorageId] = useState<number>(0);
  const [newStorageForm, setNewStorageForm] = useState({ name: '', type: 'local' });
  const [newStorageConfig, setNewStorageConfig] = useState<any>({ path: '' });
  const [testingStorage, setTestingStorage] = useState(false);
  const [jobForm, setJobForm] = useState<any>({
    schedule_type: 'manual', compression: 'gzip', retention_days: 7,
    encryption: false, encryption_key: '', is_active: true,
    custom_prefix: '', custom_folder: '', folder_grouping: '',
  });
  const [scheduleConfig, setScheduleConfig] = useState<any>({});
  const [savingJob, setSavingJob] = useState(false);

  // Running backup state: jobId -> stage (starting, dumping, processing, uploading)
  const [runningJobs, setRunningJobs] = useState<Record<number, string>>({});

  // Translated constants
  const storageTypes = storageTypesMeta.map(s => ({
    ...s,
    label: t(`storage.types.${s.value}`),
  }));
  const scheduleTypes = [
    { value: 'manual', label: t('schedule.types.manual') },
    { value: 'interval', label: t('schedule.types.interval') },
    { value: 'daily', label: t('schedule.types.daily') },
    { value: 'weekly', label: t('schedule.types.weekly') },
  ];
  const compressionTypes = [
    { value: 'none', label: t('options.compressionNone') },
    { value: 'gzip', label: 'GZIP' },
    { value: 'zip', label: 'ZIP' },
  ];
  const weekdays = Array.from({ length: 7 }, (_, i) => t(`schedule.weekdays.${i}`));

  const stageLabels: Record<string, string> = {
    starting: t('backup.stages.starting'),
    dumping: t('backup.stages.dumping'),
    processing: t('backup.stages.processing'),
    uploading: t('backup.stages.uploading'),
  };

  const getScheduleLabel = (job: any) => {
    try {
      const cfg = JSON.parse(job.schedule_config || '{}');
      switch (job.schedule_type) {
        case 'manual': return t('schedule.types.manual');
        case 'interval': return t('schedule.everyNMinutes', { n: cfg.interval_minutes });
        case 'daily': return t('schedule.dailyAt', { time: `${String(cfg.hour || 0).padStart(2, '0')}:${String(cfg.minute || 0).padStart(2, '0')}` });
        case 'weekly': {
          const days = (cfg.weekdays || []).map((d: number) => weekdays[d]?.slice(0, 3)).join(', ');
          return t('schedule.weeklyAt', { days, time: `${String(cfg.hour || 0).padStart(2, '0')}:${String(cfg.minute || 0).padStart(2, '0')}` });
        }
        default: return '-';
      }
    } catch {
      return '-';
    }
  };

  const getCompressionLabel = (v: string) => compressionTypes.find(c => c.value === v)?.label || v;

  // ==================== DATA LOADING ====================

  const loadData = useCallback(async () => {
    try {
      const [dbsData, storagesData, jobsData] = await Promise.all([
        GetDatabases(), GetStorageTargets(), GetBackupJobs(),
      ]);
      setDatabases(dbsData || []);
      setStorages(storagesData || []);
      setJobs(jobsData || []);
    } catch {
      toast.error(t('common.loadingError'), t('databases.dataLoadFailed'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => { loadData(); }, [loadData]);

  // ==================== REAL-TIME BACKUP EVENTS ====================

  useEffect(() => {
    const onStarted = (data: any) => {
      const id = data?.job_id;
      if (id) setRunningJobs(prev => ({ ...prev, [id]: 'starting' }));
      toast.info(t('backup.started'), t('backup.jobBackingUp', { name: data?.job_name || t('history.job') }));
    };
    const onProgress = (data: any) => {
      const id = data?.job_id;
      if (id) setRunningJobs(prev => ({ ...prev, [id]: data.stage }));
    };
    const onCompleted = (data: any) => {
      const id = data?.job_id;
      if (id) setRunningJobs(prev => { const n = { ...prev }; delete n[id]; return n; });
      toast.success(t('backup.completed'), t('backup.successfullyCompleted', { name: data?.job_name || t('history.job') }));
      loadData();
    };
    const onFailed = (data: any) => {
      const id = data?.job_id;
      if (id) setRunningJobs(prev => { const n = { ...prev }; delete n[id]; return n; });
      toast.error(t('backup.failed'), t('backup.failedWithError', { name: data?.job_name || t('history.job'), error: data?.error || t('backup.unknownError') }));
      loadData();
    };

    EventsOn('backup:started', onStarted);
    EventsOn('backup:progress', onProgress);
    EventsOn('backup:completed', onCompleted);
    EventsOn('backup:failed', onFailed);

    return () => {
      EventsOff('backup:started', onStarted);
      EventsOff('backup:progress', onProgress);
      EventsOff('backup:completed', onCompleted);
      EventsOff('backup:failed', onFailed);
    };
  }, [loadData]);

  // Group jobs by database_id
  const jobsByDb = jobs.reduce((acc: Record<number, any[]>, job: any) => {
    (acc[job.database_id] = acc[job.database_id] || []).push(job);
    return acc;
  }, {});

  // ==================== DB MODAL ====================

  const openDbModal = (db?: any) => {
    if (db) { setDbForm(db); setEditingDbId(db.id); }
    else { setDbForm({ ...emptyDbForm, created_at: now(), updated_at: now() }); setEditingDbId(null); }
    setShowDbModal(true);
  };

  const closeDbModal = () => { setShowDbModal(false); setDbForm(emptyDbForm); setEditingDbId(null); };

  const handleDbTypeChange = (type: string) => {
    const dbType = dbTypes.find(t => t.value === type);
    setDbForm({ ...dbForm, type, port: dbType?.port || dbForm.port, auth_database: type === 'mongodb' ? (dbForm.auth_database || '') : '' });
  };

  const handleUriPaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    const uri = e.clipboardData.getData('text').trim();
    if (!uri) return;
    const parsed = parseConnectionURI(uri);
    if (parsed) {
      e.preventDefault(); // Input'a yazmayı engelle, alanlar zaten dolacak
      setDbForm({
        ...dbForm,
        ...parsed,
        name: dbForm.name, // Kullanici girdigiyse koru
        created_at: dbForm.created_at,
        updated_at: dbForm.updated_at,
      });
      toast.success(t('databases.uriParsed'), t('databases.uriParsedDesc'));
    }
    // Geçersiz URI ise normal input olarak yazılsın, hata gösterme
  };

  const handleDbSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingDb(true);
    try {
      const payload = { ...dbForm, created_at: dbForm.created_at || now(), updated_at: now() };
      if (editingDbId) {
        await UpdateDatabase(payload);
        toast.success(t('common.updated'), t('databases.connectionUpdated'));
      } else {
        await AddDatabase(payload);
        toast.success(t('common.added'), t('databases.connectionAdded'));
      }
      await loadData();
      closeDbModal();
    } catch (error) {
      toast.error(t('common.savingError'), String(error));
    } finally {
      setSavingDb(false);
    }
  };

  const handleDbTest = async () => {
    setTestingDb(true);
    try {
      await TestDatabaseConnection({ ...dbForm, created_at: dbForm.created_at || now(), updated_at: dbForm.updated_at || now() });
      toast.success(t('common.connectionSuccess'), t('databases.connectedSuccess'));
    } catch (error) {
      toast.error(t('common.connectionFailed'), String(error));
    } finally {
      setTestingDb(false);
    }
  };

  const handleDbDelete = async (id: number) => {
    const dbJobs = jobsByDb[id] || [];
    const message = dbJobs.length > 0
      ? t('databases.deleteLinkedJobs', { count: dbJobs.length })
      : t('databases.deleteConfirm');
    const confirmed = await confirm({
      title: t('databases.deleteTitle'), message,
      confirmText: t('common.delete'), cancelText: t('common.giveUp'), type: 'danger',
    });
    if (!confirmed) return;
    try {
      await DeleteDatabase(id);
      await loadData();
      toast.success(t('common.deleted'), t('databases.connectionDeleted'));
    } catch {
      toast.error(t('common.deleteError'), t('databases.deleteFailed'));
    }
  };

  // ==================== JOB (BACKUP TARGET) MODAL ====================

  const openJobModal = (databaseId: number, job?: any) => {
    setJobDatabaseId(databaseId);
    if (job) {
      setEditingJobId(job.id);
      setStorageMode('existing');
      setSelectedStorageId(job.storage_id);
      setJobForm({
        schedule_type: job.schedule_type || 'manual',
        compression: job.compression || 'gzip',
        retention_days: job.retention_days || 7,
        encryption: job.encryption || false,
        encryption_key: job.encryption_key || '',
        is_active: job.is_active !== undefined ? job.is_active : true,
        custom_prefix: job.custom_prefix || '',
        custom_folder: job.custom_folder || '',
        folder_grouping: job.folder_grouping || '',
      });
      setScheduleConfig(JSON.parse(job.schedule_config || '{}'));
    } else {
      setEditingJobId(null);
      setStorageMode(storages.length > 0 ? 'existing' : 'new');
      setSelectedStorageId(storages[0]?.id || 0);
      setNewStorageForm({ name: '', type: 'local' });
      setNewStorageConfig({ path: '' });
      setJobForm({
        schedule_type: 'manual', compression: 'gzip', retention_days: 7,
        encryption: false, encryption_key: '', is_active: true,
        custom_prefix: '', custom_folder: '', folder_grouping: '',
      });
      setScheduleConfig({});
    }
    setShowJobModal(true);
  };

  const closeJobModal = () => {
    setShowJobModal(false);
    setEditingJobId(null);
    setJobDatabaseId(0);
  };

  const handleNewStorageTypeChange = (type: string) => {
    setNewStorageForm({ ...newStorageForm, type });
    switch (type) {
      case 'local': setNewStorageConfig({ path: '' }); break;
      case 'ftp': setNewStorageConfig({ host: '', port: 21, username: '', password: '', path: '/' }); break;
      case 'sftp': setNewStorageConfig({ host: '', port: 22, username: '', password: '', private_key: '', path: '/' }); break;
      case 's3': setNewStorageConfig({ endpoint: '', region: 'us-east-1', bucket: '', access_key_id: '', secret_access_key: '', path: '/', use_ssl: true }); break;
    }
  };

  const handleStorageTest = async () => {
    setTestingStorage(true);
    try {
      await TestStorageConnection({
        id: 0, name: newStorageForm.name, type: newStorageForm.type,
        config: JSON.stringify(newStorageConfig), created_at: now(),
      } as any);
      toast.success(t('common.connectionSuccess'), t('storage.connectedSuccess'));
    } catch (error) {
      toast.error(t('common.connectionFailed'), String(error));
    } finally {
      setTestingStorage(false);
    }
  };

  const handleSelectFolder = async () => {
    if (isWebMode()) return; // Web mode uses text input instead
    try {
      const folder = await SelectFolder();
      if (folder) setNewStorageConfig({ ...newStorageConfig, path: folder });
    } catch (error) {
      toast.error(t('storage.folderSelectFailed'), String(error));
    }
  };

  const handleJobSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingJob(true);
    try {
      let storageId = selectedStorageId;

      if (storageMode === 'new') {
        storageId = await AddStorageTarget({
          id: 0, name: newStorageForm.name, type: newStorageForm.type,
          config: JSON.stringify(newStorageConfig), created_at: now(),
        } as any);
      }

      const dbName = databases.find(d => d.id === jobDatabaseId)?.name || 'DB';
      const storageName = storageMode === 'new'
        ? newStorageForm.name
        : storages.find(s => s.id === storageId)?.name || 'Storage';
      const autoName = `${dbName} → ${storageName}`;

      const jobPayload: any = {
        name: autoName,
        database_id: jobDatabaseId,
        storage_id: storageId,
        schedule_type: jobForm.schedule_type,
        schedule_config: JSON.stringify(scheduleConfig),
        compression: jobForm.compression,
        encryption: jobForm.encryption,
        encryption_key: jobForm.encryption_key || '',
        retention_days: jobForm.retention_days,
        is_active: jobForm.is_active,
        custom_prefix: jobForm.custom_prefix || '',
        custom_folder: jobForm.custom_folder || '',
        folder_grouping: jobForm.folder_grouping || '',
        created_at: now(),
      };

      if (editingJobId) {
        jobPayload.id = editingJobId;
        jobPayload.storage_id = selectedStorageId;
        await UpdateBackupJob(jobPayload);
        toast.success(t('common.updated'), t('backup.targetUpdated'));
      } else {
        await AddBackupJob(jobPayload);
        toast.success(t('common.added'), t('backup.targetAdded'));
      }
      await loadData();
      closeJobModal();
    } catch (error) {
      toast.error(t('common.savingError'), String(error));
    } finally {
      setSavingJob(false);
    }
  };

  // ==================== JOB ACTIONS ====================

  const handleJobToggle = async (id: number, active: boolean) => {
    try {
      await ToggleBackupJobActive(id, active);
      await loadData();
      toast.info(active ? t('common.activated') : t('common.deactivated'), t('databases.jobToggleSuccess', { status: active ? t('common.active') : t('common.inactive') }));
    } catch {
      toast.error(t('common.error'), t('databases.jobToggleFailed'));
    }
  };

  const handleJobRunNow = async (id: number) => {
    setRunningJobs(prev => ({ ...prev, [id]: 'starting' }));
    try {
      await RunBackupNow(id);
    } catch (error) {
      toast.error(t('backup.failed'), String(error));
      setRunningJobs(prev => { const n = { ...prev }; delete n[id]; return n; });
    }
  };

  const handleJobDelete = async (id: number) => {
    const confirmed = await confirm({
      title: t('backup.deleteTargetTitle'),
      message: t('backup.deleteTargetMessage'),
      confirmText: t('common.delete'), cancelText: t('common.giveUp'), type: 'danger',
    });
    if (!confirmed) return;
    try {
      await DeleteBackupJob(id);
      await loadData();
      toast.success(t('common.deleted'), t('backup.targetDeleted'));
    } catch {
      toast.error(t('common.deleteError'), t('backup.targetDeleteFailed'));
    }
  };

  // ==================== RENDER HELPERS ====================

  const renderStorageConfigFields = () => {
    switch (newStorageForm.type) {
      case 'local':
        return (
          <div className="space-y-2">
            <Label>{t('storage.folderPath')}</Label>
            <div className="flex gap-2">
              <Input className="flex-1" placeholder="/path/to/backups" value={newStorageConfig.path || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, path: e.target.value })} required />
              {!isWebMode() && (
                <Button type="button" variant="secondary" onClick={handleSelectFolder}>
                  <FolderOpen className="w-4 h-4" /> {t('common.browse')}
                </Button>
              )}
            </div>
          </div>
        );
      case 'ftp': case 'sftp':
        return (
          <>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2"><Label>{t('common.host')}</Label><Input placeholder="ftp.example.com" value={newStorageConfig.host || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, host: e.target.value })} required /></div>
              <div className="space-y-2"><Label>{t('common.port')}</Label><Input type="number" value={newStorageConfig.port || (newStorageForm.type === 'ftp' ? 21 : 22)} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, port: parseInt(e.target.value) || 0 })} required /></div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2"><Label>{t('common.username')}</Label><Input value={newStorageConfig.username || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, username: e.target.value })} required /></div>
              <div className="space-y-2"><Label>{t('common.password')}</Label><Input type="password" value={newStorageConfig.password || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, password: e.target.value })} /></div>
            </div>
            <div className="space-y-2"><Label>{t('storage.remotePath')}</Label><Input placeholder="/backups" value={newStorageConfig.path || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, path: e.target.value })} required /></div>
          </>
        );
      case 's3':
        return (
          <>
            <div className="space-y-2"><Label>{t('storage.s3.endpoint')}</Label><Input placeholder="https://s3.amazonaws.com" value={newStorageConfig.endpoint || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, endpoint: e.target.value })} /></div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2"><Label>{t('storage.s3.region')}</Label><Input placeholder="us-east-1" value={newStorageConfig.region || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, region: e.target.value })} required /></div>
              <div className="space-y-2"><Label>{t('storage.s3.bucket')}</Label><Input placeholder="my-backups" value={newStorageConfig.bucket || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, bucket: e.target.value })} required /></div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2"><Label>{t('storage.s3.accessKeyId')}</Label><Input value={newStorageConfig.access_key_id || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, access_key_id: e.target.value })} required /></div>
              <div className="space-y-2"><Label>{t('storage.s3.secretAccessKey')}</Label><Input type="password" value={newStorageConfig.secret_access_key || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, secret_access_key: e.target.value })} required /></div>
            </div>
            <div className="space-y-2"><Label>{t('storage.s3.pathPrefix')}</Label><Input placeholder="/backups" value={newStorageConfig.path || ''} onChange={(e) => setNewStorageConfig({ ...newStorageConfig, path: e.target.value })} /></div>
          </>
        );
      default: return null;
    }
  };

  const renderScheduleFields = () => {
    switch (jobForm.schedule_type) {
      case 'interval':
        return (
          <div className="space-y-2">
            <Label>{t('schedule.interval')}</Label>
            <Input type="number" min={1} value={scheduleConfig.interval_minutes ?? ''} onChange={(e) => setScheduleConfig({ ...scheduleConfig, interval_minutes: e.target.value === '' ? '' : parseInt(e.target.value) })} required />
          </div>
        );
      case 'daily':
        return (
          <div className="space-y-2">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2"><Label>{t('schedule.hour')}</Label><Input type="number" min={0} max={23} value={scheduleConfig.hour ?? ''} onChange={(e) => setScheduleConfig({ ...scheduleConfig, hour: e.target.value === '' ? '' : parseInt(e.target.value) })} required /></div>
              <div className="space-y-2"><Label>{t('schedule.minute')}</Label><Input type="number" min={0} max={59} value={scheduleConfig.minute ?? ''} onChange={(e) => setScheduleConfig({ ...scheduleConfig, minute: e.target.value === '' ? '' : parseInt(e.target.value) })} required /></div>
            </div>
            <p className="text-[10px] text-zinc-500">{t('schedule.timeDescription')}</p>
          </div>
        );
      case 'weekly':
        return (
          <>
            <div className="space-y-2">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2"><Label>{t('schedule.hour')}</Label><Input type="number" min={0} max={23} value={scheduleConfig.hour ?? ''} onChange={(e) => setScheduleConfig({ ...scheduleConfig, hour: e.target.value === '' ? '' : parseInt(e.target.value) })} required /></div>
                <div className="space-y-2"><Label>{t('schedule.minute')}</Label><Input type="number" min={0} max={59} value={scheduleConfig.minute ?? ''} onChange={(e) => setScheduleConfig({ ...scheduleConfig, minute: e.target.value === '' ? '' : parseInt(e.target.value) })} required /></div>
              </div>
              <p className="text-[10px] text-zinc-500">{t('schedule.timeDescriptionWeekly')}</p>
            </div>
            <div className="space-y-2">
              <Label>{t('schedule.days')}</Label>
              <div className="flex flex-wrap gap-2">
                {weekdays.map((day, index) => (
                  <label key={index} className="flex items-center gap-2 cursor-pointer bg-zinc-800 px-3 py-1.5 rounded-lg border border-zinc-700 hover:border-zinc-600 transition-colors">
                    <input type="checkbox" className="w-3.5 h-3.5 rounded accent-blue-500"
                      checked={scheduleConfig.weekdays?.includes(index) || false}
                      onChange={(e) => {
                        const days = scheduleConfig.weekdays || [];
                        setScheduleConfig({ ...scheduleConfig, weekdays: e.target.checked ? [...days, index] : days.filter((d: number) => d !== index) });
                      }}
                    />
                    <span className="text-xs text-zinc-300">{day}</span>
                  </label>
                ))}
              </div>
            </div>
          </>
        );
      default: return null;
    }
  };

  // ==================== RENDER ====================

  if (loading) return <PageLoader />;

  return (
    <div>
      <PageHeader title={t('databases.title')} description={t('databases.description')}>
        <Button onClick={() => openDbModal()}>
          <Plus className="w-4 h-4" /> {t('databases.addNew')}
        </Button>
      </PageHeader>

      {databases.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900">
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="w-14 h-14 rounded-full bg-zinc-800 flex items-center justify-center mb-4">
              <Database className="w-7 h-7 text-zinc-600" />
            </div>
            <p className="text-zinc-400 text-sm mb-1">{t('databases.empty')}</p>
            <p className="text-zinc-500 text-xs mb-4">{t('databases.emptyDesc')}</p>
            <Button onClick={() => openDbModal()}>
              <Plus className="w-4 h-4" /> {t('databases.addFirst')}
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {databases.map((db: any) => {
            const style = getDbStyle(db.type);
            const dbJobs = jobsByDb[db.id] || [];
            return (
              <Card key={db.id} className="border-zinc-800 bg-zinc-900">
                <CardContent className="p-5">
                  {/* DB Header */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className={`w-10 h-10 rounded-lg ${style.bg} flex items-center justify-center flex-shrink-0`}>
                        <Database className={`w-5 h-5 ${style.color}`} />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold text-white text-sm truncate">{db.name}</h3>
                          <Badge variant={db.type === 'postgres' ? 'info' : db.type === 'mysql' ? 'orange' : 'success'} className="text-[10px]">
                            {style.label}
                          </Badge>
                        </div>
                        <p className="text-xs text-zinc-500 font-mono truncate">{db.host}:{db.port}/{db.database_name}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <Button variant="ghost" size="sm" onClick={() => openDbModal(db)}>
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon" className="text-red-400 hover:text-red-300 hover:bg-red-500/10" onClick={() => handleDbDelete(db.id)}>
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>

                  {/* Backup Targets Section */}
                  <div className="mt-4 pt-4 border-t border-zinc-800">
                    <div className="flex items-center justify-between mb-3">
                      <h4 className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{t('backup.targets')}</h4>
                      <Button variant="ghost" size="sm" onClick={() => openJobModal(db.id)}>
                        <Plus className="w-3.5 h-3.5" /> {t('backup.newTarget')}
                      </Button>
                    </div>

                    {dbJobs.length === 0 ? (
                      <div className="flex flex-col items-center py-6 rounded-lg border border-dashed border-zinc-800">
                        <Target className="w-6 h-6 text-zinc-700 mb-2" />
                        <p className="text-xs text-zinc-500 mb-2">{t('backup.noTargets')}</p>
                        <Button variant="outline" size="sm" onClick={() => openJobModal(db.id)}>
                          <Plus className="w-3.5 h-3.5" /> {t('backup.addFirstTarget')}
                        </Button>
                      </div>
                    ) : (
                      <div className="space-y-2">
                        {dbJobs.map((job: any) => {
                          const storage = storages.find(s => s.id === job.storage_id);
                          const stStyle = storage ? getStorageStyle(storage.type) : storageTypesMeta[0];
                          const StIcon = stStyle.icon;
                          const jobStage = runningJobs[job.id];
                          const isRunning = !!jobStage;
                          return (
                            <div key={job.id} className={`rounded-lg transition-colors ${isRunning ? 'bg-blue-500/5 border border-blue-500/20' : 'bg-zinc-800/50 hover:bg-zinc-800'}`}>
                              <div className="flex items-center gap-3 p-3">
                                <div className={`w-8 h-8 rounded-md ${stStyle.bg} flex items-center justify-center flex-shrink-0`}>
                                  <StIcon className={`w-4 h-4 ${stStyle.color}`} />
                                </div>
                                <div className="flex-1 min-w-0">
                                  <div className="flex items-center gap-2">
                                    <span className="text-sm text-white truncate">{storage?.name || t('common.unknown')}</span>
                                    <span className="text-xs text-zinc-600 truncate hidden sm:inline">
                                      ({storage ? getStorageDescription(storage) : '-'})
                                    </span>
                                  </div>
                                  <div className="flex items-center gap-2 mt-0.5">
                                    <Badge variant="info" className="text-[10px]">{getScheduleLabel(job)}</Badge>
                                    <span className="text-[10px] text-zinc-500">{getCompressionLabel(job.compression)}</span>
                                    <span className="text-[10px] text-zinc-500">{job.retention_days} {t('common.days')}</span>
                                    {job.custom_folder && <span className="text-[10px] text-zinc-500 font-mono">/{job.custom_folder}</span>}
                                  </div>
                                </div>
                                <div className="flex items-center gap-1 flex-shrink-0">
                                  <Switch checked={job.is_active} onCheckedChange={(checked) => handleJobToggle(job.id, checked)} />
                                  <Button variant="ghost" size="icon" onClick={() => handleJobRunNow(job.id)} disabled={isRunning} title={t('backup.runNow')}>
                                    {isRunning ? (
                                      <Loader2 className="w-4 h-4 animate-spin text-blue-400" />
                                    ) : (
                                      <Play className="w-4 h-4 text-emerald-400" />
                                    )}
                                  </Button>
                                  <Button variant="ghost" size="icon" onClick={() => openJobModal(db.id, job)} title={t('common.edit')}>
                                    <Pencil className="w-3.5 h-3.5" />
                                  </Button>
                                  <Button variant="ghost" size="icon" className="text-red-400 hover:text-red-300 hover:bg-red-500/10" onClick={() => handleJobDelete(job.id)} title={t('common.delete')}>
                                    <Trash2 className="w-3.5 h-3.5" />
                                  </Button>
                                </div>
                              </div>
                              {isRunning && (
                                <div className="px-3 pb-3">
                                  <div className="flex items-center gap-2 mb-1.5">
                                    <Loader2 className="w-3 h-3 animate-spin text-blue-400 flex-shrink-0" />
                                    <span className="text-xs text-blue-400">{stageLabels[jobStage] || t('backup.stages.starting')}</span>
                                  </div>
                                  <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                                    <div className="h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full animate-progress" />
                                  </div>
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* ==================== DB DIALOG ==================== */}
      <Dialog open={showDbModal} onOpenChange={(open) => { if (!open) closeDbModal(); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingDbId ? t('databases.editTitle') : t('databases.newTitle')}</DialogTitle>
            <DialogDescription>{t('databases.formDescription')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleDbSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label>{t('databases.connectionName')}</Label>
              <Input placeholder={t('databases.connectionNamePlaceholder')} value={dbForm.name || ''} onChange={(e) => setDbForm({ ...dbForm, name: e.target.value })} required />
            </div>
            <div className="space-y-2">
              <Label>{t('databases.dbType')}</Label>
              <Select value={dbForm.type || 'postgres'} onChange={(e) => handleDbTypeChange(e.target.value)}>
                {dbTypes.map((type) => (<option key={type.value} value={type.value}>{type.label}</option>))}
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>{t('common.host')}</Label>
                <Input placeholder="localhost" value={dbForm.host || ''} onChange={(e) => setDbForm({ ...dbForm, host: e.target.value })} required />
              </div>
              <div className="space-y-2">
                <Label>{t('common.port')}</Label>
                <Input type="number" value={dbForm.port || 5432} onChange={(e) => setDbForm({ ...dbForm, port: parseInt(e.target.value) || 0 })} required />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>{t('common.username')}</Label>
                <Input placeholder="root" value={dbForm.username || ''} onChange={(e) => setDbForm({ ...dbForm, username: e.target.value })} required />
              </div>
              <div className="space-y-2">
                <Label>{t('common.password')}</Label>
                <Input type="password" value={dbForm.password || ''} onChange={(e) => setDbForm({ ...dbForm, password: e.target.value })} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t('databases.dbName')}</Label>
              <Input placeholder="mydb" value={dbForm.database_name || ''} onChange={(e) => setDbForm({ ...dbForm, database_name: e.target.value })} required />
            </div>
            {dbForm.type === 'mongodb' && (
              <>
                <div className="space-y-2">
                  <Label>{t('databases.authDatabase')}</Label>
                  <Input placeholder={t('databases.authDatabasePlaceholder')} value={dbForm.auth_database || ''} onChange={(e) => setDbForm({ ...dbForm, auth_database: e.target.value })} />
                  <p className="text-[10px] text-zinc-500">{t('databases.authDatabaseDesc')}</p>
                </div>
                <div className="space-y-2">
                  <Label>{t('databases.connectionUri')}</Label>
                  <Input
                    placeholder={t('databases.connectionUriPlaceholder')}
                    onPaste={handleUriPaste}
                  />
                  <p className="text-[10px] text-zinc-500">{t('databases.connectionUriDesc')}</p>
                </div>
              </>
            )}
            <div className="flex items-center gap-3">
              <Switch checked={dbForm.ssl_enabled || false} onCheckedChange={(checked) => setDbForm({ ...dbForm, ssl_enabled: checked })} />
              <Label className="cursor-pointer">{t('common.ssl')}</Label>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleDbTest} disabled={testingDb}>
                {testingDb ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plug className="w-4 h-4" />}
                {testingDb ? t('common.testingConnection') : t('common.connectionTest')}
              </Button>
              <Button type="button" variant="secondary" onClick={closeDbModal}>{t('common.cancel')}</Button>
              <Button type="submit" disabled={savingDb}>
                {savingDb ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
                {savingDb ? t('common.saving') : t('common.save')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* ==================== JOB (BACKUP TARGET) DIALOG ==================== */}
      <Dialog open={showJobModal} onOpenChange={(open) => { if (!open) closeJobModal(); }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingJobId ? t('jobDialog.editTitle') : t('jobDialog.newTitle')}</DialogTitle>
            <DialogDescription>{t('jobDialog.description')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleJobSubmit} className="space-y-5">

            {/* Section 1: Storage */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-white">{t('storage.title')}</h3>

              {!editingJobId && storages.length > 0 && (
                <div className="flex gap-2">
                  <Button type="button" variant={storageMode === 'existing' ? 'default' : 'outline'} size="sm" onClick={() => setStorageMode('existing')}>
                    {t('storage.selectExisting')}
                  </Button>
                  <Button type="button" variant={storageMode === 'new' ? 'default' : 'outline'} size="sm" onClick={() => setStorageMode('new')}>
                    {t('storage.createNew')}
                  </Button>
                </div>
              )}

              {(storageMode === 'existing' || editingJobId) ? (
                <div className="space-y-2">
                  <Label>{t('storage.storage')}</Label>
                  <Select value={selectedStorageId || ''} onChange={(e) => setSelectedStorageId(parseInt(e.target.value))} required>
                    {storages.map((s: any) => (<option key={s.id} value={s.id}>{s.name} ({storageTypes.find(st => st.value === s.type)?.label})</option>))}
                  </Select>
                </div>
              ) : (
                <div className="space-y-3 p-3 rounded-lg border border-zinc-800 bg-zinc-800/30">
                  <div className="space-y-2">
                    <Label>{t('storage.targetName')}</Label>
                    <Input placeholder={t('storage.targetNamePlaceholder')} value={newStorageForm.name} onChange={(e) => setNewStorageForm({ ...newStorageForm, name: e.target.value })} required />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('storage.storageType')}</Label>
                    <Select value={newStorageForm.type} onChange={(e) => handleNewStorageTypeChange(e.target.value)}>
                      {storageTypes.map((type) => (<option key={type.value} value={type.value}>{type.label}</option>))}
                    </Select>
                  </div>
                  {renderStorageConfigFields()}
                  <Button type="button" variant="outline" size="sm" onClick={handleStorageTest} disabled={testingStorage}>
                    {testingStorage ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plug className="w-4 h-4" />}
                    {testingStorage ? t('common.testingConnection') : t('common.connectionTest')}
                  </Button>
                </div>
              )}
            </div>

            {/* Section 2: Schedule */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-white">{t('schedule.title')}</h3>
              <div className="space-y-2">
                <Select value={jobForm.schedule_type || 'manual'} onChange={(e) => setJobForm({ ...jobForm, schedule_type: e.target.value })}>
                  {scheduleTypes.map((type) => (<option key={type.value} value={type.value}>{type.label}</option>))}
                </Select>
              </div>
              {renderScheduleFields()}
            </div>

            {/* Section 3: File & Folder */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-white">{t('fileFolder.title')}</h3>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>{t('fileFolder.filePrefix')}</Label>
                  <Input
                    placeholder={databases.find(d => d.id === jobDatabaseId)?.database_name || 'mydb'}
                    value={jobForm.custom_prefix || ''}
                    onChange={(e) => setJobForm({ ...jobForm, custom_prefix: e.target.value })}
                  />
                  <p className="text-[10px] text-zinc-500">
                    {t('common.example')} <span className="text-zinc-400 font-mono">{jobForm.custom_prefix || databases.find(d => d.id === jobDatabaseId)?.database_name || 'mydb'}_20260221_030000.sql</span>
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>{t('fileFolder.saveFolder')}</Label>
                  <Input
                    placeholder={t('fileFolder.emptyDefault')}
                    value={jobForm.custom_folder || ''}
                    onChange={(e) => setJobForm({ ...jobForm, custom_folder: e.target.value })}
                  />
                  <p className="text-[10px] text-zinc-500">{t('common.example')} dbyedekleri/testyedekler</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-3">
                  <Switch
                    checked={!!jobForm.folder_grouping}
                    onCheckedChange={(checked) => setJobForm({ ...jobForm, folder_grouping: checked ? 'monthly' : '' })}
                  />
                  <Label className="cursor-pointer">{t('fileFolder.folderGrouping')}</Label>
                </div>
                {jobForm.folder_grouping && (
                  <div className="space-y-2 pl-12">
                    <Select value={jobForm.folder_grouping} onChange={(e) => setJobForm({ ...jobForm, folder_grouping: e.target.value })}>
                      <option value="daily">{t('fileFolder.groupDaily')}</option>
                      <option value="monthly">{t('fileFolder.groupMonthly')}</option>
                      <option value="yearly">{t('fileFolder.groupYearly')}</option>
                    </Select>
                    <p className="text-[10px] text-zinc-500">
                      {t('common.example')} <span className="text-zinc-400 font-mono">{t(`fileFolder.group${jobForm.folder_grouping.charAt(0).toUpperCase() + jobForm.folder_grouping.slice(1)}Example`)}</span>
                    </p>
                  </div>
                )}
              </div>
            </div>

            {/* Section 4: Options */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-white">{t('options.title')}</h3>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>{t('options.compression')}</Label>
                  <Select value={jobForm.compression || 'gzip'} onChange={(e) => setJobForm({ ...jobForm, compression: e.target.value })}>
                    {compressionTypes.map((type) => (<option key={type.value} value={type.value}>{type.label}</option>))}
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t('options.retentionDays')}</Label>
                  <Input type="number" min={1} value={jobForm.retention_days ?? ''} onChange={(e) => setJobForm({ ...jobForm, retention_days: e.target.value === '' ? '' : parseInt(e.target.value) })} required />
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Switch checked={jobForm.encryption || false} onCheckedChange={(checked) => setJobForm({ ...jobForm, encryption: checked })} />
                <Label className="cursor-pointer">{t('options.encryptBackup')}</Label>
              </div>
              {jobForm.encryption && (
                <div className="space-y-2">
                  <Label>{t('options.encryptionKey')}</Label>
                  <Input type="password" value={jobForm.encryption_key || ''} onChange={(e) => setJobForm({ ...jobForm, encryption_key: e.target.value })} required />
                </div>
              )}
            </div>

            <DialogFooter>
              <Button type="button" variant="secondary" onClick={closeJobModal}>{t('common.cancel')}</Button>
              <Button type="submit" disabled={savingJob}>
                {savingJob ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
                {savingJob ? t('common.saving') : t('common.save')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
