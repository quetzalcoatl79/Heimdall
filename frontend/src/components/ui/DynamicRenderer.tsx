'use client';

import React, { ReactNode, useState, useCallback, createContext, useContext, useEffect } from 'react';
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
  Wifi,
  Radio,
  Key,
  ClipboardList,
  X,
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
import { apiClient } from '@/lib/api/client';

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

// ----- Context for actions and modals -----
interface DynamicRendererContextType {
  onAction?: (actionId: string, data?: any) => void;
  openModal: (modalId: string) => void;
  closeModal: (modalId: string) => void;
  activeModals: Set<string>;
  selectedRows: Map<string, Set<string>>;
  setSelectedRows: (tableId: string, rows: Set<string>) => void;
  refreshView: () => void;
}

const DynamicRendererContext = createContext<DynamicRendererContextType>({
  openModal: () => {},
  closeModal: () => {},
  activeModals: new Set(),
  selectedRows: new Map(),
  setSelectedRows: () => {},
  refreshView: () => {},
});

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
  wifi: Wifi,
  radio: Radio,
  key: Key,
  'clipboard-list': ClipboardList,
  'file-text': FileText,
};

export function getIcon(name: string): LucideIcon {
  return iconMap[name?.toLowerCase()] || Box;
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
  onAction?: (actionId: string, data?: any) => void;
  onRefresh?: () => void;
}

