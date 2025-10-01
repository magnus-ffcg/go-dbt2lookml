package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ParsedLookMLFile represents a parsed LookML file
type ParsedLookMLFile struct {
	Explores []ParsedExplore `json:"explores"`
	Views    []ParsedView    `json:"views"`
}

// ParsedExplore represents a parsed explore block
type ParsedExplore struct {
	Name   string       `json:"name"`
	Hidden *bool        `json:"hidden,omitempty"`
	Joins  []ParsedJoin `json:"joins"`
}

// ParsedJoin represents a parsed join block
type ParsedJoin struct {
	Name         string  `json:"name"`
	ViewLabel    *string `json:"view_label,omitempty"`
	SQL          string  `json:"sql"`
	Relationship string  `json:"relationship"`
}

// ParsedView represents a parsed view block
type ParsedView struct {
	Name            string                 `json:"name"`
	SQLTableName    string                 `json:"sql_table_name"`
	Label           *string                `json:"label,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Dimensions      []ParsedDimension      `json:"dimensions"`
	DimensionGroups []ParsedDimensionGroup `json:"dimension_groups"`
	Measures        []ParsedMeasure        `json:"measures"`
}

// ParsedDimension represents a parsed dimension
type ParsedDimension struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	SQL         string  `json:"sql"`
	Description *string `json:"description,omitempty"`
	Hidden      *bool   `json:"hidden,omitempty"`
	Label       *string `json:"label,omitempty"`
}

// ParsedDimensionGroup represents a parsed dimension_group
type ParsedDimensionGroup struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	SQL        string   `json:"sql"`
	Timeframes []string `json:"timeframes,omitempty"`
	ConvertTZ  *bool    `json:"convert_tz,omitempty"`
	Datatype   *string  `json:"datatype,omitempty"`
}

// ParsedMeasure represents a parsed measure
type ParsedMeasure struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// LookMLParser handles parsing of LookML files
type LookMLParser struct {
	lines []string
	pos   int
}

// NewLookMLParser creates a new parser instance
func NewLookMLParser() *LookMLParser {
	return &LookMLParser{}
}

// ParseFile parses a LookML file and returns structured data
func (p *LookMLParser) ParseFile(filePath string) (*ParsedLookMLFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read all lines
	scanner := bufio.NewScanner(file)
	p.lines = []string{}
	for scanner.Scan() {
		p.lines = append(p.lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	p.pos = 0
	return p.parseFile()
}

// parseFile parses the entire file
func (p *LookMLParser) parseFile() (*ParsedLookMLFile, error) {
	result := &ParsedLookMLFile{
		Explores: []ParsedExplore{},
		Views:    []ParsedView{},
	}

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse explore blocks
		if strings.HasPrefix(line, "explore:") {
			explore, err := p.parseExplore()
			if err != nil {
				return nil, err
			}
			result.Explores = append(result.Explores, *explore)
			continue
		}

		// Parse view blocks
		if strings.HasPrefix(line, "view:") {
			view, err := p.parseView()
			if err != nil {
				return nil, err
			}
			result.Views = append(result.Views, *view)
			continue
		}

		p.pos++
	}

	return result, nil
}

// parseExplore parses an explore block
func (p *LookMLParser) parseExplore() (*ParsedExplore, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract explore name
	nameRegex := regexp.MustCompile(`explore:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid explore syntax at line %d: %s", p.pos+1, line)
	}

	explore := &ParsedExplore{
		Name:  strings.TrimSpace(matches[1]),
		Joins: []ParsedJoin{},
	}

	p.pos++ // Move past explore line

	// Parse explore content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse hidden property
		if strings.HasPrefix(line, "hidden:") {
			hidden := strings.Contains(line, "yes")
			explore.Hidden = &hidden
			p.pos++
			continue
		}

		// Parse join blocks
		if strings.HasPrefix(line, "join:") {
			join, err := p.parseJoin()
			if err != nil {
				return nil, err
			}
			explore.Joins = append(explore.Joins, *join)
			continue
		}

		p.pos++
	}

	return explore, nil
}

// parseJoin parses a join block
func (p *LookMLParser) parseJoin() (*ParsedJoin, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract join name
	nameRegex := regexp.MustCompile(`join:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid join syntax at line %d: %s", p.pos+1, line)
	}

	join := &ParsedJoin{
		Name: strings.TrimSpace(matches[1]),
	}

	p.pos++ // Move past join line

	// Parse join content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse join properties
		if strings.HasPrefix(line, "view_label:") {
			value := p.extractStringValue(line, "view_label:")
			join.ViewLabel = &value
		} else if strings.HasPrefix(line, "sql:") {
			join.SQL = p.extractStringValue(line, "sql:")
		} else if strings.HasPrefix(line, "relationship:") {
			join.Relationship = p.extractStringValue(line, "relationship:")
		}

		p.pos++
	}

	return join, nil
}

// parseView parses a view block
func (p *LookMLParser) parseView() (*ParsedView, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract view name
	nameRegex := regexp.MustCompile(`view:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid view syntax at line %d: %s", p.pos+1, line)
	}

	view := &ParsedView{
		Name:            strings.TrimSpace(matches[1]),
		Dimensions:      []ParsedDimension{},
		DimensionGroups: []ParsedDimensionGroup{},
		Measures:        []ParsedMeasure{},
	}

	p.pos++ // Move past view line

	// Parse view content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse view properties
		if strings.HasPrefix(line, "sql_table_name:") {
			view.SQLTableName = p.extractStringValue(line, "sql_table_name:")
		} else if strings.HasPrefix(line, "label:") {
			value := p.extractStringValue(line, "label:")
			view.Label = &value
		} else if strings.HasPrefix(line, "description:") {
			value := p.extractStringValue(line, "description:")
			view.Description = &value
		} else if strings.HasPrefix(line, "dimension:") {
			dimension, err := p.parseDimension()
			if err != nil {
				return nil, err
			}
			view.Dimensions = append(view.Dimensions, *dimension)
			continue
		} else if strings.HasPrefix(line, "dimension_group:") {
			dimensionGroup, err := p.parseDimensionGroup()
			if err != nil {
				return nil, err
			}
			view.DimensionGroups = append(view.DimensionGroups, *dimensionGroup)
			continue
		} else if strings.HasPrefix(line, "measure:") {
			measure, err := p.parseMeasure()
			if err != nil {
				return nil, err
			}
			view.Measures = append(view.Measures, *measure)
			continue
		}

		p.pos++
	}

	return view, nil
}

