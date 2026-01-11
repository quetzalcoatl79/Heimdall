'use client';

import { ReactNode, useState } from 'react';
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
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  ChevronsUpDown,
  XCircle,
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

// ----- Enhanced Table Component with Filters, Sort, Selection, Pagination -----

function TableComponent({ component }: RendererProps) {
  const { 
    columns, 
    data, 
    filterable = false,
    sortable = true,
    selectable = false,
    paginated = false,
    pageSize = 10,
    searchable = false,
    actions = [],
    emptyMessage = 'Aucune donnée',
    onRowClick,
    rowKey = 'id',
  } = component.props || {};
  
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' } | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [columnFilters, setColumnFilters] = useState<Record<string, string>>({});

  if (!columns) {
    return <div className="text-gray-500">Configuration de tableau manquante</div>;
  }

  const tableData: any[] = data || [];

  // Apply search filter
  let filteredData = tableData;
  if (searchable && searchQuery) {
    const query = searchQuery.toLowerCase();
    filteredData = filteredData.filter((row) =>
      columns.some((col: any) => {
        const value = row[col.key];
        return value && String(value).toLowerCase().includes(query);
      })
    );
  }

  // Apply column filters
  if (filterable) {
    Object.entries(columnFilters).forEach(([key, filterValue]) => {
      if (filterValue) {
        filteredData = filteredData.filter((row) => {
          const value = row[key];
          if (value === null || value === undefined) return false;
          return String(value).toLowerCase().includes(filterValue.toLowerCase());
        });
      }
    });
  }

  // Apply sorting
  if (sortConfig) {
    filteredData = [...filteredData].sort((a, b) => {
      const aVal = a[sortConfig.key];
      const bVal = b[sortConfig.key];
      
      if (aVal === bVal) return 0;
      if (aVal === null || aVal === undefined) return 1;
      if (bVal === null || bVal === undefined) return -1;
      
      const comparison = aVal < bVal ? -1 : 1;
      return sortConfig.direction === 'asc' ? comparison : -comparison;
    });
  }

  // Pagination
  const totalPages = paginated ? Math.ceil(filteredData.length / pageSize) : 1;
  const paginatedData = paginated
    ? filteredData.slice((currentPage - 1) * pageSize, currentPage * pageSize)
    : filteredData;

  // Handlers
  const handleSort = (key: string) => {
    if (!sortable) return;
    setSortConfig((current) => {
      if (current?.key === key) {
        return current.direction === 'asc' 
          ? { key, direction: 'desc' } 
          : null;
      }
      return { key, direction: 'asc' };
    });
  };

  const handleSelectAll = () => {
    if (selectedRows.size === paginatedData.length) {
      setSelectedRows(new Set());
    } else {
      setSelectedRows(new Set(paginatedData.map((row) => row[rowKey] || JSON.stringify(row))));
    }
  };

  const handleSelectRow = (rowId: string) => {
    setSelectedRows((current) => {
      const next = new Set(current);
      if (next.has(rowId)) {
        next.delete(rowId);
      } else {
        next.add(rowId);
      }
      return next;
    });
  };

  const getRowId = (row: any) => row[rowKey] || row.bssid || row.id || JSON.stringify(row);

  // Get unique values for filter dropdowns
  const getUniqueValues = (key: string) => {
    const values = new Set<string>();
    tableData.forEach((row) => {
      if (row[key] !== null && row[key] !== undefined) {
        values.add(String(row[key]));
      }
    });
    return Array.from(values).sort();
  };

  return (
    <div className="space-y-4">
      {/* Search and bulk actions bar */}
      {(searchable || (selectable && selectedRows.size > 0)) && (
        <div className="flex items-center justify-between gap-4">
          {searchable && (
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                placeholder="Rechercher..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setCurrentPage(1);
                }}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
          )}
          {selectable && selectedRows.size > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-600">
                {selectedRows.size} sélectionné(s)
              </span>
              {actions.map((action: any) => (
                <button
                  key={action.id}
                  onClick={() => action.onClick?.(Array.from(selectedRows))}
                  className={`btn btn-sm ${action.variant === 'danger' ? 'btn-danger' : 'btn-secondary'}`}
                >
                  {action.icon && (() => {
                    const Icon = getIcon(action.icon);
                    return <Icon className="h-4 w-4 mr-1" />;
                  })()}
                  {action.label}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Table */}
      <div className="overflow-x-auto border border-gray-200 rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {selectable && (
                <th className="w-12 px-4 py-3">
                  <input
                    type="checkbox"
                    checked={paginatedData.length > 0 && selectedRows.size === paginatedData.length}
                    onChange={handleSelectAll}
                    className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                </th>
              )}
              {columns.map((col: any) => (
                <th
                  key={col.key}
                  className={`px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                    col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                  } ${sortable && col.sortable !== false ? 'cursor-pointer hover:bg-gray-100 select-none' : ''}`}
                  style={{ width: col.width }}
                  onClick={() => col.sortable !== false && handleSort(col.key)}
                >
                  <div className="flex items-center gap-2">
                    <span>{col.label}</span>
                    {sortable && col.sortable !== false && (
                      <span className="text-gray-400">
                        {sortConfig?.key === col.key ? (
                          sortConfig?.direction === 'asc' ? (
                            <ChevronUp className="h-4 w-4" />
                          ) : (
                            <ChevronDown className="h-4 w-4" />
                          )
                        ) : (
                          <ChevronsUpDown className="h-4 w-4 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                </th>
              ))}
            </tr>
            {/* Filter row */}
            {filterable && (
              <tr className="bg-gray-25">
                {selectable && <th className="px-4 py-2" />}
                {columns.map((col: any) => (
                  <th key={col.key} className="px-6 py-2">
                    {col.filterable !== false && (
                      col.filterType === 'select' ? (
                        <select
                          value={columnFilters[col.key] || ''}
                          onChange={(e) => {
                            setColumnFilters((prev) => ({ ...prev, [col.key]: e.target.value }));
                            setCurrentPage(1);
                          }}
                          className="w-full text-sm border border-gray-300 rounded px-2 py-1 focus:ring-1 focus:ring-primary-500"
                        >
                          <option value="">Tous</option>
                          {getUniqueValues(col.key).map((val) => (
                            <option key={val} value={val}>{val}</option>
                          ))}
                        </select>
                      ) : (
                        <input
                          type="text"
                          placeholder={`Filtrer ${col.label.toLowerCase()}...`}
                          value={columnFilters[col.key] || ''}
                          onChange={(e) => {
                            setColumnFilters((prev) => ({ ...prev, [col.key]: e.target.value }));
                            setCurrentPage(1);
                          }}
                          className="w-full text-sm border border-gray-300 rounded px-2 py-1 focus:ring-1 focus:ring-primary-500"
                        />
                      )
                    )}
                  </th>
                ))}
              </tr>
            )}
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {paginatedData.length === 0 ? (
              <tr>
                <td 
                  colSpan={columns.length + (selectable ? 1 : 0)} 
                  className="px-6 py-12 text-center text-gray-500"
                >
                  <Box className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              paginatedData.map((row: any, i: number) => {
                const rowId = getRowId(row);
                const isSelected = selectedRows.has(rowId);
                
                return (
                  <tr 
                    key={rowId || i}
                    className={`${isSelected ? 'bg-primary-50' : 'hover:bg-gray-50'} ${onRowClick ? 'cursor-pointer' : ''}`}
                    onClick={() => onRowClick?.(row)}
                  >
                    {selectable && (
                      <td className="w-12 px-4 py-4" onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleSelectRow(rowId)}
                          className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                        />
                      </td>
                    )}
                    {columns.map((col: any) => (
                      <td 
                        key={col.key} 
                        className={`px-6 py-4 whitespace-nowrap text-sm text-gray-900 ${
                          col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                        }`}
                      >
                        {renderCellValue(row[col.key], col.render, col, row)}
                      </td>
                    ))}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {paginated && totalPages > 1 && (
        <div className="flex items-center justify-between px-2">
          <div className="text-sm text-gray-600">
            {filteredData.length} résultat(s) • Page {currentPage} sur {totalPages}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentPage(1)}
              disabled={currentPage === 1}
              className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronsLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="px-3 py-1 bg-gray-100 rounded text-sm font-medium">
              {currentPage}
            </span>
            <button
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
            <button
              onClick={() => setCurrentPage(totalPages)}
              disabled={currentPage === totalPages}
              className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronsRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

// Enhanced cell renderer with more options
function renderCellValue(value: any, render?: string, col?: any, row?: any) {
  if (value === null || value === undefined) {
    return <span className="text-gray-400">-</span>;
  }

  switch (render) {
    case 'badge': {
      const variantMap: Record<string, string> = {
        healthy: 'bg-green-100 text-green-800',
        active: 'bg-green-100 text-green-800',
        running: 'bg-green-100 text-green-800',
        online: 'bg-green-100 text-green-800',
        connected: 'bg-green-100 text-green-800',
        open: 'bg-green-100 text-green-800',
        unhealthy: 'bg-red-100 text-red-800',
        inactive: 'bg-red-100 text-red-800',
        offline: 'bg-red-100 text-red-800',
        error: 'bg-red-100 text-red-800',
        failed: 'bg-red-100 text-red-800',
        pending: 'bg-yellow-100 text-yellow-800',
        warning: 'bg-yellow-100 text-yellow-800',
        'WPA/WPA2': 'bg-orange-100 text-orange-800',
        'WPA2': 'bg-orange-100 text-orange-800',
        'WPA3': 'bg-red-100 text-red-800',
        'WEP': 'bg-yellow-100 text-yellow-800',
      };
      const classes = variantMap[value] || 'bg-gray-100 text-gray-800';
      return (
        <span className={`px-2 py-1 text-xs font-medium rounded-full ${classes}`}>
          {value}
        </span>
      );
    }
    
    case 'date':
      return new Date(value).toLocaleDateString();
    
    case 'datetime':
      return new Date(value).toLocaleString();
    
    case 'relative':
      return formatRelativeTime(value);
    
    case 'signal': {
      const signal = Number(value);
      const color = signal >= -50 ? 'text-green-600' : signal >= -70 ? 'text-yellow-600' : 'text-red-600';
      const bars = signal >= -50 ? 4 : signal >= -60 ? 3 : signal >= -70 ? 2 : 1;
      return (
        <div className="flex items-center gap-2">
          <div className="flex items-end gap-0.5 h-4">
            {[1, 2, 3, 4].map((bar) => (
              <div
                key={bar}
                className={`w-1 rounded-sm ${bar <= bars ? 'bg-current ' + color : 'bg-gray-200'}`}
                style={{ height: `${bar * 25}%` }}
              />
            ))}
          </div>
          <span className={color}>{value} dBm</span>
        </div>
      );
    }
    
    case 'boolean':
      return value ? (
        <CheckCircle className="h-5 w-5 text-green-500" />
      ) : (
        <XCircle className="h-5 w-5 text-red-500" />
      );
    
    case 'link':
      return (
        <a href={value} target="_blank" rel="noopener noreferrer" className="text-primary-600 hover:underline">
          {value}
        </a>
      );
    
    case 'code':
      return (
        <code className="px-2 py-1 bg-gray-100 rounded text-sm font-mono">
          {value}
        </code>
      );
    
    case 'percent': {
      const percent = Number(value);
      const color = percent >= 80 ? 'bg-green-500' : percent >= 50 ? 'bg-yellow-500' : 'bg-red-500';
      return (
        <div className="flex items-center gap-2">
          <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
            <div className={`h-full ${color}`} style={{ width: `${Math.min(100, percent)}%` }} />
          </div>
          <span className="text-sm">{percent}%</span>
        </div>
      );
    }
    
    default:
      return String(value);
  }
}

// Helper for relative time
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  
  if (diffSecs < 60) return `Il y a ${diffSecs}s`;
  if (diffSecs < 3600) return `Il y a ${Math.floor(diffSecs / 60)}min`;
  if (diffSecs < 86400) return `Il y a ${Math.floor(diffSecs / 3600)}h`;
  return `Il y a ${Math.floor(diffSecs / 86400)}j`;
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