export function DynamicRenderer({ schema, onAction, onRefresh }: DynamicRendererProps) {
  const IconComponent = schema.icon ? getIcon(schema.icon) : null;
  const [activeModals, setActiveModals] = useState<Set<string>>(new Set());
  const [selectedRows, setSelectedRowsState] = useState<Map<string, Set<string>>>(new Map());
  
  const openModal = useCallback((modalId: string) => {
    setActiveModals(prev => new Set(prev).add(modalId));
  }, []);
  
  const closeModal = useCallback((modalId: string) => {
    setActiveModals(prev => {
      const next = new Set(prev);
      next.delete(modalId);
      return next;
    });
  }, []);
  
  const setSelectedRows = useCallback((tableId: string, rows: Set<string>) => {
    setSelectedRowsState(prev => new Map(prev).set(tableId, rows));
  }, []);
  
  const refreshView = useCallback(() => {
    onRefresh?.();
  }, [onRefresh]);
  
  const contextValue: DynamicRendererContextType = {
    onAction,
    openModal,
    closeModal,
    activeModals,
    selectedRows,
    setSelectedRows,
    refreshView,
  };

  return (
    <DynamicRendererContext.Provider value={contextValue}>
      <div className="space-y-6">
        {/* Header */}
        {(schema.title || schema.actions?.length || schema.refresh?.enabled) && (
          <div className="flex flex-col sm:flex-row justify-between items-start gap-4">
            <div className="flex-1 min-w-0">
              {schema.title && (
                <h1 className="text-xl sm:text-2xl font-bold flex items-center gap-2 sm:gap-3">
                  {IconComponent && <IconComponent className="h-6 w-6 sm:h-7 sm:w-7 text-primary-600 flex-shrink-0" />}
                  <span className="truncate">{schema.title}</span>
                </h1>
              )}
              {schema.description && (
                <p className="text-gray-500 mt-1 text-sm sm:text-base">{schema.description}</p>
              )}
            </div>
            <div className="flex items-center gap-3 flex-shrink-0">
              {/* Auto-refresh indicator */}
              {schema.refresh?.enabled && (
                <div className="hidden sm:flex items-center gap-2 text-xs text-gray-400 bg-gray-100 px-3 py-1.5 rounded-full">
                  <RefreshCw className="h-3 w-3 animate-spin" style={{ animationDuration: '3s' }} />
                  <span>{schema.refresh.interval}s</span>
                </div>
              )}
              {/* Actions */}
              {schema.actions && schema.actions.length > 0 && (
                <div className="flex gap-2">
                  {schema.actions.map((action) => (
                    <button
                      key={action.id}
                      onClick={() => onAction?.(action.id)}
                      disabled={action.disabled}
                      className={`btn btn-sm sm:btn ${
                        action.variant === 'primary'
                          ? 'btn-primary'
                          : action.variant === 'danger'
                          ? 'btn-danger'
                          : 'btn-secondary'
                      }`}
                    >
                      {action.icon && (() => {
                        const Icon = getIcon(action.icon);
                        return <Icon className="h-4 w-4 sm:mr-2" />;
                      })()}
                      <span className="hidden sm:inline">{action.label}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Components */}
        {schema.components.map((component, index) => (
          <ComponentRenderer key={component.id || index} component={component} />
        ))}
      </div>
    </DynamicRendererContext.Provider>
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
    case 'tabs':
      return <TabsComponent component={component} />;
    case 'tab':
      return <TabComponent component={component} />;
    case 'button':
      return <ButtonComponent component={component} />;
    case 'buttonGroup':
      return <ButtonGroupComponent component={component} />;
    case 'actionBar':
      return <ActionBarComponent component={component} />;
    case 'modal':
      return <ModalComponent component={component} />;
    case 'container':
      return <ContainerComponent component={component} />;
    default:
      // Ne pas afficher d'erreur pour les types vides ou les containers
      if (!component.type || component.type === '') {
        return component.children ? (
          <>{component.children.map((child, i) => (
            <ComponentRenderer key={child.id || i} component={child} />
          ))}</>
        ) : null;
      }
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
  // Responsive: 1 colonne sur mobile, 2 sur tablette, cols sur desktop
  const getGridClass = () => {
    const colClasses: Record<number, string> = {
      1: 'grid-cols-1',
      2: 'grid-cols-1 md:grid-cols-2',
      3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
      4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
      5: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5',
      6: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6',
    };
    return colClasses[cols] || `grid-cols-1 md:grid-cols-2 lg:grid-cols-${cols}`;
  };
  
  return (
    <div className={`grid ${getGridClass()} gap-${gap} ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function RowComponent({ component }: RendererProps) {
  return (
    <div className={`flex flex-col md:flex-row flex-wrap gap-4 ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function ColComponent({ component }: RendererProps) {
  const span = component.props?.span || 1;
  // Responsive: sur mobile toutes les colonnes prennent 100%, sur desktop le span spécifié
  const getColClass = () => {
    const spanClasses: Record<number, string> = {
      1: 'md:w-1/12',
      2: 'md:w-2/12',
      3: 'md:w-3/12',
      4: 'md:w-4/12',
      5: 'md:w-5/12',
      6: 'md:w-6/12',
      7: 'md:w-7/12',
      8: 'md:w-8/12',
      9: 'md:w-9/12',
      10: 'md:w-10/12',
      11: 'md:w-11/12',
      12: 'md:w-full',
    };
    return spanClasses[span] || 'md:flex-1';
  };
  
  return (
    <div className={`w-full ${getColClass()} ${component.className || ''}`}>
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
// RESPONSIVE: Colonnes visibles + expand pour voir plus sur mobile
// La dernière colonne (souvent actions) reste toujours visible

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
    // Nombre de colonnes visibles sur mobile au début (+ dernière colonne)
    mobileVisibleCols = 2,
    // Garder la dernière colonne visible (pour actions)
    keepLastColumn = true,
  } = component.props || {};
  
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' } | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [columnFilters, setColumnFilters] = useState<Record<string, string>>({});
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  if (!columns) {
    return <div className="text-gray-500">Configuration de tableau manquante</div>;
  }

  const tableData: any[] = data || [];
  
  // Calculer les colonnes visibles/cachées avec option pour garder la dernière
  const totalCols = columns.length;
  let startVisible: any[] = [];
  let middleHidden: any[] = [];
  let endVisible: any[] = [];
  
  if (keepLastColumn && totalCols > mobileVisibleCols + 1) {
    // Garder les N premières + la dernière, cacher le milieu
    startVisible = columns.slice(0, mobileVisibleCols);
    middleHidden = columns.slice(mobileVisibleCols, totalCols - 1);
    endVisible = columns.slice(totalCols - 1);
  } else {
    // Mode standard : N premières visibles, le reste caché
    startVisible = columns.slice(0, mobileVisibleCols);
    middleHidden = columns.slice(mobileVisibleCols);
    endVisible = [];
  }
  
  const visibleColumns = [...startVisible, ...endVisible];
  const hiddenColumns = middleHidden;
  const hasHiddenColumns = hiddenColumns.length > 0;

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

  const toggleRowExpand = (rowId: string) => {
    setExpandedRows((current) => {
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
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 sm:gap-4">
          {searchable && (
            <div className="relative w-full sm:flex-1 sm:max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                placeholder="Rechercher..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setCurrentPage(1);
                }}
                className="w-full pl-10 pr-4 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
          )}
          {selectable && selectedRows.size > 0 && (
            <div className="flex items-center gap-2 flex-wrap">
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

      {/* Table with responsive expand */}
      <div className="overflow-x-auto border border-gray-200 rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {selectable && (
                <th className="w-10 px-2 py-3">
                  <input
                    type="checkbox"
                    checked={paginatedData.length > 0 && selectedRows.size === paginatedData.length}
                    onChange={handleSelectAll}
                    className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                </th>
              )}
              {/* Colonnes visibles sur mobile */}
              {visibleColumns.map((col: any) => (
                <th
                  key={col.key}
                  className={`px-2 sm:px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                    col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                  } ${sortable && col.sortable !== false ? 'cursor-pointer hover:bg-gray-100 select-none' : ''}`}
                  style={{ width: col.width, minWidth: col.minWidth || 'auto' }}
                  onClick={() => col.sortable !== false && handleSort(col.key)}
                >
                  <div className="flex items-center gap-1">
                    <span className="truncate">{col.label}</span>
                    {sortable && col.sortable !== false && (
                      <span className="text-gray-400 flex-shrink-0">
                        {sortConfig?.key === col.key ? (
                          sortConfig?.direction === 'asc' ? (
                            <ChevronUp className="h-3 w-3" />
                          ) : (
                            <ChevronDown className="h-3 w-3" />
                          )
                        ) : (
                          <ChevronsUpDown className="h-3 w-3 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                </th>
              ))}
              {/* Colonnes cachées sur mobile, visibles sur desktop */}
              {hiddenColumns.map((col: any) => (
                <th
                  key={col.key}
                  className={`hidden lg:table-cell px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                    col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                  } ${sortable && col.sortable !== false ? 'cursor-pointer hover:bg-gray-100 select-none' : ''}`}
                  style={{ width: col.width }}
                  onClick={() => col.sortable !== false && handleSort(col.key)}
                >
                  <div className="flex items-center gap-1">
                    <span className="truncate">{col.label}</span>
                    {sortable && col.sortable !== false && (
                      <span className="text-gray-400 flex-shrink-0">
                        {sortConfig?.key === col.key ? (
                          sortConfig?.direction === 'asc' ? (
                            <ChevronUp className="h-3 w-3" />
                          ) : (
                            <ChevronDown className="h-3 w-3" />
                          )
                        ) : (
                          <ChevronsUpDown className="h-3 w-3 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                </th>
              ))}
              {/* Bouton expand - visible uniquement sur mobile/tablette si colonnes cachées */}
              {hasHiddenColumns && (
                <th className="lg:hidden w-10 px-2 py-3"></th>
              )}
            </tr>
            {/* Filter row */}
            {filterable && (
              <tr className="bg-gray-25">
                {selectable && <th className="px-2 py-2" />}
                {visibleColumns.map((col: any) => (
                  <th key={col.key} className="px-2 sm:px-4 py-2">
                    {col.filterable !== false && (
                      col.filterType === 'select' ? (
                        <select
                          value={columnFilters[col.key] || ''}
                          onChange={(e) => {
                            setColumnFilters((prev) => ({ ...prev, [col.key]: e.target.value }));
                            setCurrentPage(1);
                          }}
                          className="w-full text-xs border border-gray-300 rounded px-1.5 py-1 focus:ring-1 focus:ring-primary-500"
                        >
                          <option value="">Tous</option>
                          {getUniqueValues(col.key).map((val) => (
                            <option key={val} value={val}>{val}</option>
                          ))}
                        </select>
                      ) : (
                        <input
                          type="text"
                          placeholder="Filtrer..."
                          value={columnFilters[col.key] || ''}
                          onChange={(e) => {
                            setColumnFilters((prev) => ({ ...prev, [col.key]: e.target.value }));
                            setCurrentPage(1);
                          }}
                          className="w-full text-xs border border-gray-300 rounded px-1.5 py-1 focus:ring-1 focus:ring-primary-500"
                        />
                      )
                    )}
                  </th>
                ))}
                {hiddenColumns.map((col: any) => (
                  <th key={col.key} className="hidden lg:table-cell px-4 py-2">
                    {col.filterable !== false && (
                      <input
                        type="text"
                        placeholder="Filtrer..."
                        value={columnFilters[col.key] || ''}
                        onChange={(e) => {
                          setColumnFilters((prev) => ({ ...prev, [col.key]: e.target.value }));
                          setCurrentPage(1);
                        }}
                        className="w-full text-xs border border-gray-300 rounded px-1.5 py-1 focus:ring-1 focus:ring-primary-500"
                      />
                    )}
                  </th>
                ))}
                {hasHiddenColumns && <th className="lg:hidden px-2 py-2" />}
              </tr>
            )}
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {paginatedData.length === 0 ? (
              <tr>
                <td 
                  colSpan={columns.length + (selectable ? 1 : 0) + (hasHiddenColumns ? 1 : 0)} 
                  className="px-4 py-12 text-center text-gray-500"
                >
                  <Box className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              paginatedData.map((row: any, i: number) => {
                const rowId = getRowId(row);
                const isSelected = selectedRows.has(rowId);
                const isExpanded = expandedRows.has(rowId);
                
                return (
                  <React.Fragment key={rowId || i}>
                    {/* Ligne principale */}
                    <tr 
                      className={`${isSelected ? 'bg-primary-50' : 'hover:bg-gray-50'} ${onRowClick ? 'cursor-pointer' : ''}`}
                      onClick={() => onRowClick?.(row)}
                    >
                      {selectable && (
                        <td className="w-10 px-2 py-3" onClick={(e) => e.stopPropagation()}>
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => handleSelectRow(rowId)}
                            className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                          />
                        </td>
                      )}
                      {/* Colonnes visibles */}
                      {visibleColumns.map((col: any) => (
                        <td 
                          key={col.key} 
                          className={`px-2 sm:px-4 py-3 text-sm text-gray-900 ${
                            col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                          }`}
                        >
                          <div className="truncate max-w-[120px] sm:max-w-[180px] lg:max-w-none">
                            {renderCellValue(row[col.key], col.render, col, row)}
                          </div>
                        </td>
                      ))}
                      {/* Colonnes cachées sur mobile */}
                      {hiddenColumns.map((col: any) => (
                        <td 
                          key={col.key} 
                          className={`hidden lg:table-cell px-4 py-3 text-sm text-gray-900 ${
                            col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : ''
                          }`}
                        >
                          {renderCellValue(row[col.key], col.render, col, row)}
                        </td>
                      ))}
                      {/* Bouton expand */}
                      {hasHiddenColumns && (
                        <td className="lg:hidden w-10 px-2 py-3">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              toggleRowExpand(rowId);
                            }}
                            className="p-1 rounded hover:bg-gray-200 text-gray-500"
                            title={isExpanded ? 'Réduire' : 'Voir plus'}
                          >
                            {isExpanded ? (
                              <ChevronUp className="h-4 w-4" />
                            ) : (
                              <ChevronDown className="h-4 w-4" />
                            )}
                          </button>
                        </td>
                      )}
                    </tr>
                    {/* Ligne expandée avec colonnes cachées */}
                    {hasHiddenColumns && isExpanded && (
                      <tr className="lg:hidden bg-gray-50">
                        <td 
                          colSpan={visibleColumns.length + (selectable ? 1 : 0) + 1}
                          className="px-4 py-3"
                        >
                          <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                            {hiddenColumns.map((col: any) => (
                              <div key={col.key}>
                                <span className="text-gray-500 text-xs font-medium">{col.label}</span>
                                <div className="text-gray-900 mt-0.5">
                                  {renderCellValue(row[col.key], col.render, col, row)}
                                </div>
                              </div>
                            ))}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination - responsive */}
      {paginated && totalPages > 1 && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-2 px-2">
          <div className="text-xs sm:text-sm text-gray-600 order-2 sm:order-1">
            {filteredData.length} résultat(s) • Page {currentPage}/{totalPages}
          </div>
          <div className="flex items-center gap-1 sm:gap-2 order-1 sm:order-2">
            <button
              onClick={() => setCurrentPage(1)}
              disabled={currentPage === 1}
              className="p-1.5 sm:p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronsLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="p-1.5 sm:p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="px-2 sm:px-3 py-1 bg-gray-100 rounded text-xs sm:text-sm font-medium min-w-[2rem] text-center">
              {currentPage}
            </span>
            <button
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="p-1.5 sm:p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
            <button
              onClick={() => setCurrentPage(totalPages)}
              disabled={currentPage === totalPages}
              className="p-1.5 sm:p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
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
function renderCellValue(value: any, render?: string, col?: any, row?: any): React.ReactNode {
  if (value === null || value === undefined) {
    return <span className="text-gray-400">-</span>;
  }

  // Handle rowActions type (from backend)
  if (typeof value === 'object' && value.type === 'rowActions') {
    return <RowActionsRenderer actions={value.items || []} row={row} />;
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

// RowActions renderer for table row actions (deauth, edit, delete, etc.)
interface RowAction {
  id: string;
  label: string;
  icon?: string;
  variant?: 'primary' | 'secondary' | 'danger';
  endpoint?: string;
  method?: string;
  data?: Record<string, any>;
  confirm?: string;
  disabled?: boolean;
  action?: string;
}

function RowActionsRenderer({ actions, row }: { actions: RowAction[]; row: any }) {
  const ctx = useContext(DynamicRendererContext);
  const [loading, setLoading] = useState<string | null>(null);

  const handleAction = async (action: RowAction) => {
    // Handle confirmation
    if (action.confirm && !window.confirm(action.confirm)) {
      return;
    }

    // Handle custom action
    if (action.action && ctx.onAction) {
      ctx.onAction(action.action, { ...action.data, row });
      return;
    }

    // Handle API endpoint
    if (action.endpoint) {
      setLoading(action.id);
      try {
        // Replace {id} or {bssid} placeholders with actual values from row
        let endpoint = action.endpoint;
        Object.keys(row).forEach((key) => {
          endpoint = endpoint.replace(`{${key}}`, row[key]);
        });

        // Merge action data with row data for context
        const form = document.getElementById('wifi-interface-form') as HTMLFormElement;
        const selectElement = form?.querySelector('select[name="interface"]') as HTMLSelectElement | null;
        const payload = {
          ...action.data,
          interface: selectElement?.value || row.interface,
        };

        if (action.method?.toUpperCase() === 'DELETE') {
          await apiClient.delete(endpoint);
        } else if (action.method?.toUpperCase() === 'GET') {
          await apiClient.get(endpoint);
        } else {
          await apiClient.post(endpoint, payload);
        }

        ctx.refreshView();
      } catch (error) {
        console.error('Row action failed:', error);
      } finally {
        setLoading(null);
      }
    }
  };

  return (
    <div className="flex items-center gap-1">
      {actions.map((action) => {
        const Icon = action.icon ? getIcon(action.icon) : null;
        const isLoading = loading === action.id;
        
        const variantClasses: Record<string, string> = {
          primary: 'bg-primary-600 hover:bg-primary-700 text-white',
          secondary: 'bg-gray-100 hover:bg-gray-200 text-gray-700',
          danger: 'bg-red-600 hover:bg-red-700 text-white',
        };

        return (
          <button
            key={action.id}
            onClick={(e) => {
              e.stopPropagation();
              handleAction(action);
            }}
            disabled={action.disabled || isLoading}
            title={action.label}
            className={`p-1.5 rounded text-xs font-medium transition-colors disabled:opacity-50 ${
              variantClasses[action.variant || 'secondary']
            }`}
          >
            {isLoading ? (
              <RefreshCw className="h-3.5 w-3.5 animate-spin" />
            ) : Icon ? (
              <Icon className="h-3.5 w-3.5" />
            ) : (
              action.label
            )}
          </button>
        );
      })}
    </div>
  );
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

// ----- Tabs Component -----

function TabsComponent({ component }: RendererProps) {
  const { defaultTab } = component.props || {};
  const [activeTab, setActiveTab] = useState<string>(defaultTab || '');
  
  // Filter only tab children
  const tabs = (component.children || []).filter(c => c.type === 'tab');
  
  // Set default tab if not set
  useEffect(() => {
    if (!activeTab && tabs.length > 0) {
      setActiveTab(tabs[0].id || tabs[0].props?.label || '0');
    }
  }, [tabs, activeTab]);

  return (
    <div className={component.className || ''}>
      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex gap-4" aria-label="Tabs">
          {tabs.map((tab, index) => {
            const tabId = tab.id || tab.props?.label || String(index);
            const isActive = activeTab === tabId;
            const Icon = tab.props?.icon ? getIcon(tab.props.icon) : null;
            const badge = tab.props?.badge;
            
            return (
              <button
                key={tabId}
                onClick={() => setActiveTab(tabId)}
                className={`
                  flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors
                  ${isActive 
                    ? 'border-primary-600 text-primary-600' 
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}
                `}
              >
                {Icon && <Icon className="h-4 w-4" />}
                {tab.props?.label}
                {badge && (
                  <span className="ml-2 px-2 py-0.5 text-xs bg-primary-100 text-primary-700 rounded-full">
                    {badge}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
      </div>
      
      {/* Active Tab Content */}
      {tabs.map((tab, index) => {
        const tabId = tab.id || tab.props?.label || String(index);
        if (activeTab !== tabId) return null;
        
        return (
          <div key={tabId}>
            {tab.children?.map((child, i) => (
              <ComponentRenderer key={child.id || i} component={child} />
            ))}
          </div>
        );
      })}
    </div>
  );
}

function TabComponent({ component }: RendererProps) {
  // Tab is rendered by TabsComponent, this is just a wrapper
  return (
    <div className={component.className || ''}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

// ----- Button Component -----

function ButtonComponent({ component }: RendererProps) {
  const { 
    label, 
    variant = 'secondary', 
    icon, 
    endpoint, 
    method = 'POST', 
    action,
    disabled,
    formId,
    requiresRows,
    confirm: confirmMsg,
  } = component.props || {};
  
  const ctx = useContext(DynamicRendererContext);
  const [loading, setLoading] = useState(false);
  const Icon = icon ? getIcon(icon) : null;

  const handleClick = async () => {
    // Handle confirmation
    if (confirmMsg && !window.confirm(confirmMsg)) {
      return;
    }

    // Handle modal opening
    if (action === 'openModal' && component.props?.modal) {
      ctx.openModal(component.props.modal);
      return;
    }

    // Handle API endpoint first (if both action and endpoint exist, endpoint wins)
    if (endpoint) {
      setLoading(true);
      try {
        // Get form data if formId is specified
        let data: any = {};
        if (formId) {
          const form = document.getElementById(formId) as HTMLFormElement;
          if (form) {
            const formData = new FormData(form);
            formData.forEach((value, key) => {
              data[key] = value;
            });
          }
        }
        
        // Get selected rows if required
        if (requiresRows) {
          const rows = ctx.selectedRows.get('wifi-networks');
          if (rows && rows.size > 0) {
            data.targets = Array.from(rows);
          }
        }

        if (method.toUpperCase() === 'GET') {
          await apiClient.get(endpoint);
        } else if (method.toUpperCase() === 'DELETE') {
          await apiClient.delete(endpoint);
        } else {
          await apiClient.post(endpoint, data);
        }
        
        ctx.refreshView();
      } catch (error) {
        console.error('Button action failed:', error);
      } finally {
        setLoading(false);
      }
      return;
    }

    // Handle custom action (only if no endpoint)
    if (action && ctx.onAction) {
      ctx.onAction(action);
      return;
    }
  };

  const variantClasses: Record<string, string> = {
    primary: 'btn btn-primary',
    secondary: 'btn btn-secondary',
    danger: 'btn btn-danger',
    ghost: 'btn btn-ghost',
  };

  return (
    <button
      onClick={handleClick}
      disabled={disabled || loading}
      className={`${variantClasses[variant] || variantClasses.secondary} ${loading ? 'opacity-50' : ''}`}
    >
      {loading ? (
        <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
      ) : Icon ? (
        <Icon className="h-4 w-4 mr-2" />
      ) : null}
      {label}
    </button>
  );
}

function ButtonGroupComponent({ component }: RendererProps) {
  return (
    <div className={`flex gap-2 ${component.className || ''}`}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

function ActionBarComponent({ component }: RendererProps) {
  const { position = 'bottom' } = component.props || {};
  
  return (
    <div className={`
      flex items-center gap-3 py-4
      ${position === 'top' ? 'mb-4 border-b border-gray-200' : 'mt-4 pt-4 border-t border-gray-200'}
      ${component.className || ''}
    `}>
      {component.children?.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}

// ----- Modal Component -----

function ModalComponent({ component }: RendererProps) {
  const ctx = useContext(DynamicRendererContext);
  const modalId = component.id || '';
  const isOpen = ctx.activeModals.has(modalId);
  
  const { title, icon, size = 'md' } = component.props || {};
  const Icon = icon ? getIcon(icon) : null;

  if (!isOpen) return null;

  const sizeClasses: Record<string, string> = {
    sm: 'max-w-md',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl',
    full: 'max-w-full mx-4',
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={() => ctx.closeModal(modalId)}
      />
      
      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className={`relative bg-white rounded-xl shadow-xl w-full ${sizeClasses[size] || sizeClasses.md}`}>
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              {Icon && <Icon className="h-5 w-5 text-primary-600" />}
              {title}
            </h3>
            <button
              onClick={() => ctx.closeModal(modalId)}
              className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <X className="h-5 w-5 text-gray-500" />
            </button>
          </div>
          
          {/* Content */}
          <div className="p-4">
            {component.children?.map((child, i) => (
              <ComponentRenderer key={child.id || i} component={child} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ----- Container Component -----

function ContainerComponent({ component }: RendererProps) {
  // Container est un simple wrapper qui rend ses enfants
  if (!component.children || component.children.length === 0) {
    return null;
  }
  
  return (
    <div className={component.className || ''}>
      {component.children.map((child, i) => (
        <ComponentRenderer key={child.id || i} component={child} />
      ))}
    </div>
  );
}