// parseDimension parses a dimension block
func (p *LookMLParser) parseDimension() (*ParsedDimension, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract dimension name
	nameRegex := regexp.MustCompile(`dimension:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid dimension syntax at line %d: %s", p.pos+1, line)
	}

	dimension := &ParsedDimension{
		Name: strings.TrimSpace(matches[1]),
	}

	p.pos++ // Move past dimension line

	// Parse dimension content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse dimension properties
		if strings.HasPrefix(line, "type:") {
			dimension.Type = p.extractStringValue(line, "type:")
		} else if strings.HasPrefix(line, "sql:") {
			dimension.SQL = p.extractStringValue(line, "sql:")
		} else if strings.HasPrefix(line, "description:") {
			value := p.extractStringValue(line, "description:")
			dimension.Description = &value
		} else if strings.HasPrefix(line, "label:") {
			value := p.extractStringValue(line, "label:")
			dimension.Label = &value
		} else if strings.HasPrefix(line, "hidden:") {
			hidden := strings.Contains(line, "yes")
			dimension.Hidden = &hidden
		}

		p.pos++
	}

	return dimension, nil
}

// parseDimensionGroup parses a dimension_group block
func (p *LookMLParser) parseDimensionGroup() (*ParsedDimensionGroup, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract dimension_group name
	nameRegex := regexp.MustCompile(`dimension_group:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid dimension_group syntax at line %d: %s", p.pos+1, line)
	}

	dimensionGroup := &ParsedDimensionGroup{
		Name:       strings.TrimSpace(matches[1]),
		Timeframes: []string{},
	}

	p.pos++ // Move past dimension_group line

	// Parse dimension_group content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse dimension_group properties
		if strings.HasPrefix(line, "type:") {
			dimensionGroup.Type = p.extractStringValue(line, "type:")
		} else if strings.HasPrefix(line, "sql:") {
			dimensionGroup.SQL = p.extractStringValue(line, "sql:")
		} else if strings.HasPrefix(line, "timeframes:") {
			dimensionGroup.Timeframes = p.extractArrayValue(line, "timeframes:")
		} else if strings.HasPrefix(line, "convert_tz:") {
			convertTZ := !strings.Contains(line, "no")
			dimensionGroup.ConvertTZ = &convertTZ
		} else if strings.HasPrefix(line, "datatype:") {
			value := p.extractStringValue(line, "datatype:")
			dimensionGroup.Datatype = &value
		}

		p.pos++
	}

	return dimensionGroup, nil
}

// parseMeasure parses a measure block
func (p *LookMLParser) parseMeasure() (*ParsedMeasure, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Extract measure name
	nameRegex := regexp.MustCompile(`measure:\s*([^{]+)\s*{`)
	matches := nameRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid measure syntax at line %d: %s", p.pos+1, line)
	}

	measure := &ParsedMeasure{
		Name: strings.TrimSpace(matches[1]),
	}

	p.pos++ // Move past measure line

	// Parse measure content
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "}" {
			p.pos++
			break
		}

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Parse measure properties
		if strings.HasPrefix(line, "type:") {
			measure.Type = p.extractStringValue(line, "type:")
		} else if strings.HasPrefix(line, "label:") {
			measure.Label = p.extractStringValue(line, "label:")
		}

		p.pos++
	}

	return measure, nil
}

// extractStringValue extracts a string value from a LookML property line
func (p *LookMLParser) extractStringValue(line, prefix string) string {
	// Remove prefix and trim
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))

	// Remove quotes and semicolons
	value = strings.Trim(value, `"'`)
	value = strings.TrimSuffix(value, ";;")
	value = strings.TrimSpace(value)

	return value
}

// extractArrayValue extracts an array value from a LookML property line
func (p *LookMLParser) extractArrayValue(line, prefix string) []string {
	// Remove prefix and trim
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))

	// Remove brackets and semicolons
	value = strings.Trim(value, "[]")
	value = strings.TrimSuffix(value, ";;")
	value = strings.TrimSpace(value)

	if value == "" {
		return []string{}
	}

	// Split by comma and clean up each item
	items := strings.Split(value, ",")
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = strings.TrimSpace(strings.Trim(item, `"'`))
	}

	return result
}
