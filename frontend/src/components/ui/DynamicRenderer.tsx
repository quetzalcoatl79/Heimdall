'use client';

import { ReactNode } from 'react';
import {
  AlertCircle,
  CheckCircle,
  Info,
  AlertTriangle,
  Clock,
  Database,
  Server,
  Activity,
  Plug,
  Settings,
  Users,
  Zap,
  Box,
  Download,
  RefreshCw,
  Upload,
  Trash2,
  Edit,
  Eye,
  Plus,
  Minus,
  Search,
  Filter,
  MoreHorizontal,
  Mail,
  Phone,
  MapPin,
  Calendar,
  FileText,
  Folder,
  Home,
  BarChart3,
  PieChart as PieChartIcon,
  TrendingUp,
  type LucideIcon,
} from 'lucide-react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  AreaChart,
  Area,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';

// ----- Types -----

export interface ViewSchema {
  title?: string;
  description?: string;
  icon?: string;
  actions?: Action[];
  components: UIComponent[];
  refresh?: RefreshConfig;
  meta?: Record<string, any>;
}

export interface RefreshConfig {
  enabled: boolean;
  interval: number; // seconds
}

export interface Action {
  id: string;
  label: string;
  icon?: string;
  variant?: 'primary' | 'secondary' | 'danger';
  href?: string;
  onClick?: string;
  disabled?: boolean;
}

export interface UIComponent {
  type: string;
  id?: string;
  props?: Record<string, any>;
  children?: UIComponent[];
  dataSource?: DataSource;
  visible?: boolean;
  className?: string;
}

export interface DataSource {
  url: string;
  method?: string;
  refreshOn?: string;
}

// ----- Icon Mapping -----

const iconMap: Record<string, LucideIcon> = {
  clock: Clock,
  database: Database,
  server: Server,
  activity: Activity,
  plug: Plug,
  settings: Settings,
  users: Users,
  zap: Zap,
  box: Box,
  info: Info,
  success: CheckCircle,
  warning: AlertTriangle,
  error: AlertCircle,
  check: CheckCircle,
  download: Download,
  refresh: RefreshCw,
  upload: Upload,
  trash: Trash2,
  edit: Edit,
  eye: Eye,
  plus: Plus,
  minus: Minus,
  search: Search,
  filter: Filter,
  more: MoreHorizontal,
  mail: Mail,
  phone: Phone,
  location: MapPin,
  calendar: Calendar,
  file: FileText,
  folder: Folder,
  home: Home,
};

export function getIcon(name: string): LucideIcon {
  return iconMap[name.toLowerCase()] || Box;
}

// ----- Component Registry -----

interface RendererProps {
  component: UIComponent;
}

const componentRegistry: Record<string, React.FC<RendererProps>> = {};

export function registerComponent(type: string, component: React.FC<RendererProps>) {
  componentRegistry[type] = component;
}

// ----- Dynamic Renderer -----

interface DynamicRendererProps {
  schema: ViewSchema;
  onAction?: (actionId: string) => void;
}

