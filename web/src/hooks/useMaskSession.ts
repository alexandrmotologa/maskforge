import { useState, useEffect, useCallback } from 'react';
import { ScanReport, Config, TransformerMeta, PipelineMetrics } from '../types';
import { DEMO_SQL_DUMP } from '../data/demoSQL';

export function useMaskSession() {
  const [report, setReport] = useState<ScanReport | null>(null);
  const [config, setConfig] = useState<Config>({
    version: '1.0',
    salt: 'prod-safe-salt-' + Math.random().toString(36).substring(2, 10),
    global: {
      preserve_nulls: true,
      ignore_columns: ['id', 'created_at', 'updated_at'],
    },
    tables: {},
  });
  const [transformers, setTransformers] = useState<TransformerMeta[]>([]);
  const [sqlInput, setSqlInput] = useState<string>(DEMO_SQL_DUMP);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'tables' | 'editor' | 'diff' | 'referential' | 'simulator' | 'yaml'>('tables');
  const [selectedTable, setSelectedTable] = useState<string>('users');

  // Load available transformers on mount
  useEffect(() => {
    fetch('/api/transformers')
      .then((res) => (res.ok ? res.json() : []))
      .then((data: TransformerMeta[]) => setTransformers(data))
      .catch(() => {});
  }, []);

  // Run scan on initial mount with demo SQL
  const runScan = useCallback(async (customSQL?: string, customSalt?: string) => {
    setIsLoading(true);
    setError(null);
    const sqlToScan = customSQL !== undefined ? customSQL : sqlInput;
    const activeSalt = customSalt || config.salt;

    try {
      const res = await fetch('/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sql: sqlToScan,
          max_samples: 10,
          salt: activeSalt,
        }),
      });

      if (!res.ok) {
        throw new Error(`Scan failed: ${res.statusText}`);
      }

      const scanResult: ScanReport = await res.json();
      setReport(scanResult);

      // Auto-populate config tables with recommended rules
      const newTables: Record<string, { columns: Record<string, string>; ignore_columns: string[] }> = {};
      scanResult.tables.forEach((t) => {
        const cols: Record<string, string> = {};
        t.columns.forEach((c) => {
          if (c.recommended_rule && c.recommended_rule !== 'passthrough') {
            cols[c.name] = c.recommended_rule;
          }
        });
        if (Object.keys(cols).length > 0) {
          newTables[t.name] = {
            columns: cols,
            ignore_columns: ['id'],
          };
        }
      });

      setConfig((prev) => ({
        ...prev,
        salt: scanResult.generation_salt || prev.salt,
        tables: newTables,
      }));

      if (scanResult.tables.length > 0 && !scanResult.tables.some((t) => t.name === selectedTable)) {
        setSelectedTable(scanResult.tables[0].name);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Unknown scanning error';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, [sqlInput, config.salt, selectedTable]);

  // Initial trigger
  useEffect(() => {
    runScan();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const updateColumnRule = (tableName: string, columnName: string, newRule: string) => {
    setConfig((prev) => {
      const currentTable = prev.tables[tableName] || { columns: {}, ignore_columns: [] };
      const currentCols = { ...(currentTable.columns || {}) };

      if (!newRule || newRule === 'passthrough') {
        delete currentCols[columnName];
      } else {
        currentCols[columnName] = newRule;
      }

      return {
        ...prev,
        tables: {
          ...prev.tables,
          [tableName]: {
            ...currentTable,
            columns: currentCols,
          },
        },
      };
    });
  };

  const updateSalt = (newSalt: string) => {
    setConfig((prev) => ({ ...prev, salt: newSalt }));
    runScan(sqlInput, newSalt);
  };

  const simulateStream = async (sql: string, configYAML: string): Promise<{ maskedSQL: string; metrics: PipelineMetrics }> => {
    const res = await fetch('/api/stream/simulate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sql, config_yaml: configYAML }),
    });
    if (!res.ok) {
      throw new Error(`Simulation error: ${res.statusText}`);
    }
    const data = await res.json();
    return {
      maskedSQL: data.masked_sql,
      metrics: data.metrics,
    };
  };

  const testReferentialMatch = async (
    rawValue: string,
    rule: string,
    tableA: string,
    colA: string,
    tableB: string,
    colB: string
  ) => {
    const res = await fetch('/api/referential/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        salt: config.salt,
        raw_value: rawValue,
        rule,
        table_a: tableA,
        column_a: colA,
        table_b: tableB,
        column_b: colB,
      }),
    });
    if (!res.ok) {
      throw new Error(`Referential test error: ${res.statusText}`);
    }
    return res.json();
  };

  return {
    report,
    config,
    transformers,
    sqlInput,
    setSqlInput,
    isLoading,
    error,
    activeTab,
    setActiveTab,
    selectedTable,
    setSelectedTable,
    runScan,
    updateColumnRule,
    updateSalt,
    simulateStream,
    testReferentialMatch,
  };
}
