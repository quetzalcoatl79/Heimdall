'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { pluginApi } from '@/lib/api/plugins';
import { Power, Settings } from 'lucide-react';
import { useI18n } from '@/lib/i18n/I18nProvider';

export default function PluginsPage() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const { data, isLoading } = useQuery({
    queryKey: ['plugins'],
    queryFn: () => pluginApi.list(),
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      enabled ? pluginApi.disable(id) : pluginApi.enable(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plugins'] });
    },
  });

  if (isLoading) {
    return <div className="text-center py-8">{t('common.loading')}</div>;
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{t('admin.plugins.title')}</h1>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {data?.plugins?.map((plugin: any) => (
          <div key={plugin.id} className="card">
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="font-semibold text-lg">{plugin.name}</h3>
                <p className="text-sm text-gray-500">v{plugin.version}</p>
              </div>
              <span
                className={`px-2 py-1 text-xs font-medium rounded-full ${
                  plugin.enabled
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                {plugin.enabled ? t('admin.plugins.active') : t('admin.plugins.inactive')}
              </span>
            </div>

            <p className="text-gray-600 text-sm mb-4">
              {plugin.description || t('admin.plugins.noDescription')}
            </p>

            <div className="flex gap-2">
              <button
                onClick={() =>
                  toggleMutation.mutate({ id: plugin.id, enabled: plugin.enabled })
                }
                className={`btn flex items-center gap-2 ${
                  plugin.enabled ? 'btn-secondary' : 'btn-primary'
                }`}
              >
                <Power className="h-4 w-4" />
                {plugin.enabled ? t('admin.plugins.disable') : t('admin.plugins.enable')}
              </button>
              <button className="btn btn-secondary">
                <Settings className="h-4 w-4" />
              </button>
            </div>
          </div>
        ))}

        {(!data?.plugins || data.plugins.length === 0) && (
          <div className="col-span-full text-center py-8 text-gray-500">
            {t('admin.plugins.empty')}
          </div>
        )}
      </div>
    </div>
  );
}
