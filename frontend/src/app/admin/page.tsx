'use client';

import { useQuery } from '@tanstack/react-query';
import { Users, Plug, Briefcase, Activity } from 'lucide-react';
import { userApi } from '@/lib/api/users';
import { pluginApi } from '@/lib/api/plugins';
import { useI18n } from '@/lib/i18n/I18nProvider';

export default function AdminDashboard() {
  const { t } = useI18n();

  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: () => userApi.list({ page: 1, page_size: 1 }),
  });

  const { data: pluginsData } = useQuery({
    queryKey: ['plugins'],
    queryFn: () => pluginApi.list(),
  });

  const stats = [
    {
      name: t('admin.dashboard.stats.users'),
      value: usersData?.total || 0,
      icon: Users,
      color: 'bg-blue-500',
    },
    {
      name: t('admin.dashboard.stats.plugins'),
      value: pluginsData?.plugins?.filter((p: any) => p.enabled).length || 0,
      icon: Plug,
      color: 'bg-green-500',
    },
    {
      name: t('admin.dashboard.stats.jobs'),
      value: 0,
      icon: Briefcase,
      color: 'bg-yellow-500',
    },
    {
      name: t('admin.dashboard.stats.status'),
      value: t('admin.dashboard.stats.statusOk'),
      icon: Activity,
      color: 'bg-emerald-500',
    },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">{t('admin.dashboard.title')}</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => (
          <div key={stat.name} className="card flex items-center gap-4">
            <div className={`p-3 rounded-lg ${stat.color}`}>
              <stat.icon className="h-6 w-6 text-white" />
            </div>
            <div>
              <p className="text-sm text-gray-600">{stat.name}</p>
              <p className="text-2xl font-bold">{stat.value}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">{t('admin.dashboard.recent')}</h2>
          <p className="text-gray-500">{t('admin.dashboard.noRecent')}</p>
        </div>

        <div className="card">
          <h2 className="text-lg font-semibold mb-4">{t('admin.dashboard.quick')}</h2>
          <div className="space-y-2">
            <button className="btn btn-secondary w-full text-left">
              {t('admin.dashboard.createUser')}
            </button>
            <button className="btn btn-secondary w-full text-left">
              {t('admin.dashboard.installPlugin')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
