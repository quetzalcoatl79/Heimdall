import { apiClient } from './client';

interface Plugin {
  id: string;
  name: string;
  version: string;
  description: string;
  enabled: boolean;
  config: Record<string, any>;
  manifest: Record<string, any>;
  installed_at: string;
}

interface PluginListResponse {
  plugins: Plugin[];
}

interface PluginManifest {
  plugins: Array<{
    id: string;
    name: string;
    version: string;
    routes: Array<{
      path: string;
      component: string;
    }>;
    permissions: string[];
    menu_items: Array<{
      label: string;
      path: string;
      icon: string;
      position: number;
    }>;
  }>;
}

export const pluginApi = {
  list: async (): Promise<PluginListResponse> => {
    const response = await apiClient.get('/admin/plugins');
    return response.data;
  },

  get: async (id: string): Promise<Plugin> => {
    const response = await apiClient.get(`/admin/plugins/${id}`);
    return response.data;
  },

  enable: async (id: string): Promise<Plugin> => {
    const response = await apiClient.post(`/admin/plugins/${id}/enable`);
    return response.data;
  },

  disable: async (id: string): Promise<Plugin> => {
    const response = await apiClient.post(`/admin/plugins/${id}/disable`);
    return response.data;
  },

  getManifest: async (): Promise<PluginManifest> => {
    const response = await apiClient.get('/admin/plugins/manifest');
    return response.data;
  },
};
