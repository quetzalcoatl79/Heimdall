'use client';

import { useEffect, useCallback, useRef, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { 
  Wifi, 
  WifiOff, 
  RefreshCw, 
  Radio, 
  AlertTriangle, 
  Shield,
  Play,
  Square,
  Loader2,
  CheckCircle,
  XCircle
} from 'lucide-react';

// ========================================
// UUID Generator (simple version to avoid dependency)
// ========================================

const generateUUID = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

// ========================================
// Types
// ========================================

interface WiFiInterface {
  name: string;
  mac: string;
  monitor: boolean;
  up: boolean;
  driver?: string;
}

interface WiFiState {
  interface: string;
  original_mode: string;
  current_mode: string;
  is_disabled: boolean;
  disabled_at: string;
  session_id: string;
}

interface WiFiStateResponse {
  states: WiFiState[];
  active_session: string;
  has_disabled: boolean;
}

interface WiFiNetwork {
  ssid: string;
  bssid: string;
  channel: number;
  signal: number;
  security: string;
  vendor?: string;
  wps?: boolean;
}

// ========================================
// Session Management
// ========================================

const getSessionId = (): string => {
  if (typeof window === 'undefined') return '';
  
  let sessionId = sessionStorage.getItem('wifi-session-id');
  if (!sessionId) {
    sessionId = generateUUID();
    sessionStorage.setItem('wifi-session-id', sessionId);
  }
  return sessionId;
};

// ========================================
// Component
// ========================================

export default function WifiView() {
  const queryClient = useQueryClient();
  const sessionId = useRef<string>(getSessionId());
  const heartbeatInterval = useRef<NodeJS.Timeout | null>(null);
  const [selectedInterface, setSelectedInterface] = useState<string>('');
  const [isScanning, setIsScanning] = useState(false);

  // ========================================
  // Queries
  // ========================================

  // Fetch WiFi interfaces
  const { data: interfacesData, refetch: refetchInterfaces } = useQuery({
    queryKey: ['wifi', 'interfaces'],
    queryFn: async () => {
      const res = await apiClient.get('/plugins/wifi/interfaces');
      return res.data as { interfaces: WiFiInterface[] };
    },
    refetchInterval: 10000,
  });

  // Fetch WiFi state
  const { data: stateData, refetch: refetchState } = useQuery({
    queryKey: ['wifi', 'state'],
    queryFn: async () => {
      const res = await apiClient.get('/plugins/wifi/state');
      return res.data as WiFiStateResponse;
    },
    refetchInterval: 5000,
  });

  // ========================================
  // Mutations
  // ========================================

  // Start scan
  const scanMutation = useMutation({
    mutationFn: async (iface: string) => {
      const res = await apiClient.post('/plugins/wifi/scan', {
        interface: iface,
        session_id: sessionId.current,
      });
      return res.data;
    },
    onSuccess: () => {
      setIsScanning(true);
      refetchState();
      startHeartbeat();
    },
    onError: (error) => {
      console.error('Scan failed:', error);
      handleRestoreWifi();
    },
  });

  // Restore WiFi
  const restoreMutation = useMutation({
    mutationFn: async (iface?: string) => {
      const res = await apiClient.post('/plugins/wifi/restore', {
        interface: iface || '',
      });
      return res.data;
    },
    onSuccess: () => {
      setIsScanning(false);
      stopHeartbeat();
      refetchState();
      refetchInterfaces();
      queryClient.invalidateQueries({ queryKey: ['wifi'] });
    },
  });

  // Heartbeat
  const sendHeartbeat = useCallback(async () => {
    try {
      await apiClient.post('/plugins/wifi/heartbeat', {
        session_id: sessionId.current,
      });
    } catch (error) {
      console.error('Heartbeat failed:', error);
    }
  }, []);

  // ========================================
  // Heartbeat Management
  // ========================================

  const startHeartbeat = useCallback(() => {
    if (heartbeatInterval.current) return;
    
    // Envoyer immédiatement un heartbeat
    sendHeartbeat();
    
    // Puis toutes les 10 secondes
    heartbeatInterval.current = setInterval(sendHeartbeat, 10000);
  }, [sendHeartbeat]);

  const stopHeartbeat = useCallback(() => {
    if (heartbeatInterval.current) {
      clearInterval(heartbeatInterval.current);
      heartbeatInterval.current = null;
    }
  }, []);

  // ========================================
  // Restore WiFi Handler
  // ========================================

  const handleRestoreWifi = useCallback(async () => {
    try {
      await restoreMutation.mutateAsync(undefined);
    } catch (error) {
      console.error('Failed to restore WiFi:', error);
    }
  }, [restoreMutation]);

  // ========================================
  // Page Unload Handler (browser close/navigate away)
  // ========================================

  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      // Si un scan est en cours, avertir l'utilisateur
      if (stateData?.has_disabled) {
        // Envoyer le signal de déconnexion via sendBeacon (fiable même en fermeture)
        const data = JSON.stringify({ session_id: sessionId.current });
        navigator.sendBeacon('/api/plugins/wifi/disconnect', data);
        
        // Optionnel: Afficher un message de confirmation
        event.preventDefault();
        event.returnValue = 'Un scan WiFi est en cours. Le WiFi sera restauré automatiquement.';
        return event.returnValue;
      }
    };

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'hidden' && stateData?.has_disabled) {
        // Page cachée/fermée, envoyer heartbeat de sécurité
        sendHeartbeat();
      }
    };

    const handleUnload = () => {
      // Dernier recours: envoyer signal de déconnexion
      if (stateData?.has_disabled) {
        const data = JSON.stringify({ session_id: sessionId.current });
        navigator.sendBeacon('/api/plugins/wifi/disconnect', data);
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    window.addEventListener('unload', handleUnload);
    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
      window.removeEventListener('unload', handleUnload);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [stateData?.has_disabled, sendHeartbeat]);

  // ========================================
  // Component Cleanup
  // ========================================

  useEffect(() => {
    // Démarrer le heartbeat si un scan est déjà en cours
    if (stateData?.has_disabled) {
      startHeartbeat();
    }

    return () => {
      stopHeartbeat();
    };
  }, [stateData?.has_disabled, startHeartbeat, stopHeartbeat]);

  // ========================================
  // Error Recovery
  // ========================================

  useEffect(() => {
    // Gestion des erreurs globales pour restaurer le WiFi
    const handleError = (event: ErrorEvent) => {
      if (stateData?.has_disabled) {
        console.error('Critical error detected, restoring WiFi...', event.error);
        handleRestoreWifi();
      }
    };

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      if (stateData?.has_disabled) {
        console.error('Unhandled rejection, restoring WiFi...', event.reason);
        // Ne pas restaurer immédiatement pour les erreurs réseau temporaires
        // mais logger pour investigation
      }
    };

    window.addEventListener('error', handleError);
    window.addEventListener('unhandledrejection', handleUnhandledRejection);

    return () => {
      window.removeEventListener('error', handleError);
      window.removeEventListener('unhandledrejection', handleUnhandledRejection);
    };
  }, [stateData?.has_disabled, handleRestoreWifi]);

  // ========================================
  // Render
  // ========================================

  const interfaces = interfacesData?.interfaces || [];
  const hasDisabledInterfaces = stateData?.has_disabled || false;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-3">
            <Wifi className="h-7 w-7 text-primary-600" />
            Pentest Wi-Fi
          </h1>
          <p className="text-gray-500 mt-1">
            Scan, capture de handshakes et bruteforce WPA/WEP
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => {
              refetchInterfaces();
              refetchState();
            }}
            className="btn btn-secondary"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Rafraîchir
          </button>
          
          {/* Bouton Restaurer WiFi - Toujours visible si interfaces désactivées */}
          {hasDisabledInterfaces && (
            <button
              onClick={handleRestoreWifi}
              disabled={restoreMutation.isPending}
              className="btn btn-danger animate-pulse"
            >
              {restoreMutation.isPending ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <WifiOff className="h-4 w-4 mr-2" />
              )}
              🔄 Restaurer WiFi
            </button>
          )}
        </div>
      </div>

      {/* Alerte WiFi désactivé */}
      {hasDisabledInterfaces && (
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 rounded-r-lg">
          <div className="flex items-start">
            <AlertTriangle className="h-5 w-5 text-yellow-400 mr-3 mt-0.5" />
            <div>
              <h3 className="text-sm font-medium text-yellow-800">
                ⚠️ Mode Pentest actif
              </h3>
              <p className="mt-1 text-sm text-yellow-700">
                Le WiFi standard est désactivé pour permettre le scan.
                Cliquez sur <strong>"Restaurer WiFi"</strong> pour le réactiver.
              </p>
              <p className="mt-2 text-xs text-yellow-600">
                Le WiFi sera automatiquement restauré si vous fermez cette page ou en cas d'erreur.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Interface Selection */}
      <div className="card">
        <h2 className="font-semibold mb-4 flex items-center gap-2">
          <Radio className="h-5 w-5" />
          Interface Wi-Fi
        </h2>
        
        {interfaces.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <WifiOff className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>Aucune interface WiFi détectée</p>
            <p className="text-sm mt-1">
              Connectez un adaptateur WiFi compatible mode monitor
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {interfaces.map((iface) => (
              <div
                key={iface.name}
                onClick={() => setSelectedInterface(iface.name)}
                className={`p-4 border rounded-lg cursor-pointer transition-all ${
                  selectedInterface === iface.name
                    ? 'border-primary-500 bg-primary-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${
                      iface.up ? 'bg-green-100' : 'bg-gray-100'
                    }`}>
                      {iface.up ? (
                        <Wifi className="h-5 w-5 text-green-600" />
                      ) : (
                        <WifiOff className="h-5 w-5 text-gray-400" />
                      )}
                    </div>
                    <div>
                      <p className="font-medium">{iface.name}</p>
                      <p className="text-sm text-gray-500">{iface.mac}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {iface.monitor && (
                      <span className="px-2 py-1 text-xs font-medium bg-purple-100 text-purple-700 rounded">
                        Monitor
                      </span>
                    )}
                    {iface.driver && (
                      <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded">
                        {iface.driver}
                      </span>
                    )}
                    {iface.up ? (
                      <CheckCircle className="h-5 w-5 text-green-500" />
                    ) : (
                      <XCircle className="h-5 w-5 text-gray-400" />
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Actions */}
        {selectedInterface && (
          <div className="mt-4 flex gap-2">
            <button
              onClick={() => scanMutation.mutate(selectedInterface)}
              disabled={scanMutation.isPending || isScanning}
              className="btn btn-primary"
            >
              {scanMutation.isPending ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : isScanning ? (
                <Square className="h-4 w-4 mr-2" />
              ) : (
                <Play className="h-4 w-4 mr-2" />
              )}
              {isScanning ? 'Scan en cours...' : 'Démarrer le scan'}
            </button>
            
            {isScanning && (
              <button
                onClick={handleRestoreWifi}
                className="btn btn-secondary"
              >
                <Square className="h-4 w-4 mr-2" />
                Arrêter
              </button>
            )}
          </div>
        )}
      </div>

      {/* Networks Table */}
      <div className="card">
        <h2 className="font-semibold mb-4 flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Réseaux Wi-Fi détectés
        </h2>
        
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  SSID
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  BSSID
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Canal
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Signal
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Sécurité
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isScanning ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center">
                    <Loader2 className="h-8 w-8 mx-auto mb-2 animate-spin text-primary-600" />
                    <p className="text-gray-500">Scan en cours...</p>
                  </td>
                </tr>
              ) : (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-gray-500">
                    Aucun réseau détecté. Lancez un scan pour découvrir les réseaux.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Status Card */}
      <div className="card">
        <h2 className="font-semibold mb-4">État du plugin</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-xs text-gray-500 uppercase">Session</p>
            <p className="font-mono text-sm truncate">{sessionId.current.slice(0, 8)}...</p>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-xs text-gray-500 uppercase">Interfaces</p>
            <p className="font-semibold">{interfaces.length}</p>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-xs text-gray-500 uppercase">Mode</p>
            <p className={`font-semibold ${hasDisabledInterfaces ? 'text-yellow-600' : 'text-green-600'}`}>
              {hasDisabledInterfaces ? 'Pentest' : 'Normal'}
            </p>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-xs text-gray-500 uppercase">Heartbeat</p>
            <p className={`font-semibold ${heartbeatInterval.current ? 'text-green-600' : 'text-gray-400'}`}>
              {heartbeatInterval.current ? 'Actif' : 'Inactif'}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
