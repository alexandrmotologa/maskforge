package discovery

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/alexandrmotologa/maskforge/internal/config"
)

// GenerateConfigFromReport produces a populated mask.yaml configuration from a PII audit report.
func GenerateConfigFromReport(report *ScanReport) *config.Config {
	salt := report.GenerationSalt
	if salt == "" {
		saltBytes := make([]byte, 16)
		_, _ = rand.Read(saltBytes)
		salt = hex.EncodeToString(saltBytes)
	}

	cfg := &config.Config{
		Version: "1.0",
		Salt:    salt,
		Global: config.GlobalConfig{
			PreserveNulls: true,
			IgnoreColumns: []string{"id", "created_at", "updated_at", "deleted_at"},
		},
		Tables: make(map[string]config.TableConfig),
	}

	for _, tbl := range report.Tables {
		tc := config.TableConfig{
			Columns:       make(map[string]string),
			IgnoreColumns: []string{},
		}

		for _, col := range tbl.Columns {
			if col.RecommendedRule != "" && col.RecommendedRule != "passthrough" {
				tc.Columns[col.Name] = col.RecommendedRule
			}
		}

		if len(tc.Columns) > 0 {
			cfg.Tables[tbl.Name] = tc
		}
	}

	return cfg
}
