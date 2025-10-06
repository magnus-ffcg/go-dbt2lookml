package generators

import (
	"bytes"
	"text/template"
)

// Global template registry
var lookMLTemplates *template.Template

const (
	viewTemplate = `view: {{ .Name }} {
  sql_table_name: {{ .SQLTableName }} ;;
{{- if .Label }}
  label: "{{ .Label }}"
{{- end }}
{{- if .Description }}
  description: "{{ .Description }}"
{{- end }}

{{- range .Dimensions }}
  {{- template "dimension" . -}}
{{- end }}

{{- range .DimensionGroups }}
  {{- template "dimensionGroup" . -}}
{{- end }}

{{- range .Measures }}
  {{- template "measure" . -}}
{{- end }}
}
`

	dimensionTemplate = `
  dimension: {{ .Name }} {
    type: {{ .Type }}
    sql: {{ .SQL }} ;;
    {{- if .GroupLabel }}
    group_label: "{{ .GroupLabel }}"
    {{- end }}
    {{- if .GroupItemLabel }}
    group_item_label: "{{ .GroupItemLabel }}"
    {{- end }}
    {{- if .Label }}
    label: "{{ .Label }}"
    {{- end }}
    {{- if .Description }}
    description: "{{ .Description }}"
    {{- end }}
    {{- if .Hidden }}
    hidden: yes
    {{- end }}
  }
`

	dimensionGroupTemplate = `
  dimension_group: {{ .Name }} {
    type: time
    timeframes: [
      raw,
      time,
      date,
      week,
      month,
      quarter,
      year
    ]
    sql: {{ .SQL }} ;;
    {{- if .Label }}
    label: "{{ .Label }}"
    {{- end }}
    {{- if .Description }}
    description: "{{ .Description }}"
    {{- end }}
  }
`

	measureTemplate = `
  measure: {{ .Name }} {
    type: {{ .Type }}
    {{- if .SQL }}
    sql: {{ .SQL }} ;;
    {{- end }}
    {{- if .Label }}
    label: "{{ .Label }}"
    {{- end }}
    {{- if .Description }}
    description: "{{ .Description }}"
    {{- end }}
    {{- if .ValueFormatName }}
    value_format_name: {{ .ValueFormatName }}
    {{- end }}
  }
`

	exploreTemplate = `
# Un-hide and use this explore, or copy the joins into another explore, to get all the fully nested relationships from this view
explore: {{ .Name }} {
  hidden: yes
{{- range .Joins }}
  {{- template "join" . -}}
{{- end }}
}
`

	joinTemplate = `
  join: {{ .Name }} {
    {{- if .ViewLabel }}
    view_label: "{{ .ViewLabel }}"
    {{- end }}
    {{- if .SQL }}
    sql: {{ .SQL }} ;;
    {{- end }}
    {{- if .Relationship }}
    relationship: {{ .Relationship }}
    {{- end }}
  }
`
)

func init() {
	lookMLTemplates = template.Must(template.New("view").Parse(viewTemplate))
	template.Must(lookMLTemplates.New("dimension").Parse(dimensionTemplate))
	template.Must(lookMLTemplates.New("dimensionGroup").Parse(dimensionGroupTemplate))
	template.Must(lookMLTemplates.New("measure").Parse(measureTemplate))
	template.Must(lookMLTemplates.New("explore").Parse(exploreTemplate))
	template.Must(lookMLTemplates.New("join").Parse(joinTemplate))
}

// renderLookML executes the given template with the provided data.
func renderLookML(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := lookMLTemplates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
