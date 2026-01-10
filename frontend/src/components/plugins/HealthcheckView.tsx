'use client';

import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { Activity, Database, Server, Clock, CheckCircle, XCircle, RefreshCw } from 'lucide-react';
import { useState, useEffect } from 'react';

interface HealthData {
  plugin: string;
  status: string;
  timestamp: string;
  started_at: string;
  uptime_seconds: number;
}

interface ReadyData {
  plugin: string;
  status: string;
  timestamp: string;
  checks: Record<string, string>;
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  const parts = [];
  if (days > 0) parts.push(`${days}j`);
  if (hours > 0) parts.push(`${hours}h`);
  if (mins > 0) parts.push(`${mins}m`);
  parts.push(`${secs}s`);

  return parts.join(' ');
}

export function HealthcheckView() {
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  const healthQuery = useQuery({
    queryKey: ['plugin', 'healthcheck', 'health'],
    queryFn: async (): Promise<HealthData> => {
      const res = await apiClient.get('/plugins/healthcheck/health');
      return res.data;
    },
    refetchInterval: autoRefresh ? 5000 : false,
  });

  const readyQuery = useQuery({
    queryKey: ['plugin', 'healthcheck', 'ready'],
    queryFn: async (): Promise<ReadyData> => {
      const res = await apiClient.get('/plugins/healthcheck/ready');
      return res.data;
    },
    refetchInterval: autoRefresh ? 5000 : false,
  });

  useEffect(() => {
    if (healthQuery.dataUpdatedAt) {
      setLastRefresh(new Date(healthQuery.dataUpdatedAt));
    }
  }, [healthQuery.dataUpdatedAt]);

  const health = healthQuery.data;
  const ready = readyQuery.data;
  const isLoading = healthQuery.isLoading || readyQuery.isLoading;
  const hasError = healthQuery.isError || readyQuery.isError;

  const overallStatus = ready?.status === 'ready' && health?.status === 'ok';

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-3">
            <Activity className="h-7 w-7 text-primary-600" />
            État de santé
          </h1>
          <p className="text-gray-500 mt-1">Surveillance en temps réel de l'application</p>
        </div>
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded border-gray-300"
            />
            Auto-refresh (5s)
          </label>
          <button
            onClick={() => {
              healthQuery.refetch();
              readyQuery.refetch();
            }}
            className="btn btn-secondary flex items-center gap-2"
            disabled={isLoading}
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            Rafraîchir
          </button>
        </div>
      </div>

      {/* Overall Status Banner */}
      <div
        className={`p-6 rounded-xl flex items-center gap-4 ${
          hasError
            ? 'bg-red-50 border border-red-200'
            : overallStatus
            ? 'bg-green-50 border border-green-200'
            : 'bg-yellow-50 border border-yellow-200'
        }`}
      >
        {hasError ? (
          <XCircle className="h-10 w-10 text-red-500" />
        ) : overallStatus ? (
          <CheckCircle className="h-10 w-10 text-green-500" />
        ) : (
          <Activity className="h-10 w-10 text-yellow-500" />
        )}
        <div>
          <h2 className="text-xl font-semibold">
            {hasError
              ? 'Erreur de connexion'
              : overallStatus
              ? 'Tous les systèmes sont opérationnels'
              : 'Certains services ne sont pas disponibles'}
          </h2>
          <p className="text-sm text-gray-600">
            Dernière vérification : {lastRefresh.toLocaleTimeString()}
          </p>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Uptime */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Clock className="h-5 w-5 text-blue-600" />
            </div>
            <h3 className="font-semibold">Uptime</h3>
          </div>
          <p className="text-3xl font-bold text-gray-900">
            {health ? formatUptime(health.uptime_seconds) : '--'}
          </p>
          <p className="text-sm text-gray-500 mt-1">
            Démarré le {health ? new Date(health.started_at).toLocaleString() : '--'}
          </p>
        </div>

        {/* Database Status */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <div
              className={`p-2 rounded-lg ${
                ready?.checks?.database === 'healthy' ? 'bg-green-100' : 'bg-red-100'
              }`}
            >
              <Database
                className={`h-5 w-5 ${
                  ready?.checks?.database === 'healthy' ? 'text-green-600' : 'text-red-600'
                }`}
              />
            </div>
            <h3 className="font-semibold">Base de données</h3>
          </div>
          <div className="flex items-center gap-2">
            {ready?.checks?.database === 'healthy' ? (
              <>
                <CheckCircle className="h-6 w-6 text-green-500" />
                <span className="text-lg font-medium text-green-700">Connectée</span>
              </>
            ) : (
              <>
                <XCircle className="h-6 w-6 text-red-500" />
                <span className="text-lg font-medium text-red-700">
                  {ready?.checks?.database || 'Non disponible'}
                </span>
              </>
            )}
          </div>
        </div>

        {/* Redis Status */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <div
              className={`p-2 rounded-lg ${
                ready?.checks?.redis === 'healthy' ? 'bg-green-100' : 'bg-red-100'
              }`}
            >
              <Server
                className={`h-5 w-5 ${
                  ready?.checks?.redis === 'healthy' ? 'text-green-600' : 'text-red-600'
                }`}
              />
            </div>
            <h3 className="font-semibold">Redis (Cache)</h3>
          </div>
          <div className="flex items-center gap-2">
            {ready?.checks?.redis === 'healthy' ? (
              <>
                <CheckCircle className="h-6 w-6 text-green-500" />
                <span className="text-lg font-medium text-green-700">Connecté</span>
              </>
            ) : (
              <>
                <XCircle className="h-6 w-6 text-red-500" />
                <span className="text-lg font-medium text-red-700">
                  {ready?.checks?.redis || 'Non disponible'}
                </span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Raw Data (for debugging) */}
      <details className="card">
        <summary className="cursor-pointer font-semibold text-gray-700">
          Données brutes (JSON)
        </summary>
        <div className="mt-4 space-y-4">
          <div>
            <h4 className="text-sm font-medium text-gray-500 mb-2">/health</h4>
            <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-auto">
              {JSON.stringify(health, null, 2)}
            </pre>
          </div>
          <div>
            <h4 className="text-sm font-medium text-gray-500 mb-2">/ready</h4>
            <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-auto">
              {JSON.stringify(ready, null, 2)}
            </pre>
          </div>
        </div>
      </details>
    </div>
  );
}