export function DynamicRenderer({ schema, onAction }: DynamicRendererProps) {
  const IconComponent = schema.icon ? getIcon(schema.icon) : null;

  return (
    <div className="space-y-6">
      {/* Header */}
      {(schema.title || schema.actions?.length) && (
        <div className="flex justify-between items-start">
          <div>
            {schema.title && (
              <h1 className="text-2xl font-bold flex items-center gap-3">
                {IconComponent && <IconComponent className="h-7 w-7 text-primary-600" />}
                {schema.title}
              </h1>
            )}
            {schema.description && (
              <p className="text-gray-500 mt-1">{schema.description}</p>
            )}
          </div>
          {schema.actions && schema.actions.length > 0 && (
            <div className="flex gap-2">
              {schema.actions.map((action) => (
                <button
                  key={action.id}
                  onClick={() => onAction?.(action.id)}
                  disabled={action.disabled}
                  className={`btn ${
                    action.variant === 'primary'
                      ? 'btn-primary'
                      : action.variant === 'danger'
                      ? 'btn-danger'
                      : 'btn-secondary'
                  }`}
                >
                  {action.icon && (() => {
                    const Icon = getIcon(action.icon);
                    return <Icon className="h-4 w-4 mr-2" />;
                  })()}
                  {action.label}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Components */}
      {schema.components.map((component, index) => (
        <ComponentRenderer key={component.id || index} component={component} />
      ))}
    </div>
  );
}

// ----- Component Renderer -----

export function ComponentRenderer({ component }: { component: UIComponent }) {
  if (component.visible === false) return null;

  // Check for custom registered component
  const CustomComponent = componentRegistry[component.type];
  if (CustomComponent) {
    return <CustomComponent component={component} />;
  }

  // Built-in components
  switch (component.type) {
    case 'card':
      return <CardComponent component={component} />;
    case 'grid':
      return <GridComponent component={component} />;
    case 'row':
      return <RowComponent component={component} />;
    case 'col':
      return <ColComponent component={component} />;
    case 'stats':
      return <StatsComponent component={component} />;
    case 'stat':
      return <StatComponent component={component} />;
    case 'alert':
      return <AlertComponent component={component} />;
    case 'table':
      return <TableComponent component={component} />;
    case 'text':
      return <TextComponent component={component} />;
    case 'heading':
      return <HeadingComponent component={component} />;
    case 'badge':
      return <BadgeComponent component={component} />;
    case 'progress':
      return <ProgressComponent component={component} />;
    case 'list':
      return <ListComponent component={component} />;
    case 'listItem':
      return <ListItemComponent component={component} />;
    case 'divider':
      return <hr className="my-4 border-gray-200" />;
    case 'empty':
      return <EmptyComponent component={component} />;
    case 'json':
      return <JSONComponent component={component} />;
    case 'codeBlock':
      return <CodeBlockComponent component={component} />;
    case 'form':
      return <FormComponent component={component} />;
    case 'chart':
      return <ChartComponent component={component} />;
    default:
      return (
        <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg text-yellow-700">
          Unknown component type: {component.type}
        </div>
      );
  }
}

// ----- Built-in Components -----

function CardComponent({ component }: RendererProps) {
  const { title, subtitle, footer } = component.props || {};
  return (
    <div className={`card ${component.className || ''}`}>
      {(title || subtitle) && (
        <div className="mb-4">
          {title && <h3 className="font-semibold text-lg">{title}</h3>}
          {subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}
        </div>
      )}
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
      {footer && <div className="mt-4 pt-4 border-t border-gray-100 text-sm text-gray-500">{footer}</div>}
    </div>
  );
}

function GridComponent({ component }: RendererProps) {
  const cols = component.props?.cols || 3;
  const gap = component.props?.gap || 6;
  return (
    <div className={`grid grid-cols-1 md:grid-cols-${cols} gap-${gap} ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function RowComponent({ component }: RendererProps) {
  return (
    <div className={`flex flex-wrap gap-4 ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function ColComponent({ component }: RendererProps) {
  const span = component.props?.span || 1;
  return (
    <div className={`col-span-${span} ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function StatsComponent({ component }: RendererProps) {
  return (
    <div className={`grid grid-cols-1 md:grid-cols-3 gap-6 ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function StatComponent({ component }: RendererProps) {
  const { label, value, icon, color, trend } = component.props || {};
  const Icon = icon ? getIcon(icon) : null;
  
  const colorClasses: Record<string, string> = {
    blue: 'bg-blue-100 text-blue-600',
    green: 'bg-green-100 text-green-600',
    red: 'bg-red-100 text-red-600',
    yellow: 'bg-yellow-100 text-yellow-600',
    gray: 'bg-gray-100 text-gray-600',
    purple: 'bg-purple-100 text-purple-600',
    indigo: 'bg-indigo-100 text-indigo-600',
    pink: 'bg-pink-100 text-pink-600',
    orange: 'bg-orange-100 text-orange-600',
    cyan: 'bg-cyan-100 text-cyan-600',
  };

  return (
    <div className="flex items-start gap-4">
      {Icon && (
        <div className={`p-2 rounded-lg ${colorClasses[color] || colorClasses.gray}`}>
          <Icon className="h-5 w-5" />
        </div>
      )}
      <div>
        <p className="text-sm text-gray-500">{label}</p>
        <p className="text-2xl font-bold text-gray-900">{value}</p>
        {trend && (
          <p className={`text-sm ${trend.value >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            {trend.value >= 0 ? '+' : ''}{trend.value}% {trend.label}
          </p>
        )}
      </div>
    </div>
  );
}

function AlertComponent({ component }: RendererProps) {
  const { variant, message, title } = component.props || {};
  
  const variants: Record<string, { bg: string; border: string; text: string; icon: LucideIcon }> = {
    info: { bg: 'bg-blue-50', border: 'border-blue-200', text: 'text-blue-800', icon: Info },
    success: { bg: 'bg-green-50', border: 'border-green-200', text: 'text-green-800', icon: CheckCircle },
    warning: { bg: 'bg-yellow-50', border: 'border-yellow-200', text: 'text-yellow-800', icon: AlertTriangle },
    error: { bg: 'bg-red-50', border: 'border-red-200', text: 'text-red-800', icon: AlertCircle },
  };

  const style = variants[variant] || variants.info;
  const Icon = style.icon;

  return (
    <div className={`p-4 rounded-lg border ${style.bg} ${style.border} ${style.text} flex items-start gap-3`}>
      <Icon className="h-5 w-5 mt-0.5" />
      <div>
        {title && <p className="font-semibold">{title}</p>}
        <p>{message}</p>
      </div>
    </div>
  );
}

function TableComponent({ component }: RendererProps) {
  const { columns, data } = component.props || {};
  
  if (!columns || !data) {
    return <div className="text-gray-500">No data available</div>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {columns.map((col: any) => (
              <th
                key={col.key}
                className={`px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                  col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                }`}
                style={{ width: col.width }}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {data.map((row: any, i: number) => (
            <tr key={i}>
              {columns.map((col: any) => (
                <td key={col.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {renderCellValue(row[col.key], col.render)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function renderCellValue(value: any, render?: string) {
  if (render === 'badge') {
    const variant = value === 'healthy' || value === 'active' ? 'green' : 'red';
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full bg-${variant}-100 text-${variant}-800`}>
        {value}
      </span>
    );
  }
  if (render === 'date' && value) {
    return new Date(value).toLocaleDateString();
  }
  return value;
}

function TextComponent({ component }: RendererProps) {
  return <p className={`text-gray-600 ${component.className || ''}`}>{component.props?.content}</p>;
}

function HeadingComponent({ component }: RendererProps) {
  const { level, content } = component.props || {};
  const Tag = `h${level || 2}` as keyof JSX.IntrinsicElements;
  const sizes: Record<number, string> = {
    1: 'text-3xl',
    2: 'text-2xl',
    3: 'text-xl',
    4: 'text-lg',
    5: 'text-base',
    6: 'text-sm',
  };
  return <Tag className={`font-bold ${sizes[level] || sizes[2]} ${component.className || ''}`}>{content}</Tag>;
}

function BadgeComponent({ component }: RendererProps) {
  const { text, variant } = component.props || {};
  const variants: Record<string, string> = {
    primary: 'bg-primary-100 text-primary-800',
    success: 'bg-green-100 text-green-800',
    warning: 'bg-yellow-100 text-yellow-800',
    danger: 'bg-red-100 text-red-800',
    info: 'bg-blue-100 text-blue-800',
    gray: 'bg-gray-100 text-gray-800',
    purple: 'bg-purple-100 text-purple-800',
    indigo: 'bg-indigo-100 text-indigo-800',
    pink: 'bg-pink-100 text-pink-800',
    orange: 'bg-orange-100 text-orange-800',
    cyan: 'bg-cyan-100 text-cyan-800',
  };
  return (
    <span className={`px-2 py-1 text-xs font-medium rounded-full ${variants[variant] || variants.gray}`}>
      {text}
    </span>
  );
}

function ProgressComponent({ component }: RendererProps) {
  const { value, max, color } = component.props || {};
  const percentage = Math.min(100, Math.max(0, (value / (max || 100)) * 100));
  const colorClasses: Record<string, string> = {
    primary: 'bg-primary-600',
    blue: 'bg-blue-600',
    green: 'bg-green-600',
    red: 'bg-red-600',
    yellow: 'bg-yellow-600',
    purple: 'bg-purple-600',
    indigo: 'bg-indigo-600',
    orange: 'bg-orange-600',
    pink: 'bg-pink-600',
  };
  return (
    <div className="w-full bg-gray-200 rounded-full h-2.5">
      <div className={`${colorClasses[color] || colorClasses.primary} h-2.5 rounded-full`} style={{ width: `${percentage}%` }} />
    </div>
  );
}

function ListComponent({ component }: RendererProps) {
  return (
    <ul className={`space-y-2 ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <li key={child.id || i}>
          <ComponentRenderer component={child} />
        </li>
      ))}
    </ul>
  );
}

function ListItemComponent({ component }: RendererProps) {
  const { title, subtitle } = component.props || {};
  return (
    <div className="flex justify-between items-center py-2">
      <span className="text-sm text-gray-500">{title}</span>
      <span className="text-sm font-medium text-gray-900">{subtitle}</span>
    </div>
  );
}

function EmptyComponent({ component }: RendererProps) {
  const { message, icon } = component.props || {};
  const Icon = icon ? getIcon(icon) : Box;
  return (
    <div className="text-center py-12">
      <Icon className="mx-auto h-12 w-12 text-gray-400" />
      <p className="mt-4 text-gray-500">{message}</p>
    </div>
  );
}

function JSONComponent({ component }: RendererProps) {
  return (
    <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-auto">
      {JSON.stringify(component.props?.data, null, 2)}
    </pre>
  );
}

function CodeBlockComponent({ component }: RendererProps) {
  const { code, language } = component.props || {};
  return (
    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-auto">
      <code className={`language-${language || 'text'}`}>{code}</code>
    </pre>
  );
}

function FormComponent({ component }: RendererProps) {
  const { id, fields, submitLabel, submitUrl, method } = component.props || {};
  
  return (
    <form id={id} className="space-y-4">
      {fields?.map((field: any) => (
        <FormFieldRenderer key={field.name} field={field} />
      ))}
      {submitUrl && (
        <button type="submit" className="btn btn-primary">
          {submitLabel || 'Submit'}
        </button>
      )}
    </form>
  );
}

function FormFieldRenderer({ field }: { field: any }) {
  const baseInputClass = "mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm";
  
  return (
    <div className={field.colSpan ? `col-span-${field.colSpan}` : ''}>
      {field.type !== 'hidden' && (
        <label htmlFor={field.name} className="block text-sm font-medium text-gray-700">
          {field.label}
          {field.required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      
      {field.type === 'select' ? (
        <select
          id={field.name}
          name={field.name}
          defaultValue={field.default}
          required={field.required}
          disabled={field.disabled}
          className={baseInputClass}
        >
          {field.options?.map((opt: any) => (
            <option key={opt.value} value={opt.value} disabled={opt.disabled}>
              {opt.label}
            </option>
          ))}
        </select>
      ) : field.type === 'textarea' ? (
        <textarea
          id={field.name}
          name={field.name}
          placeholder={field.placeholder}
          defaultValue={field.default}
          required={field.required}
          disabled={field.disabled}
          readOnly={field.readOnly}
          className={baseInputClass}
          rows={4}
        />
      ) : field.type === 'checkbox' ? (
        <input
          type="checkbox"
          id={field.name}
          name={field.name}
          defaultChecked={field.default}
          disabled={field.disabled}
          className="mt-1 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        />
      ) : (
        <input
          type={field.type}
          id={field.name}
          name={field.name}
          placeholder={field.placeholder}
          defaultValue={field.default}
          required={field.required}
          disabled={field.disabled}
          readOnly={field.readOnly}
          className={baseInputClass}
        />
      )}
      
      {field.help && <p className="mt-1 text-sm text-gray-500">{field.help}</p>}
    </div>
  );
}

// ----- Chart Component -----

interface ChartSeries {
  name: string;
  data: number[];
  color?: string;
}

const CHART_COLORS = [
  '#3b82f6', // blue
  '#10b981', // green
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#f97316', // orange
  '#84cc16', // lime
  '#6366f1', // indigo
];

function ChartComponent({ component }: RendererProps) {
  const { 
    type = 'line', 
    labels = [], 
    series = [], 
    showLegend = true, 
    showGrid = true,
    stacked = false,
    height = 300,
    title
  } = component.props || {};

  // Transform data for recharts format
  const chartData = labels.map((label: string, index: number) => {
    const dataPoint: Record<string, any> = { name: label };
    series.forEach((s: ChartSeries) => {
      dataPoint[s.name] = s.data[index] ?? 0;
    });
    return dataPoint;
  });

  // For pie/donut charts, transform differently
  const pieData = series.map((s: ChartSeries, index: number) => ({
    name: s.name,
    value: s.data[0] ?? 0,
    color: s.color || CHART_COLORS[index % CHART_COLORS.length],
  }));

  const renderChart = () => {
    switch (type) {
      case 'line':
        return (
          <LineChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />}
            <XAxis dataKey="name" tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <YAxis tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
            {series.map((s: ChartSeries, index: number) => (
              <Line
                key={s.name}
                type="monotone"
                dataKey={s.name}
                stroke={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                strokeWidth={2}
                dot={{ fill: s.color || CHART_COLORS[index % CHART_COLORS.length], strokeWidth: 2 }}
                activeDot={{ r: 6 }}
              />
            ))}
          </LineChart>
        );

      case 'bar':
        return (
          <BarChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />}
            <XAxis dataKey="name" tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <YAxis tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
            {series.map((s: ChartSeries, index: number) => (
              <Bar
                key={s.name}
                dataKey={s.name}
                fill={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                stackId={stacked ? 'stack' : undefined}
                radius={[4, 4, 0, 0]}
              />
            ))}
          </BarChart>
        );

      case 'area':
        return (
          <AreaChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />}
            <XAxis dataKey="name" tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <YAxis tick={{ fill: '#6b7280' }} axisLine={{ stroke: '#e5e7eb' }} />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
            {series.map((s: ChartSeries, index: number) => (
              <Area
                key={s.name}
                type="monotone"
                dataKey={s.name}
                stroke={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                fill={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                fillOpacity={0.3}
                stackId={stacked ? 'stack' : undefined}
              />
            ))}
          </AreaChart>
        );

      case 'pie':
        return (
          <PieChart>
            <Pie
              data={pieData}
              cx="50%"
              cy="50%"
              outerRadius={height / 3}
              dataKey="value"
              label={({ name, percent }) => {
                const pct = percent ?? 0;
                return `${name} ${(pct * 100).toFixed(0)}%`;
              }}
              labelLine={{ stroke: '#6b7280' }}
            >
              {pieData.map((entry: any, index: number) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
          </PieChart>
        );

      case 'donut':
        return (
          <PieChart>
            <Pie
              data={pieData}
              cx="50%"
              cy="50%"
              innerRadius={height / 5}
              outerRadius={height / 3}
              dataKey="value"
              label={({ name, percent }) => {
                const pct = percent ?? 0;
                return `${name} ${(pct * 100).toFixed(0)}%`;
              }}
              labelLine={{ stroke: '#6b7280' }}
            >
              {pieData.map((entry: any, index: number) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
          </PieChart>
        );

      case 'radar':
        return (
          <RadarChart data={chartData} cx="50%" cy="50%" outerRadius={height / 3}>
            <PolarGrid stroke="#e5e7eb" />
            <PolarAngleAxis dataKey="name" tick={{ fill: '#6b7280' }} />
            <PolarRadiusAxis tick={{ fill: '#6b7280' }} />
            {series.map((s: ChartSeries, index: number) => (
              <Radar
                key={s.name}
                name={s.name}
                dataKey={s.name}
                stroke={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                fill={s.color || CHART_COLORS[index % CHART_COLORS.length]}
                fillOpacity={0.3}
              />
            ))}
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#fff', 
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
              }} 
            />
            {showLegend && <Legend />}
          </RadarChart>
        );

      default:
        return (
          <div className="flex items-center justify-center h-full text-gray-500">
            Chart type &quot;{type}&quot; not supported
          </div>
        );
    }
  };

  return (
    <div className={`w-full ${component.className || ''}`}>
      {title && <h4 className="text-sm font-medium text-gray-700 mb-2">{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        {renderChart()}
      </ResponsiveContainer>
    </div>
  );
}
