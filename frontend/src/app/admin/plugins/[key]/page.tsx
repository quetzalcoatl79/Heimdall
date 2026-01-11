'use client';

import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { pluginApi } from '@/lib/api/plugins';
import { apiClient } from '@/lib/api/client';
import { DynamicRenderer, ViewSchema } from '@/components/ui/DynamicRenderer';
import { WiFiPentestView } from '@/components/plugins/WiFiPentestView';
import { Plug, AlertCircle, RefreshCw } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

export default function PluginPage() {
  const params = useParams();
  const pluginKey = params.key as string;

  // Fetch plugin info
  const { data: plugins, isLoading: pluginsLoading } = useQuery({
    queryKey: ['plugins'],
    queryFn: () => pluginApi.list(),
  });

  const plugin = plugins?.plugins?.find(
    (p: any) => p.name.toLowerCase() === pluginKey.toLowerCase()
  );

  // Résultats de scan Wi-Fi
  const { data: scanResults, refetch: refetchScanResults } = useQuery({
    queryKey: ['plugin', pluginKey, 'scan-results'],
    queryFn: async () => {
      const res = await apiClient.get(`/plugins/${pluginKey}/scan/results`);
      return res.data;
    },
    enabled: pluginKey === 'wifi' && !!plugin?.enabled,
    refetchInterval: pluginKey === 'wifi' ? 10000 : false,
  });

  // Fetch view schema from plugin
  const { 
    data: viewSchema, 
    isLoading: viewLoading, 
    error: viewError,
    refetch: refetchView 
  } = useQuery({
    queryKey: ['plugin', pluginKey, 'view'],
    queryFn: async (): Promise<ViewSchema | null> => {
      try {
        const res = await apiClient.get(`/plugins/${pluginKey}/view`);
        return res.data;
      } catch (e: any) {
        // If view endpoint doesn't exist, return null (use default view)
        if (e.response?.status === 404) {
          return null;
        }
        throw e;
      }
    },
    enabled: !!plugin?.enabled,
    // react-query v5 passes a query object to refetchInterval
    refetchInterval: (query) => {
      const view = query.state.data as ViewSchema | null;
      // Auto-refresh if schema specifies it
      if (view?.refresh?.enabled) {
        return view.refresh.interval * 1000;
      }
      return false;
    },
  });

  // WiFi-specific: fetch interfaces and inject options into the select field
  const { data: wifiInterfaces } = useQuery({
    queryKey: ['plugin', pluginKey, 'interfaces'],
    queryFn: async () => {
      const res = await apiClient.get(`/plugins/${pluginKey}/interfaces`);
      return res.data?.interfaces ?? [];
    },
    enabled: pluginKey === 'wifi' && !!viewSchema,
    refetchInterval: 10000,
  });

  const schemaToRender = useMemo(() => {
    if (pluginKey !== 'wifi' || !viewSchema || !wifiInterfaces) {
      return viewSchema;
    }

    const options = [
      { value: '', label: 'Choisir', disabled: true },
      ...wifiInterfaces.map((iface: any) => ({
        value: iface.name,
        label: `${iface.name}${iface.monitor ? ' [monitor]' : ''}${iface.driver ? ` (${iface.driver})` : ''}`,
        disabled: iface.up === false,
      })),
    ];

    const patchComponents = (components: any[]): any[] =>
      components.map((component) => {
        if (component.type === 'form' && component.props?.id === 'wifi-select-iface') {
          const patchedFields = (component.props.fields || []).map((field: any) =>
            field.name === 'interface'
              ? { ...field, options, default: options.find((o) => !o.disabled)?.value || '' }
              : field
          );
          return { ...component, props: { ...component.props, fields: patchedFields } };
        }

        if (component.type === 'table' && component.id === 'wifi-networks') {
          return {
            ...component,
            props: {
              ...(component.props || {}),
              data: scanResults?.results || [],
            },
          };
        }

        if (component.children) {
          return { ...component, children: patchComponents(component.children) };
        }
        return component;
      });

    return { ...viewSchema, components: patchComponents(viewSchema.components) } as ViewSchema;
  }, [pluginKey, viewSchema, wifiInterfaces, scanResults]);

  const isLoading = pluginsLoading || viewLoading;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (!plugin) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <AlertCircle className="h-12 w-12 text-red-500 mb-4" />
        <h1 className="text-xl font-semibold text-gray-900">Plugin introuvable</h1>
        <p className="text-gray-500 mt-2">
          Le plugin &quot;{pluginKey}&quot; n&apos;existe pas ou n&apos;est pas activé.
        </p>
      </div>
    );
  }

  if (!plugin.enabled) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <AlertCircle className="h-12 w-12 text-yellow-500 mb-4" />
        <h1 className="text-xl font-semibold text-gray-900">Plugin désactivé</h1>
        <p className="text-gray-500 mt-2">
          Le plugin &quot;{plugin.name}&quot; est actuellement désactivé.
        </p>
      </div>
    );
  }

  // Use dedicated WiFi component for better UX
  if (pluginKey === 'wifi') {
    return <WiFiPentestView />;
  }

  // If plugin has a view schema, render it dynamically
  if (schemaToRender) {
    return (
      <div className="relative">
        {/* Refresh indicator */}
        {schemaToRender.refresh?.enabled && (
          <div className="absolute top-0 right-0 flex items-center gap-2 text-sm text-gray-500">
            <RefreshCw className="h-4 w-4 animate-spin-slow" />
            Auto-refresh: {schemaToRender.refresh.interval}s
          </div>
        )}
        <DynamicRenderer 
          schema={schemaToRender} 
          onAction={(actionId) => {
            if (pluginKey !== 'wifi') return;

            const getSelectedInterface = () => {
              const form = document.getElementById('wifi-select-iface') as HTMLFormElement | null;
              if (form) {
                const data = new FormData(form);
                const iface = (data.get('interface') as string) || '';
                if (iface) return iface;
              }
              const first = wifiInterfaces?.find((i: any) => !i.up === false) || wifiInterfaces?.[0];
              return first?.name || '';
            };

            const iface = getSelectedInterface();

            if (actionId === 'refresh') {
              refetchView();
              refetchScanResults();
              return;
            }

            if (actionId === 'scan') {
              apiClient
                .post(`/plugins/${pluginKey}/scan`, { interface: iface })
                .then(() => refetchScanResults())
                .catch(() => {});
            }
          }}
        />
      </div>
    );
  }

  // Default view if no schema is provided
  return <DefaultPluginView plugin={plugin} />;
}

// Default plugin view when no custom component exists
function DefaultPluginView({ plugin }: { plugin: any }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="p-3 bg-primary-100 rounded-xl">
          <Plug className="h-8 w-8 text-primary-600" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">{plugin.name}</h1>
          <p className="text-gray-500">Version {plugin.version}</p>
        </div>
      </div>

      <div className="card">
        <h2 className="font-semibold mb-2">Description</h2>
        <p className="text-gray-600">{plugin.description || 'Aucune description disponible.'}</p>
      </div>

      {plugin.manifest && (
        <div className="card">
          <h2 className="font-semibold mb-4">Manifest</h2>
          <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-auto">
            {JSON.stringify(plugin.manifest, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
