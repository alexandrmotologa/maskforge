package config

import (
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/faker"
)

// Config represents the top-level declarative mask.yaml structure.
type Config struct {
	Version string                 `yaml:"version" json:"version"`
	Salt    string                 `yaml:"salt" json:"salt"`
	Global  GlobalConfig           `yaml:"global,omitempty" json:"global,omitempty"`
	Tables  map[string]TableConfig `yaml:"tables" json:"tables"`
}

// GlobalConfig provides defaults across all tables and columns.
type GlobalConfig struct {
	PreserveNulls bool     `yaml:"preserve_nulls" json:"preserve_nulls"`
	DefaultRule   string   `yaml:"default_rule,omitempty" json:"default_rule,omitempty"`
	IgnoreTables  []string `yaml:"ignore_tables,omitempty" json:"ignore_tables,omitempty"`
	IgnoreColumns []string `yaml:"ignore_columns,omitempty" json:"ignore_columns,omitempty"`
	SampleRate    float64  `yaml:"sample_rate,omitempty" json:"sample_rate,omitempty"` // 0.0 to 1.0
}

// TableConfig defines masking specifications for a specific table.
type TableConfig struct {
	Columns       map[string]string `yaml:"columns,omitempty" json:"columns,omitempty"`
	Conditions    map[string]string `yaml:"conditions,omitempty" json:"conditions,omitempty"`
	IgnoreColumns []string          `yaml:"ignore_columns,omitempty" json:"ignore_columns,omitempty"`
	Truncate      bool              `yaml:"truncate,omitempty" json:"truncate,omitempty"`
	SampleLimit   int64             `yaml:"sample_limit,omitempty" json:"sample_limit,omitempty"`
}

// CleanTableName strips schema qualifiers (e.g. "public.users" -> "users", "`db`.`users`" -> "users").
func CleanTableName(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'` + "`")
	if idx := strings.LastIndex(raw, "."); idx != -1 {
		raw = raw[idx+1:]
	}
	raw = strings.Trim(raw, `"'` + "`")
	return strings.ToLower(raw)
}

// CleanColumnName strips quotes and backticks from column names.
func CleanColumnName(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'` + "`")
	return strings.ToLower(raw)
}

// GetTableConfig finds the configuration for a given table name.
func (c *Config) GetTableConfig(tableName string) (TableConfig, bool) {
	if c.Tables == nil {
		return TableConfig{}, false
	}
	clean := CleanTableName(tableName)

	// Direct match
	if tc, ok := c.Tables[clean]; ok {
		return tc, true
	}

	// Case-insensitive / qualified match
	for k, v := range c.Tables {
		if CleanTableName(k) == clean {
			return v, true
		}
	}

	return TableConfig{}, false
}

// IsTableIgnored returns true if the table is explicitly ignored in GlobalConfig.
func (c *Config) IsTableIgnored(tableName string) bool {
	clean := CleanTableName(tableName)
	for _, ign := range c.Global.IgnoreTables {
		if CleanTableName(ign) == clean {
			return true
		}
	}
	return false
}

// GetColumnRule resolves the masking rule string for a given table and column.
func (c *Config) GetColumnRule(tableName, columnName string) (rule string, shouldMask bool) {
	if c.IsTableIgnored(tableName) {
		return "passthrough", false
	}

	cleanCol := CleanColumnName(columnName)

	// Check global ignore columns
	for _, ign := range c.Global.IgnoreColumns {
		if CleanColumnName(ign) == cleanCol {
			return "passthrough", false
		}
	}

	tc, found := c.GetTableConfig(tableName)
	if !found {
		if c.Global.DefaultRule != "" {
			return c.Global.DefaultRule, true
		}
		return "passthrough", false
	}

	// Check table-level ignored columns
	for _, ign := range tc.IgnoreColumns {
		if CleanColumnName(ign) == cleanCol {
			return "passthrough", false
		}
	}

	// Match column rules
	if tc.Columns != nil {
		if r, ok := tc.Columns[cleanCol]; ok && r != "" {
			return r, true
		}
		for k, r := range tc.Columns {
			if CleanColumnName(k) == cleanCol && r != "" {
				return r, true
			}
		}
	}

	if c.Global.DefaultRule != "" {
		return c.Global.DefaultRule, true
	}

	return "passthrough", false
}

// GetColumnCondition returns the optional conditional expression for masking a column.
func (c *Config) GetColumnCondition(tableName, columnName string) string {
	tc, found := c.GetTableConfig(tableName)
	cleanCol := CleanColumnName(columnName)

	if found && tc.Conditions != nil {
		if cond, ok := tc.Conditions[cleanCol]; ok && cond != "" {
			return cond
		}
		for k, cond := range tc.Conditions {
			if CleanColumnName(k) == cleanCol && cond != "" {
				return cond
			}
		}
	}

	// Also check if specified in rule parameters when="..."
	rule, shouldMask := c.GetColumnRule(tableName, columnName)
	if shouldMask {
		_, params := faker.ParseRule(rule)
		if whenCond, ok := params["when"]; ok && whenCond != "" {
			return whenCond
		}
	}

	return ""
}
