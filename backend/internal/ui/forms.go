package ui

// Form creates a form component
func Form(id string, fields []FormField, opts ...FormOption) Component {
	props := map[string]any{
		"id":     id,
		"fields": fields,
	}
	for _, opt := range opts {
		opt(props)
	}
	return Component{
		Type:  ComponentForm,
		Props: props,
	}
}

type FormOption func(map[string]any)

func WithSubmitURL(url, method string) FormOption {
	return func(p map[string]any) {
		p["submitUrl"] = url
		p["method"] = method
	}
}

func WithSubmitLabel(label string) FormOption {
	return func(p map[string]any) { p["submitLabel"] = label }
}

func WithCancelURL(url string) FormOption {
	return func(p map[string]any) { p["cancelUrl"] = url }
}

// FormField represents a form field
type FormField struct {
	Name        string         `json:"name"`
	Label       string         `json:"label"`
	Type        string         `json:"type"` // text, email, password, number, select, checkbox, textarea, date, etc.
	Placeholder string         `json:"placeholder,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Disabled    bool           `json:"disabled,omitempty"`
	ReadOnly    bool           `json:"readOnly,omitempty"`
	Default     any            `json:"default,omitempty"`
	Value       any            `json:"value,omitempty"`
	Options     []SelectOption `json:"options,omitempty"` // for select, radio, checkbox-group
	Validation  *Validation    `json:"validation,omitempty"`
	Help        string         `json:"help,omitempty"`
	ClassName   string         `json:"className,omitempty"`
	ColSpan     int            `json:"colSpan,omitempty"` // for grid layout
}

// SelectOption for select/radio/checkbox fields
type SelectOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
}

// Validation rules for a field
type Validation struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	MinLength *int     `json:"minLength,omitempty"`
	MaxLength *int     `json:"maxLength,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Message   string   `json:"message,omitempty"` // custom error message
}

// ----- Field Builders -----

func TextField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "text"}
}

func EmailField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "email"}
}

func PasswordField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "password"}
}

func NumberField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "number"}
}

func TextareaField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "textarea"}
}

func SelectField(name, label string, options []SelectOption) FormField {
	return FormField{Name: name, Label: label, Type: "select", Options: options}
}

func CheckboxField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "checkbox"}
}

func DateField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "date"}
}

func DateTimeField(name, label string) FormField {
	return FormField{Name: name, Label: label, Type: "datetime-local"}
}

func HiddenField(name string, value any) FormField {
	return FormField{Name: name, Type: "hidden", Value: value}
}

// Field modifiers (chainable pattern)
func (f FormField) WithPlaceholder(p string) FormField {
	f.Placeholder = p
	return f
}

func (f FormField) WithRequired() FormField {
	f.Required = true
	return f
}

func (f FormField) WithDefault(v any) FormField {
	f.Default = v
	return f
}

func (f FormField) WithValue(v any) FormField {
	f.Value = v
	return f
}

func (f FormField) WithHelp(h string) FormField {
	f.Help = h
	return f
}

func (f FormField) WithDisabled() FormField {
	f.Disabled = true
	return f
}

func (f FormField) WithReadOnly() FormField {
	f.ReadOnly = true
	return f
}

func (f FormField) WithColSpan(span int) FormField {
	f.ColSpan = span
	return f
}

func (f FormField) WithValidation(v Validation) FormField {
	f.Validation = &v
	return f
}
