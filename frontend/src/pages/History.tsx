import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { GetBackupHistory, GetBackupJobs } from '../api/client';
import Spinner from '../components/Spinner';
import PageHeader from '../components/PageHeader';
import { Card, CardContent } from '../components/ui/card';
import { Select } from '../components/ui/select';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '../components/ui/table';
import { History as HistoryIcon, CheckCircle2, XCircle, Loader2 } from 'lucide-react';

export default function History() {
  const { t, i18n } = useTranslation();
  const [history, setHistory] = useState<any[]>([]);
  const [jobs, setJobs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({ job_id: 0, status: '', limit: 50 });

  useEffect(() => {
    const loadAll = async () => {
      setLoading(true);
      try {
        const [jobsData, historyData] = await Promise.all([
          GetBackupJobs(),
          GetBackupHistory(filter),
        ]);
        setJobs(jobsData || []);
        setHistory(historyData || []);
      } catch (error) {
        console.error('History load error:', error);
      } finally {
        setLoading(false);
      }
    };
    loadAll();
  }, [filter]);

  const formatDate = (dateStr: string) => new Date(dateStr).toLocaleString(i18n.language === 'en' ? 'en-US' : 'tr-TR');
  const formatDuration = (start: string, end?: string) => {
    if (!end) return '-';
    const diff = Math.round((new Date(end).getTime() - new Date(start).getTime()) / 1000);
    if (diff < 60) return `${diff} ${t('common.seconds')}`;
    if (diff < 3600) return `${Math.floor(diff / 60)} ${t('common.minutes')}`;
    return `${Math.floor(diff / 3600)} ${t('common.hours')}`;
  };
  const formatSize = (bytes: number) => {
    if (!bytes || bytes <= 0 || !isFinite(bytes)) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), sizes.length - 1);
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };
  const getJobName = (id: number) => jobs.find((j: any) => j.id === id)?.name || t('history.jobFallback', { id });

  const successCount = history.filter((h: any) => h.status === 'success').length;
  const failedCount = history.filter((h: any) => h.status === 'failed').length;

  return (
    <div>
      <PageHeader title={t('history.title')} description={t('history.description')} />

      {/* Filters */}
      <Card className="border-zinc-800 bg-zinc-900 mb-6">
        <CardContent className="p-4">
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[180px] space-y-1.5">
              <Label className="text-xs">{t('history.job')}</Label>
              <Select value={filter.job_id} onChange={(e) => setFilter({ ...filter, job_id: parseInt(e.target.value) })}>
                <option value={0}>{t('history.allJobs')}</option>
                {jobs.map((job: any) => (<option key={job.id} value={job.id}>{job.name}</option>))}
              </Select>
            </div>
            <div className="flex-1 min-w-[180px] space-y-1.5">
              <Label className="text-xs">{t('common.status')}</Label>
              <Select value={filter.status} onChange={(e) => setFilter({ ...filter, status: e.target.value })}>
                <option value="">{t('history.allStatuses')}</option>
                <option value="success">{t('common.successful')}</option>
                <option value="failed">{t('common.failed')}</option>
                <option value="running">{t('common.running')}</option>
              </Select>
            </div>
            <div className="flex-1 min-w-[180px] space-y-1.5">
              <Label className="text-xs">{t('history.limit')}</Label>
              <Select value={filter.limit} onChange={(e) => setFilter({ ...filter, limit: parseInt(e.target.value) })}>
                <option value={10}>{t('history.last10')}</option>
                <option value={25}>{t('history.last25')}</option>
                <option value={50}>{t('history.last50')}</option>
                <option value={100}>{t('history.last100')}</option>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {loading ? (
        <div className="flex items-center justify-center py-16"><Spinner size="lg" /></div>
      ) : history.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900">
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="w-14 h-14 rounded-full bg-zinc-800 flex items-center justify-center mb-4">
              <HistoryIcon className="w-7 h-7 text-zinc-600" />
            </div>
            <p className="text-zinc-400 text-sm">{t('history.noHistory')}</p>
          </CardContent>
        </Card>
      ) : (
        <>
          <Card className="border-zinc-800 bg-zinc-900 p-0">
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead>{t('common.status')}</TableHead>
                  <TableHead>{t('history.job')}</TableHead>
                  <TableHead>{t('common.file')}</TableHead>
                  <TableHead>{t('common.size')}</TableHead>
                  <TableHead>{t('history.startTime')}</TableHead>
                  <TableHead>{t('history.duration')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {history.map((item: any) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Badge variant={item.status === 'success' ? 'success' : item.status === 'failed' ? 'destructive' : 'warning'}>
                        {item.status === 'success' ? (
                          <><CheckCircle2 className="w-3 h-3 mr-1" /> {t('common.successful')}</>
                        ) : item.status === 'failed' ? (
                          <><XCircle className="w-3 h-3 mr-1" /> {t('common.failed')}</>
                        ) : (
                          <><Loader2 className="w-3 h-3 mr-1 animate-spin" /> {t('common.running')}</>
                        )}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-medium text-white">{getJobName(item.job_id)}</TableCell>
                    <TableCell className="font-mono text-xs max-w-[200px] truncate">{item.file_name || '-'}</TableCell>
                    <TableCell>{formatSize(item.file_size)}</TableCell>
                    <TableCell className="text-xs">{formatDate(item.started_at)}</TableCell>
                    <TableCell>{formatDuration(item.started_at, item.completed_at)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          {/* Summary */}
          <div className="flex items-center gap-4 mt-4 text-xs text-zinc-500">
            <span>{t('history.totalRecords', { count: history.length })}</span>
            <span className="flex items-center gap-1">
              <CheckCircle2 className="w-3 h-3 text-emerald-500" /> {t('history.successCount', { count: successCount })}
            </span>
            <span className="flex items-center gap-1">
              <XCircle className="w-3 h-3 text-red-500" /> {t('history.failedCount', { count: failedCount })}
            </span>
          </div>
        </>
      )}
    </div>
  );
}
