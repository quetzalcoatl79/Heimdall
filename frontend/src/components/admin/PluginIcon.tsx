'use client';

import {
  Plug,
  Heart,
  Activity,
  Settings,
  Box,
  Zap,
  Database,
  Shield,
  Bell,
  BarChart,
  type LucideIcon,
} from 'lucide-react';

const iconMap: Record<string, LucideIcon> = {
  plug: Plug,
  heart: Heart,
  activity: Activity,
  settings: Settings,
  box: Box,
  zap: Zap,
  database: Database,
  shield: Shield,
  bell: Bell,
  chart: BarChart,
  healthcheck: Activity,
};

interface PluginIconProps {
  name: string;
  className?: string;
}

export function PluginIcon({ name, className = 'h-5 w-5' }: PluginIconProps) {
  const Icon = iconMap[name.toLowerCase()] || Plug;
  return <Icon className={className} />;
}
