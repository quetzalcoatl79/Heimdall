package ui

// ViewSchema represents a complete page/view definition that the frontend will render.
// This is similar to nano-core's template system but uses JSON instead of Django templates.
type ViewSchema struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Icon        string         `json:"icon,omitempty"`
	Actions     []Action       `json:"actions,omitempty"`
	Components  []Component    `json:"components"`
	Refresh     *RefreshConfig `json:"refresh,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// RefreshConfig defines auto-refresh behavior
type RefreshConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"` // in seconds
}

// Component represents any UI component that can be rendered
type Component struct {
	Type       string         `json:"type"`
	ID         string         `json:"id,omitempty"`
	Props      map[string]any `json:"props,omitempty"`
	Children   []Component    `json:"children,omitempty"`
	DataSource *DataSource    `json:"dataSource,omitempty"`
	Visible    *bool          `json:"visible,omitempty"`
	ClassName  string         `json:"className,omitempty"`
}

// DataSource defines where to fetch data for a component
type DataSource struct {
	URL       string `json:"url"`
	Method    string `json:"method,omitempty"` // GET by default
	RefreshOn string `json:"refreshOn,omitempty"`
}

// Action represents a button/action that can be performed
type Action struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Icon     string         `json:"icon,omitempty"`
	Variant  string         `json:"variant,omitempty"` // primary, secondary, danger
	Href     string         `json:"href,omitempty"`
	OnClick  string         `json:"onClick,omitempty"` // action identifier
	Confirm  *ConfirmDialog `json:"confirm,omitempty"`
	Disabled bool           `json:"disabled,omitempty"`
}

// ConfirmDialog for actions that need confirmation
type ConfirmDialog struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Confirm string `json:"confirm,omitempty"`
	Cancel  string `json:"cancel,omitempty"`
}

// ----- Component Type Constants -----

const (
	ComponentCard      = "card"
	ComponentTable     = "table"
	ComponentForm      = "form"
	ComponentStats     = "stats"
	ComponentStatItem  = "stat"
	ComponentAlert     = "alert"
	ComponentTabs      = "tabs"
	ComponentTab       = "tab"
	ComponentGrid      = "grid"
	ComponentRow       = "row"
	ComponentCol       = "col"
	ComponentText      = "text"
	ComponentHeading   = "heading"
	ComponentBadge     = "badge"
	ComponentProgress  = "progress"
	ComponentList      = "list"
	ComponentListItem  = "listItem"
	ComponentDivider   = "divider"
	ComponentEmpty     = "empty"
	ComponentSkeleton  = "skeleton"
	ComponentChart     = "chart"
	ComponentTimeline  = "timeline"
	ComponentCodeBlock = "codeBlock"
	ComponentJSON      = "json"
	ComponentCustom    = "custom"
)

// ----- Builder Helpers -----

// NewView creates a new ViewSchema
func NewView(title string) *ViewSchema {
	return &ViewSchema{
		Title:      title,
		Components: []Component{},
	}
}

func (v *ViewSchema) WithDescription(desc string) *ViewSchema {
	v.Description = desc
	return v
}

func (v *ViewSchema) WithIcon(icon string) *ViewSchema {
	v.Icon = icon
	return v
}

func (v *ViewSchema) WithRefresh(interval int) *ViewSchema {
	v.Refresh = &RefreshConfig{Enabled: true, Interval: interval}
	return v
}

func (v *ViewSchema) AddAction(action Action) *ViewSchema {
	v.Actions = append(v.Actions, action)
	return v
}

func (v *ViewSchema) AddComponent(c Component) *ViewSchema {
	v.Components = append(v.Components, c)
	return v
}

// ----- Component Builders -----

// Card creates a card component
func Card(title string, children ...Component) Component {
	return Component{
		Type:     ComponentCard,
		Props:    map[string]any{"title": title},
		Children: children,
	}
}

// CardWithProps creates a card with custom props
func CardWithProps(props map[string]any, children ...Component) Component {
	return Component{
		Type:     ComponentCard,
		Props:    props,
		Children: children,
	}
}

// Table creates a table component
func Table(columns []TableColumn, data []map[string]any) Component {
	return Component{
		Type: ComponentTable,
		Props: map[string]any{
			"columns": columns,
			"data":    data,
		},
	}
}

// TableColumn defines a table column
type TableColumn struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Sortable   bool   `json:"sortable,omitempty"`
	Filterable bool   `json:"filterable,omitempty"`
	FilterType string `json:"filterType,omitempty"` // text, select
	Width      string `json:"width,omitempty"`
	Align      string `json:"align,omitempty"`  // left, center, right
	Render     string `json:"render,omitempty"` // badge, date, datetime, relative, signal, boolean, link, code, percent
	ClassName  string `json:"className,omitempty"`
}

