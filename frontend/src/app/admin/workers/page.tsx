'use client';

import { useQuery } from '@tanstack/react-query';
import { RefreshCw, Play, XCircle, Zap, Clock, CheckCircle, AlertCircle } from 'lucide-react';
import { apiClient } from '@/lib/api/client';
import { useI18n } from '@/lib/i18n/I18nProvider';

interface WorkerStats {
  active_workers: number;
  total_workers: number;
  queue_length: number;
  jobs_pending: number;
  jobs_running: number;
  jobs_completed: number;
  jobs_failed: number;
  workers: WorkerInfo[];
}

interface WorkerInfo {
  id: string;
  started_at: string;
  last_seen: string;
  jobs_handled: number;
  status: string;
}

interface Job {
  id: string;
  type: string;
  status: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
}

export default function WorkersPage() {
  const { t, locale } = useI18n();
  // Fetch worker stats
  const { data: stats, isLoading: statsLoading, refetch } = useQuery<WorkerStats>({
    queryKey: ['worker-stats'],
    queryFn: async () => {
      const response = await apiClient.get('/admin/workers/stats');
      return response.data;
    },
    refetchInterval: 5000, // Auto-refresh every 5 seconds
  });

  // Fetch recent jobs
  const { data: jobsData, isLoading: jobsLoading } = useQuery<{ jobs: Job[] }>({
    queryKey: ['jobs'],
    queryFn: async () => {
      const response = await apiClient.get('/admin/jobs?limit=50');
      return response.data;
    },
    refetchInterval: 5000,
  });

  const jobs = jobsData?.jobs || [];

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{t('admin.workers.title')}</h1>
        <button 
          onClick={() => refetch()}
          className="btn btn-secondary flex items-center gap-2"
        >
          <RefreshCw className="h-4 w-4" />
          {t('admin.workers.refresh')}
        </button>
      </div>

      {/* Worker Stats */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
        <div className="card">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <Zap className="h-5 w-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{t('admin.workers.stats.active')}</p>
              <p className="text-2xl font-bold text-green-600">
                {statsLoading ? '...' : stats?.active_workers || 0}
              </p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-yellow-100 rounded-lg">
              <Clock className="h-5 w-5 text-yellow-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{t('admin.workers.stats.pending')}</p>
              <p className="text-2xl font-bold text-yellow-600">
                {statsLoading ? '...' : (stats?.jobs_pending || 0) + (stats?.queue_length || 0)}
              </p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <RefreshCw className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{t('admin.workers.stats.running')}</p>
              <p className="text-2xl font-bold text-blue-600">
                {statsLoading ? '...' : stats?.jobs_running || 0}
              </p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <CheckCircle className="h-5 w-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{t('admin.workers.stats.completed')}</p>
              <p className="text-2xl font-bold text-green-600">
                {statsLoading ? '...' : stats?.jobs_completed || 0}
              </p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-red-100 rounded-lg">
              <AlertCircle className="h-5 w-5 text-red-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{t('admin.workers.stats.failed')}</p>
              <p className="text-2xl font-bold text-red-600">
                {statsLoading ? '...' : stats?.jobs_failed || 0}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Active Workers */}
      {stats?.workers && stats.workers.length > 0 && (
        <div className="card mb-6">
          <h2 className="font-semibold mb-4">{t('admin.workers.tableWorkers.title')}</h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t('admin.workers.tableWorkers.id')}</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t('admin.workers.tableWorkers.status')}</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t('admin.workers.tableWorkers.started')}</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t('admin.workers.tableWorkers.lastSeen')}</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t('admin.workers.tableWorkers.handled')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {stats.workers.map((worker) => (
                  <tr key={worker.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 text-sm font-mono">{worker.id.substring(0, 20)}...</td>
                    <td className="px-4 py-2">
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                        worker.status === 'running' 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {worker.status}
                      </span>
                    </td>
                    <td className="px-4 py-2 text-sm text-gray-500">
                      {new Date(worker.started_at).toLocaleString(locale)}
                    </td>
                    <td className="px-4 py-2 text-sm text-gray-500">
                      {new Date(worker.last_seen).toLocaleTimeString(locale)}
                    </td>
                    <td className="px-4 py-2 text-sm font-medium">{worker.jobs_handled}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Jobs table */}
      <div className="card overflow-hidden">
        <h2 className="font-semibold mb-4">{t('admin.workers.tableJobs.title')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  {t('admin.workers.tableJobs.id')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  {t('admin.workers.tableJobs.type')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  {t('admin.workers.tableJobs.status')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  {t('admin.workers.tableJobs.created')}
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                  {t('admin.workers.tableJobs.actions')}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {jobs.map((job) => (
                <tr key={job.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm font-mono">
                    {job.id.substring(0, 8)}
                  </td>
                  <td className="px-6 py-4">{job.type}</td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded-full ${
                        job.status === 'completed'
                          ? 'bg-green-100 text-green-800'
                          : job.status === 'failed'
                          ? 'bg-red-100 text-red-800'
                          : job.status === 'running'
                          ? 'bg-blue-100 text-blue-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {job.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {new Date(job.created_at).toLocaleString(locale)}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex justify-end gap-2">
                      {job.status === 'failed' && (
                        <button className="text-blue-600 hover:text-blue-800" title="Relancer">
                          <Play className="h-4 w-4" />
                        </button>
                      )}
                      {job.status === 'pending' && (
                        <button className="text-red-600 hover:text-red-800" title="Annuler">
                          <XCircle className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {jobs.length === 0 && !jobsLoading && (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                    {t('admin.workers.tableJobs.empty')}
                  </td>
                </tr>
              )}
              {jobsLoading && (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                    {t('common.loading')}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
