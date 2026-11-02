import React, { useState } from 'react';
import { Sliders, Play, CheckCircle, Code, HelpCircle } from 'lucide-react';
import { DiscoveredTable, TransformerMeta } from '../types';

interface MaskRuleEditorProps {
  tables: DiscoveredTable[];
  selectedTable: string;
  onSelectTable: (name: string) => void;
  columnRules: Record<string, Record<string, string>>;
  onUpdateRule: (table: string, column: string, rule: string) => void;
  transformers: TransformerMeta[];
  salt: string;
}

export const MaskRuleEditor: React.FC<MaskRuleEditorProps> = ({
  tables,
  selectedTable,
  onSelectTable,
  columnRules,
  onUpdateRule,
  transformers,
  salt,
}) => {
  const currentTable = tables.find((t) => t.name === selectedTable) || tables[0];
  const [selectedColumn, setSelectedColumn] = useState<string>(
    currentTable?.columns[0]?.name || ''
  );
  const [testInput, setTestInput] = useState<string>('alice.test@enterprise.com');
  const [testOutput, setTestOutput] = useState<string | null>(null);
  const [isTesting, setIsTesting] = useState<boolean>(false);

  const activeRule =
    columnRules[currentTable?.name]?.[selectedColumn] ||
    currentTable?.columns.find((c) => c.name === selectedColumn)?.recommended_rule ||
    'passthrough';

  const selectedMeta = transformers.find((t) => activeRule.startsWith(t.name));

  const handleTestPreview = async () => {
    setIsTesting(true);
    try {
      const res = await fetch('/api/transform/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          salt,
          table: currentTable?.name,
          column: selectedColumn,
          raw_value: testInput,
          rule: activeRule,
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setTestOutput(data.masked_value);
      }
    } catch {
      setTestOutput('Error testing transformation');
    } finally {
      setIsTesting(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Controls Bar */}
      <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[#1e293b] pb-5 mb-5">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-teal-500/10 border border-teal-500/20 text-teal-400 flex items-center justify-center">
              <Sliders className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">Rule Configurator & Transformer Studio</h2>
              <p className="text-xs text-slate-400">Select target table and column to tune strategy parameters.</p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            {/* Table picker */}
            <div className="flex items-center gap-2">
              <label className="text-xs text-slate-400 font-medium">Table:</label>
              <select
                value={selectedTable}
                onChange={(e) => {
                  onSelectTable(e.target.value);
                  const newT = tables.find((t) => t.name === e.target.value);
                  if (newT && newT.columns.length > 0) {
                    setSelectedColumn(newT.columns[0].name);
                  }
                }}
                className="bg-[#121929] border border-[#2b3a56] text-white font-mono text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-teal-500"
              >
                {tables.map((t) => (
                  <option key={t.name} value={t.name}>
                    {t.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Column picker */}
            <div className="flex items-center gap-2">
              <label className="text-xs text-slate-400 font-medium">Column:</label>
              <select
                value={selectedColumn}
                onChange={(e) => setSelectedColumn(e.target.value)}
                className="bg-[#121929] border border-[#2b3a56] text-emerald-300 font-mono text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-emerald-500"
              >
                {currentTable?.columns.map((c) => (
                  <option key={c.name} value={c.name}>
                    {c.name} ({c.risk})
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Transformer Selection Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {transformers.map((t) => {
            const isSelected = activeRule.startsWith(t.name);
            return (
              <button
                key={t.name}
                onClick={() => onUpdateRule(currentTable.name, selectedColumn, t.name)}
                className={`p-3.5 rounded-xl border text-left transition flex flex-col justify-between cursor-pointer ${
                  isSelected
                    ? 'bg-[#152338] border-teal-400/60 shadow-lg shadow-teal-500/10'
                    : 'bg-[#121929]/50 border-[#1e293b] hover:bg-[#151f33] text-slate-300'
                }`}
              >
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-mono text-xs font-bold text-teal-300">{t.name}</span>
                    <span className="text-[10px] uppercase font-semibold text-slate-400 bg-[#090d16] px-1.5 py-0.5 rounded border border-[#1e293b]">
                      {t.category}
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 leading-relaxed mb-2">{t.description}</p>
                </div>
                <div className="text-[11px] font-mono text-slate-500 truncate bg-[#090d16]/70 px-2 py-1 rounded border border-[#1e293b]">
                  {t.example}
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Interactive Live Value Tester */}
      <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl">
        <h3 className="text-sm font-bold text-white flex items-center gap-2 mb-4">
          <Code className="w-4 h-4 text-emerald-400" />
          Live Value Transformation Sandbox
        </h3>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Input */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-400">Raw Sensitive Value (Test Input):</label>
            <div className="flex gap-2">
              <input
                type="text"
                value={testInput}
                onChange={(e) => setTestInput(e.target.value)}
                className="w-full bg-[#121929] border border-[#2b3a56] text-white font-mono text-xs rounded-xl px-3.5 py-2.5 focus:outline-none focus:border-emerald-500"
                placeholder="e.g. secret.email@acme.com"
              />
              <button
                onClick={handleTestPreview}
                disabled={isTesting}
                className="bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold px-4 rounded-xl text-xs flex items-center gap-1.5 transition cursor-pointer disabled:opacity-50"
              >
                <Play className="w-3.5 h-3.5 fill-current" />
                <span>Mask</span>
              </button>
            </div>
          </div>

          {/* Output */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-400">Deterministic Anonymized Output:</label>
            <div className="bg-[#090d16] border border-[#1e293b] rounded-xl px-3.5 py-2.5 font-mono text-xs text-emerald-400 flex items-center justify-between min-h-[42px]">
              <span>{testOutput || 'Click Mask to preview transformed value'}</span>
              {testOutput && <CheckCircle className="w-4 h-4 text-emerald-400" />}
            </div>
          </div>
        </div>

        {selectedMeta && (
          <div className="mt-4 p-3 bg-[#121929]/50 border border-[#1e293b] rounded-xl text-xs text-slate-400 flex items-start gap-2">
            <HelpCircle className="w-4 h-4 text-teal-400 shrink-0 mt-0.5" />
            <div>
              <span className="font-semibold text-slate-300">Selected Rule: </span>
              <code className="text-teal-300 font-mono">{activeRule}</code>
              <p className="mt-0.5 text-slate-400">{selectedMeta.description}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