// TableOption is a functional option for configuring tables
type TableOption func(map[string]any)

// TableFilterable makes the table filterable
func TableFilterable() TableOption {
	return func(p map[string]any) { p["filterable"] = true }
}

// TableSearchable adds a global search bar
func TableSearchable() TableOption {
	return func(p map[string]any) { p["searchable"] = true }
}

// TableSelectable enables row selection
func TableSelectable() TableOption {
	return func(p map[string]any) { p["selectable"] = true }
}

// TablePaginated enables pagination
func TablePaginated(pageSize int) TableOption {
	return func(p map[string]any) {
		p["paginated"] = true
		p["pageSize"] = pageSize
	}
}

// TableRowKey sets the unique key for rows
func TableRowKey(key string) TableOption {
	return func(p map[string]any) { p["rowKey"] = key }
}

// TableEmptyMessage sets the message when no data
func TableEmptyMessage(msg string) TableOption {
	return func(p map[string]any) { p["emptyMessage"] = msg }
}

// TableWithOptions creates a table with options
func TableWithOptions(columns []TableColumn, data []map[string]any, opts ...TableOption) Component {
	props := map[string]any{
		"columns": columns,
		"data":    data,
	}
	for _, opt := range opts {
		opt(props)
	}
	return Component{
		Type:  ComponentTable,
		Props: props,
	}
}

// TableWithSource creates a table that fetches data from an endpoint
func TableWithSource(columns []TableColumn, dataSourceURL string) Component {
	return Component{
		Type: ComponentTable,
		Props: map[string]any{
			"columns": columns,
		},
		DataSource: &DataSource{URL: dataSourceURL},
	}
}

// Stats creates a statistics grid
func Stats(items ...Component) Component {
	return Component{
		Type:     ComponentStats,
		Children: items,
	}
}

// Stat creates a single stat item
func Stat(label string, value any, opts ...StatOption) Component {
	props := map[string]any{
		"label": label,
		"value": value,
	}
	for _, opt := range opts {
		opt(props)
	}
	return Component{
		Type:  ComponentStatItem,
		Props: props,
	}
}

type StatOption func(map[string]any)

func WithIcon(icon string) StatOption {
	return func(p map[string]any) { p["icon"] = icon }
}

func WithColor(color string) StatOption {
	return func(p map[string]any) { p["color"] = color }
}

func WithTrend(value float64, label string) StatOption {
	return func(p map[string]any) {
		p["trend"] = map[string]any{"value": value, "label": label}
	}
}

// Alert creates an alert/notification component
func Alert(variant, message string) Component {
	return Component{
		Type: ComponentAlert,
		Props: map[string]any{
			"variant": variant, // info, success, warning, error
			"message": message,
		},
	}
}

// Grid creates a grid layout
func Grid(cols int, children ...Component) Component {
	return Component{
		Type:     ComponentGrid,
		Props:    map[string]any{"cols": cols},
		Children: children,
	}
}

// Row creates a flex row
func Row(children ...Component) Component {
	return Component{
		Type:     ComponentRow,
		Children: children,
	}
}

// Col creates a column in a grid/row
func Col(span int, children ...Component) Component {
	return Component{
		Type:     ComponentCol,
		Props:    map[string]any{"span": span},
		Children: children,
	}
}

// Text creates a text component
func Text(content string) Component {
	return Component{
		Type:  ComponentText,
		Props: map[string]any{"content": content},
	}
}

// Heading creates a heading component
func Heading(level int, content string) Component {
	return Component{
		Type:  ComponentHeading,
		Props: map[string]any{"level": level, "content": content},
	}
}

// Badge creates a badge component
func Badge(text, variant string) Component {
	return Component{
		Type:  ComponentBadge,
		Props: map[string]any{"text": text, "variant": variant},
	}
}

// ProgressOption is an option for the Progress component
type ProgressOption func(map[string]any)

// ProgressWithColor sets the color for the progress bar
func ProgressWithColor(color string) ProgressOption {
	return func(p map[string]any) { p["color"] = color }
}

// ProgressWithLabel sets a label for the progress bar
func ProgressWithLabel(label string) ProgressOption {
	return func(p map[string]any) { p["label"] = label }
}

// Progress creates a progress bar
func Progress(value, max int, opts ...ProgressOption) Component {
	props := map[string]any{"value": value, "max": max}
	for _, opt := range opts {
		opt(props)
	}
	return Component{
		Type:  ComponentProgress,
		Props: props,
	}
}

// List creates a list component
func List(items ...Component) Component {
	return Component{
		Type:     ComponentList,
		Children: items,
	}
}

