import React from 'react';
import { Table as TableIcon, ShieldAlert, Sparkles, CheckCircle2, Lock } from 'lucide-react';
import { DiscoveredTable, RiskLevel, TransformerMeta } from '../types';

interface SchemaTableViewProps {
  tables: DiscoveredTable[];
  selectedTable: string;
  onSelectTable: (name: string) => void;
  columnRules: Record<string, Record<string, string>>;
  onUpdateRule: (table: string, column: string, rule: string) => void;
  transformers: TransformerMeta[];
}

export const SchemaTableView: React.FC<SchemaTableViewProps> = ({
  tables,
  selectedTable,
  onSelectTable,
  columnRules,
  onUpdateRule,
  transformers,
}) => {
  const currentTable = tables.find((t) => t.name === selectedTable) || tables[0];

  const getRiskBadge = (risk: RiskLevel) => {
    switch (risk) {
      case 'CRITICAL':
        return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/40">CRITICAL</span>;
      case 'HIGH':
        return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/40">HIGH</span>;
      case 'MEDIUM':
        return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-sky-500/20 text-sky-300 border border-sky-500/40">MEDIUM</span>;
      default:
        return <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">LOW</span>;
    }
  };

  if (!tables || tables.length === 0) {
    return (
      <div className="bg-[#0f172a]/60 border border-[#1e293b] rounded-2xl p-12 text-center text-slate-400">
        <TableIcon className="w-12 h-12 mx-auto mb-3 text-slate-600 animate-pulse" />
        <p className="text-base font-medium text-slate-300">No database tables detected</p>
        <p className="text-xs text-slate-500 mt-1">Upload or paste an SQL dump in the Simulator tab to audit PII.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
      {/* Table Selector Sidebar */}
      <div className="lg:col-span-1 space-y-2">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 px-1 mb-2">
          Discovered Tables ({tables.length})
        </h3>
        <div className="space-y-1.5">
          {tables.map((table) => {
            const isSelected = table.name === currentTable?.name;
            const piiCols = table.columns.filter((c) => c.risk !== 'LOW').length;
            const hasCritical = table.columns.some((c) => c.risk === 'CRITICAL');

            return (
              <button
                key={table.name}
                onClick={() => onSelectTable(table.name)}
                className={`w-full text-left p-3 rounded-xl border transition flex items-center justify-between cursor-pointer ${
                  isSelected
                    ? 'bg-[#172033] border-emerald-500/50 shadow-md shadow-emerald-500/5 text-white'
                    : 'bg-[#0f172a]/70 border-[#1e293b] hover:bg-[#151f33] text-slate-300'
                }`}
              >
                <div className="flex items-center gap-2.5">
                  <TableIcon className={`w-4 h-4 ${isSelected ? 'text-emerald-400' : 'text-slate-500'}`} />
                  <div>
                    <span className="font-mono text-sm font-semibold block">{table.name}</span>
                    <span className="text-[11px] text-slate-500">{table.columns.length} columns &bull; {table.total_rows} rows</span>
                  </div>
                </div>

                {piiCols > 0 && (
                  <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border flex items-center gap-1 ${
                    hasCritical
                      ? 'bg-rose-500/20 text-rose-300 border-rose-500/40'
                      : 'bg-amber-500/20 text-amber-300 border-amber-500/40'
                  }`}>
                    <ShieldAlert className="w-2.5 h-2.5" />
                    {piiCols} PII
                  </span>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* Table Detail & Column PII Matrix */}
      <div className="lg:col-span-3">
        {currentTable && (
          <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl overflow-hidden shadow-xl">
            {/* Table Header Banner */}
            <div className="p-5 border-b border-[#1e293b] bg-[#121c2e]/70 flex flex-wrap items-center justify-between gap-4">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-lg font-bold text-white font-mono">{currentTable.name}</h2>
                  <span className="text-xs text-slate-400">({currentTable.columns.length} columns)</span>
                </div>
                <p className="text-xs text-slate-400 mt-0.5">
                  Inspect detected sensitive attributes and configure deterministic faker transformations.
                </p>
              </div>

              <div className="flex items-center gap-2">
                <span className="text-xs font-mono text-slate-400 bg-[#090d16] px-2.5 py-1 rounded-md border border-[#1e293b]">
                  {currentTable.total_rows} estimated rows
                </span>
              </div>
            </div>

            {/* Column Matrix Table */}
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-[#1e293b] bg-[#0c1220] text-slate-400 text-[11px] uppercase tracking-wider font-semibold">
                    <th className="py-3 px-4">Column Name</th>
                    <th className="py-3 px-4">Classification</th>
                    <th className="py-3 px-4">Risk Level</th>
                    <th className="py-3 px-4">Masking Strategy</th>
                    <th className="py-3 px-4">Sample Before &rarr; After</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#1e293b]">
                  {currentTable.columns.map((col) => {
                    const activeRule =
                      columnRules[currentTable.name]?.[col.name] !== undefined
                        ? columnRules[currentTable.name][col.name]
                        : col.recommended_rule;

                    const isMasked = activeRule && activeRule !== 'passthrough';
                    const sampleOrig = col.sample_values[0] || '—';
                    const sampleMasked = col.masked_sample_values[0] || sampleOrig;

                    return (
                      <tr key={col.name} className="hover:bg-[#131b2e]/50 transition">
                        {/* Column Name */}
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <span className="font-mono font-medium text-slate-200">{col.name}</span>
                            {col.compliance.length > 0 && (
                              <span className="text-[10px] text-teal-400 bg-teal-500/10 px-1.5 py-0.2 rounded border border-teal-500/20" title={col.compliance.join(', ')}>
                                {col.compliance[0]}
                              </span>
                            )}
                          </div>
                        </td>

                        {/* Classification */}
                        <td className="py-3 px-4">
                          <span className="text-slate-300 flex items-center gap-1.5">
                            {isMasked ? <Lock className="w-3 h-3 text-emerald-400" /> : null}
                            {col.detected_type}
                          </span>
                        </td>

                        {/* Risk Level */}
                        <td className="py-3 px-4">{getRiskBadge(col.risk)}</td>

                        {/* Strategy Selector */}
                        <td className="py-3 px-4">
                          <select
                            value={activeRule || 'passthrough'}
                            onChange={(e) => onUpdateRule(currentTable.name, col.name, e.target.value)}
                            className="bg-[#121929] border border-[#2b3a56] text-emerald-300 font-mono text-xs rounded-lg px-2.5 py-1 focus:outline-none focus:border-emerald-500 transition cursor-pointer"
                          >
                            <option value="passthrough">passthrough (unmasked)</option>
                            <optgroup label="Recommended for this type">
                              {col.recommended_rule && col.recommended_rule !== 'passthrough' && (
                                <option value={col.recommended_rule}>{col.recommended_rule} ★</option>
                              )}
                            </optgroup>
                            <optgroup label="All Transformers">
                              {transformers.map((t) => (
                                <option key={t.name} value={t.name}>
                                  {t.name}
                                </option>
                              ))}
                            </optgroup>
                          </select>
                        </td>

                        {/* Live Sample Preview */}
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2 font-mono text-[11px]">
                            <span className="text-slate-400 line-through truncate max-w-[120px]" title={sampleOrig}>
                              {sampleOrig}
                            </span>
                            <span className="text-slate-600">&rarr;</span>
                            <span
                              className={`truncate max-w-[160px] font-semibold ${
                                isMasked ? 'text-emerald-300 bg-emerald-500/10 px-1.5 py-0.5 rounded border border-emerald-500/20' : 'text-slate-400'
                              }`}
                              title={sampleMasked}
                            >
                              {sampleMasked}
                            </span>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Bottom info banner */}
            <div className="p-4 bg-[#0a0f1d] border-t border-[#1e293b] flex items-center justify-between text-xs text-slate-400">
              <span className="flex items-center gap-1.5 text-emerald-400 font-medium">
                <CheckCircle2 className="w-4 h-4" />
                Deterministic Salt preserves relational foreign keys across tables
              </span>
              <span className="flex items-center gap-1 text-slate-500">
                <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                Automatic format preservation (RFC 5322 & Luhn Mod-10)
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
