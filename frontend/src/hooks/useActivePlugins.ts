'use client';

import { useQuery } from '@tanstack/react-query';
import { pluginApi } from '@/lib/api/plugins';

export interface PluginMenuItem {
  label: string;
  path: string;
  icon: string;
  position: number;
}

export interface ActivePlugin {
  id: string;
  key: string;
  name: string;
  version: string;
  description?: string;
  routes: Array<{
    path: string;
    component: string;
  }>;
  permissions: string[];
  menu_items: PluginMenuItem[];
}

export function useActivePlugins() {
  return useQuery({
    queryKey: ['plugins', 'manifest'],
    queryFn: async (): Promise<ActivePlugin[]> => {
      try {
        const manifest = await pluginApi.getManifest();
        return (manifest.plugins || []).map((p: any) => ({
          id: p.id,
          key: p.key || p.name?.toLowerCase(),
          name: p.name,
          version: p.version,
          description: p.description,
          routes: p.routes || [],
          permissions: p.permissions || [],
          menu_items: p.menu_items || [],
        }));
      } catch {
        return [];
      }
    },
    staleTime: 30_000, // 30 seconds
    refetchOnWindowFocus: false,
  });
}
