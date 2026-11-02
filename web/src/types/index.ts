export type RiskLevel = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';

export interface DiscoveredColumn {
  name: string;
  detected_type: string;
  risk: RiskLevel;
  compliance: string[];
  recommended_rule: string;
  confidence: number;
  sample_values: string[];
  masked_sample_values: string[];
}

export interface DiscoveredTable {
  name: string;
  columns: DiscoveredColumn[];
  total_rows: number;
  sample_rows: string[][];
}

export interface ScanReport {
  tables: DiscoveredTable[];
  total_tables: number;
  total_columns: number;
  pii_columns: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  estimated_risk: RiskLevel;
  generation_salt: string;
}

export interface TransformerMeta {
  name: string;
  category: string;
  description: string;
  example: string;
  parameters: string[];
}

export interface PipelineMetrics {
  LinesProcessed: number;
  RowsMasked: number;
  BytesRead: number;
  BytesWritten: number;
  Duration: number;
  RowsPerSec: number;
  MBPerSec: number;
}

export interface TableConfig {
  columns?: Record<string, string>;
  ignore_columns?: string[];
  truncate?: boolean;
}

export interface Config {
  version: string;
  salt: string;
  global?: {
    preserve_nulls?: boolean;
    default_rule?: string;
    ignore_tables?: string[];
    ignore_columns?: string[];
  };
  tables: Record<string, TableConfig>;
}
