'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import Image from 'next/image';
import {
  LayoutDashboard,
  Users,
  Plug,
  Briefcase,
  Settings,
  LogOut,
} from 'lucide-react';
import { useAuthStore } from '@/lib/store/auth';
import { useRouter } from 'next/navigation';
import { useActivePlugins } from '@/hooks/useActivePlugins';
import { PluginIcon } from './PluginIcon';
import { useI18n } from '@/lib/i18n/I18nProvider';
import { LanguageSwitcher } from '@/components/i18n/LanguageSwitcher';

const menuItems = [
  { href: '/admin', labelKey: 'sidebar.dashboard', icon: LayoutDashboard },
  { href: '/admin/users', labelKey: 'sidebar.users', icon: Users },
  { href: '/admin/plugins', labelKey: 'sidebar.plugins', icon: Plug },
  { href: '/admin/workers', labelKey: 'sidebar.workers', icon: Briefcase },
  { href: '/admin/settings', labelKey: 'sidebar.settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { logout } = useAuthStore();
  const { data: activePlugins = [] } = useActivePlugins();
  const { t } = useI18n();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  // Build dynamic plugin menu items
  const pluginMenuItems = activePlugins.flatMap((plugin) =>
    plugin.menu_items.map((item) => ({
      href: `/admin/plugins/${plugin.key}`,
      label: item.label,
      icon: item.icon,
      position: item.position,
      pluginKey: plugin.key,
    }))
  ).sort((a, b) => a.position - b.position);

  return (
    <aside className="w-64 bg-slate-900 text-white flex flex-col">
      <div className="p-6">
        <div className="flex flex-col items-start gap-3">
          <div className="h-10 w-10 relative">
            <Image src="/Heimdall.png" alt="Logo Heimdall" fill className="object-contain" sizes="40px" />
          </div>
          <h1 className="text-xl font-bold">{t('sidebar.brand')}</h1>
        </div>
      </div>

      <nav className="flex-1 px-4 overflow-y-auto">
        <ul className="space-y-1">
          {menuItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                    isActive
                      ? 'bg-primary-600 text-white'
                      : 'text-gray-300 hover:bg-slate-800'
                  }`}
                >
                  <item.icon className="h-5 w-5" />
                  {t(item.labelKey)}
                </Link>
              </li>
            );
          })}
        </ul>

        {/* Dynamic plugin menu items */}
        {pluginMenuItems.length > 0 && (
          <>
            <div className="mt-6 mb-2 px-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">
              {t('sidebar.extensions')}
            </div>
            <ul className="space-y-1">
              {pluginMenuItems.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <li key={item.href}>
                    <Link
                      href={item.href}
                      className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                        isActive
                          ? 'bg-primary-600 text-white'
                          : 'text-gray-300 hover:bg-slate-800'
                      }`}
                    >
                      <PluginIcon name={item.icon} />
                      {item.label}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </>
        )}
      </nav>

      <div className="p-4 border-t border-slate-700 space-y-1">
        <LanguageSwitcher variant="sidebar" />
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 px-4 py-3 w-full text-gray-300 hover:bg-slate-800 rounded-lg transition-colors"
        >
          <LogOut className="h-5 w-5" />
          {t('sidebar.logout')}
        </button>
      </div>
    </aside>
  );
}
