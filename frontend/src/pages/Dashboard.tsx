import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { GetDashboardStats, GetBackupHistory, RunBackupNow, GetBackupJobs } from '../api/client';
import { EventsOn, EventsOff } from '../api/events';
import { PageLoader } from '../components/Spinner';
import PageHeader from '../components/PageHeader';
import { Card, CardContent } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Button } from '../components/ui/button';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '../components/ui/table';
import { Database, HardDrive, Clock, CheckCircle2, XCircle, Play, ArrowRight, Activity, Loader2 } from 'lucide-react';
import { useToast } from '../components/Toast';

export default function Dashboard() {
  const { t, i18n } = useTranslation();
  const [stats, setStats] = useState<any>(null);
  const [recentBackups, setRecentBackups] = useState<any[]>([]);
  const [jobs, setJobs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [runningJobs, setRunningJobs] = useState<Record<number, string>>({});
  const navigate = useNavigate();
  const toast = useToast();

  const stageLabels: Record<string, string> = {
    starting: t('backup.stages.starting'),
    dumping: t('backup.stages.dumping'),
    processing: t('backup.stages.processing'),
    uploading: t('backup.stages.uploading'),
  };

  const loadData = useCallback(async () => {
    try {
      const [statsData, historyData, jobsData] = await Promise.all([
        GetDashboardStats(),
        GetBackupHistory({ job_id: 0, status: '', limit: 5 }),
        GetBackupJobs(),
      ]);
      setStats(statsData);
      setRecentBackups(historyData || []);
      setJobs(jobsData || []);
    } catch (error) {
      console.error(t('dashboard.loadFailed'), error);
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => { loadData(); }, [loadData]);

  // Real-time backup events
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

  const handleQuickBackup = async (jobId: number) => {
    setRunningJobs(prev => ({ ...prev, [jobId]: 'starting' }));
    try {
      await RunBackupNow(jobId);
    } catch (error) {
      toast.error(t('common.error'), String(error));
      setRunningJobs(prev => { const n = { ...prev }; delete n[jobId]; return n; });
    }
  };

  const formatDate = (dateStr: string) => new Date(dateStr).toLocaleString(i18n.language === 'en' ? 'en-US' : 'tr-TR');
  const formatSize = (bytes: number) => {
    if (!bytes || bytes <= 0 || !isFinite(bytes)) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), sizes.length - 1);
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (loading) return <PageLoader />;

  const statCards = [
    { label: t('dashboard.database'), value: stats?.total_databases || 0, icon: Database, color: 'text-blue-400', bg: 'bg-blue-500/10' },
    { label: t('dashboard.storage'), value: stats?.total_storages || 0, icon: HardDrive, color: 'text-purple-400', bg: 'bg-purple-500/10' },
    { label: t('dashboard.activeJobs'), value: `${stats?.active_jobs || 0}/${stats?.total_jobs || 0}`, icon: Clock, color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
    { label: t('dashboard.last24Hours'), value: null, icon: Activity, color: 'text-amber-400', bg: 'bg-amber-500/10',
      custom: (
        <div className="flex items-center gap-2 mt-1">
          <span className="flex items-center gap-1 text-emerald-400 font-semibold text-lg">
            <CheckCircle2 className="w-4 h-4" /> {stats?.last_24h_success || 0}
          </span>
          <span className="text-zinc-600">/</span>
          <span className="flex items-center gap-1 text-red-400 font-semibold text-lg">
            <XCircle className="w-4 h-4" /> {stats?.last_24h_failed || 0}
          </span>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader title={t('dashboard.title')} description={t('dashboard.description')} />

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {statCards.map((stat) => (
          <Card key={stat.label} className="border-zinc-800 bg-zinc-900 hover:border-zinc-700 transition-colors">
            <CardContent className="p-5">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{stat.label}</p>
                  {stat.custom ? stat.custom : (
                    <p className="text-2xl font-bold text-white mt-1">{stat.value}</p>
                  )}
                </div>
                <div className={`w-10 h-10 rounded-lg ${stat.bg} flex items-center justify-center`}>
                  <stat.icon className={`w-5 h-5 ${stat.color}`} />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Backups */}
        <div className="lg:col-span-2">
          <Card className="border-zinc-800 bg-zinc-900">
            <div className="flex items-center justify-between p-5 pb-0">
              <h3 className="text-sm font-semibold text-white">{t('dashboard.recentBackups')}</h3>
              <Button variant="ghost" size="sm" onClick={() => navigate('/history')}>
                {t('dashboard.viewAll')} <ArrowRight className="w-3 h-3" />
              </Button>
            </div>
            <CardContent className="p-5">
              {recentBackups.length === 0 ? (
                <div className="text-center py-8">
                  <Activity className="w-10 h-10 text-zinc-700 mx-auto mb-3" />
                  <p className="text-sm text-zinc-500">{t('dashboard.noBackups')}</p>
                  <Button variant="outline" size="sm" className="mt-3" onClick={() => navigate('/databases')}>
                    {t('dashboard.createJob')}
                  </Button>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent">
                      <TableHead>{t('common.status')}</TableHead>
                      <TableHead>{t('common.file')}</TableHead>
                      <TableHead>{t('common.size')}</TableHead>
                      <TableHead>{t('common.date')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentBackups.map((backup: any) => (
                      <TableRow key={backup.id}>
                        <TableCell>
                          <Badge variant={backup.status === 'success' ? 'success' : backup.status === 'failed' ? 'destructive' : 'warning'}>
                            {backup.status === 'success' ? t('common.successful') : backup.status === 'failed' ? t('common.failed') : t('common.running')}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs max-w-[200px] truncate">{backup.file_name || '-'}</TableCell>
                        <TableCell>{formatSize(backup.file_size)}</TableCell>
                        <TableCell className="text-xs">{formatDate(backup.started_at)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <div>
          <Card className="border-zinc-800 bg-zinc-900">
            <div className="p-5 pb-0">
              <h3 className="text-sm font-semibold text-white">{t('dashboard.quickBackup')}</h3>
            </div>
            <CardContent className="p-5">
              {jobs.length === 0 ? (
                <div className="text-center py-4">
                  <p className="text-sm text-zinc-500 mb-3">{t('dashboard.noJobs')}</p>
                  <Button variant="outline" size="sm" onClick={() => navigate('/databases')}>
                    {t('dashboard.createJob')}
                  </Button>
                </div>
              ) : (
                <div className="space-y-2">
                  {jobs.slice(0, 5).map((job: any) => {
                    const jobStage = runningJobs[job.id];
                    const isRunning = !!jobStage;
                    return (
                      <div key={job.id} className={`rounded-lg transition-colors ${isRunning ? 'bg-blue-500/5 border border-blue-500/20' : 'hover:bg-zinc-800/50'}`}>
                        <div className="flex items-center justify-between gap-2 p-2">
                          <div className="min-w-0">
                            <p className="text-sm text-white truncate">{job.name}</p>
                            <p className="text-xs text-zinc-500">{job.is_active ? t('common.active') : t('common.inactive')}</p>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleQuickBackup(job.id)}
                            disabled={isRunning}
                            className="flex-shrink-0"
                          >
                            {isRunning ? (
                              <Loader2 className="w-4 h-4 animate-spin text-blue-400" />
                            ) : (
                              <Play className="w-4 h-4 text-emerald-400" />
                            )}
                          </Button>
                        </div>
                        {isRunning && (
                          <div className="px-2 pb-2">
                            <p className="text-[11px] text-blue-400 mb-1">{stageLabels[jobStage] || t('backup.stages.starting')}</p>
                            <div className="h-1 bg-zinc-800 rounded-full overflow-hidden">
                              <div className="h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full animate-progress" />
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