// ListItem creates a list item
func ListItem(title string, subtitle string) Component {
	return Component{
		Type:  ComponentListItem,
		Props: map[string]any{"title": title, "subtitle": subtitle},
	}
}

// Divider creates a divider
func Divider() Component {
	return Component{Type: ComponentDivider}
}

// Empty creates an empty state
func Empty(message string, icon string) Component {
	return Component{
		Type:  ComponentEmpty,
		Props: map[string]any{"message": message, "icon": icon},
	}
}

// JSON displays raw JSON data
func JSON(data any) Component {
	return Component{
		Type:  ComponentJSON,
		Props: map[string]any{"data": data},
	}
}

// CodeBlock displays code with syntax highlighting
func CodeBlock(code, language string) Component {
	return Component{
		Type:  ComponentCodeBlock,
		Props: map[string]any{"code": code, "language": language},
	}
}

// ----- Chart Components -----

// ChartType represents the type of chart
type ChartType string

const (
	ChartLine  ChartType = "line"
	ChartBar   ChartType = "bar"
	ChartPie   ChartType = "pie"
	ChartArea  ChartType = "area"
	ChartDonut ChartType = "donut"
	ChartRadar ChartType = "radar"
)

// ChartSeries represents a data series for the chart
type ChartSeries struct {
	Name  string    `json:"name"`
	Data  []float64 `json:"data"`
	Color string    `json:"color,omitempty"`
}

// ChartConfig represents the configuration for a chart
type ChartConfig struct {
	Labels     []string      `json:"labels"`
	Series     []ChartSeries `json:"series"`
	ShowLegend bool          `json:"showLegend,omitempty"`
	ShowGrid   bool          `json:"showGrid,omitempty"`
	Stacked    bool          `json:"stacked,omitempty"`
	Height     int           `json:"height,omitempty"`
}

// ChartOption is an option for the Chart component
type ChartOption func(map[string]any)

// ChartWithLegend enables the legend
func ChartWithLegend() ChartOption {
	return func(p map[string]any) { p["showLegend"] = true }
}

// ChartWithGrid enables the grid
func ChartWithGrid() ChartOption {
	return func(p map[string]any) { p["showGrid"] = true }
}

// ChartWithStacked enables stacking for bar/area charts
func ChartWithStacked() ChartOption {
	return func(p map[string]any) { p["stacked"] = true }
}

// ChartWithHeight sets the chart height in pixels
func ChartWithHeight(height int) ChartOption {
	return func(p map[string]any) { p["height"] = height }
}

// ChartWithTitle sets the chart title
func ChartWithTitle(title string) ChartOption {
	return func(p map[string]any) { p["title"] = title }
}

// Chart creates a chart component
func Chart(chartType ChartType, labels []string, series []ChartSeries, opts ...ChartOption) Component {
	props := map[string]any{
		"type":       string(chartType),
		"labels":     labels,
		"series":     series,
		"showLegend": true,
		"showGrid":   true,
		"height":     300,
	}
	for _, opt := range opts {
		opt(props)
	}
	return Component{
		Type:  ComponentChart,
		Props: props,
	}
}

// LineChart creates a line chart
func LineChart(labels []string, series []ChartSeries, opts ...ChartOption) Component {
	return Chart(ChartLine, labels, series, opts...)
}

// BarChart creates a bar chart
func BarChart(labels []string, series []ChartSeries, opts ...ChartOption) Component {
	return Chart(ChartBar, labels, series, opts...)
}

// PieChart creates a pie chart
func PieChart(labels []string, values []float64, colors []string, opts ...ChartOption) Component {
	series := []ChartSeries{}
	for i, label := range labels {
		color := ""
		if i < len(colors) {
			color = colors[i]
		}
		val := float64(0)
		if i < len(values) {
			val = values[i]
		}
		series = append(series, ChartSeries{Name: label, Data: []float64{val}, Color: color})
	}
	return Chart(ChartPie, labels, series, opts...)
}

// AreaChart creates an area chart
func AreaChart(labels []string, series []ChartSeries, opts ...ChartOption) Component {
	return Chart(ChartArea, labels, series, opts...)
}

// DonutChart creates a donut chart
func DonutChart(labels []string, values []float64, colors []string, opts ...ChartOption) Component {
	series := []ChartSeries{}
	for i, label := range labels {
		color := ""
		if i < len(colors) {
			color = colors[i]
		}
		val := float64(0)
		if i < len(values) {
			val = values[i]
		}
		series = append(series, ChartSeries{Name: label, Data: []float64{val}, Color: color})
	}
	return Chart(ChartDonut, labels, series, opts...)
}
