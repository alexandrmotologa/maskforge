package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/alexandrmotologa/maskforge/internal/faker"
	"gopkg.in/yaml.v3"
)

// DefaultConfig returns an empty default configuration with a fresh random salt.
func DefaultConfig() *Config {
	saltBytes := make([]byte, 16)
	_, _ = rand.Read(saltBytes)
	return &Config{
		Version: "1.0",
		Salt:    hex.EncodeToString(saltBytes),
		Global: GlobalConfig{
			PreserveNulls: true,
			DefaultRule:   "",
		},
		Tables: make(map[string]TableConfig),
	}
}

// LoadFromFile parses a YAML configuration file from disk.
func LoadFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", filePath, err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses a YAML configuration from a byte array.
func LoadFromBytes(data []byte) (*Config, error) {
	cfg := DefaultConfig()
	if len(data) == 0 {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML configuration: %w", err)
	}

	if cfg.Salt == "" {
		saltBytes := make([]byte, 16)
		_, _ = rand.Read(saltBytes)
		cfg.Salt = hex.EncodeToString(saltBytes)
	}
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}
	if cfg.Tables == nil {
		cfg.Tables = make(map[string]TableConfig)
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateConfig verifies that all transformer rules referenced in the config exist in the registry.
func ValidateConfig(cfg *Config) error {
	reg := faker.NewRegistry()

	if cfg.Global.DefaultRule != "" {
		name, _ := faker.ParseRule(cfg.Global.DefaultRule)
		ctx := faker.TransformContext{
			Salt:       cfg.Salt,
			RawValue:   "test",
			RuleString: name,
		}
		if res := reg.Transform(ctx); res == "test" && name != "passthrough" && name != "ignore" {
			return fmt.Errorf("unknown global default rule: %s", cfg.Global.DefaultRule)
		}
	}

	for tbl, tcfg := range cfg.Tables {
		for col, ruleStr := range tcfg.Columns {
			name, _ := faker.ParseRule(ruleStr)
			ctx := faker.TransformContext{
				Salt:       cfg.Salt,
				Table:      tbl,
				Column:     col,
				RawValue:   "test",
				RuleString: name,
			}
			if res := reg.Transform(ctx); res == "test" && name != "passthrough" && name != "ignore" {
				return fmt.Errorf("unknown rule %q for column %s.%s", ruleStr, tbl, col)
			}
		}
	}

	return nil
}

// ToYAML serializes the configuration to clean YAML.
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}
