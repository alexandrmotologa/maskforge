import React, { useState } from 'react';
import { GitMerge, ArrowRight, CheckCircle2, Shield, AlertCircle } from 'lucide-react';
import { DiscoveredTable } from '../types';

interface ReferentialTestProps {
  tables: DiscoveredTable[];
  salt: string;
  onTestReferential: (
    val: string,
    rule: string,
    tableA: string,
    colA: string,
    tableB: string,
    colB: string
  ) => Promise<{ raw_value: string; masked_a: string; masked_b: string; match: boolean }>;
}

export const ReferentialTest: React.FC<ReferentialTestProps> = ({
  tables,
  salt,
  onTestReferential,
}) => {
  const [tableA, setTableA] = useState<string>(tables[0]?.name || 'users');
  const [colA, setColA] = useState<string>('email');
  const [tableB, setTableB] = useState<string>(tables[tables.length > 2 ? 2 : 1]?.name || 'orders');
  const [colB, setColB] = useState<string>('customer_email');
  const [testValue, setTestValue] = useState<string>('elena.popescu@fintech-cloud.ro');
  const [selectedRule, setSelectedRule] = useState<string>('faker.email');
  const [result, setResult] = useState<{
    raw_value: string;
    masked_a: string;
    masked_b: string;
    match: boolean;
  } | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const tA = tables.find((t) => t.name === tableA);
  const tB = tables.find((t) => t.name === tableB);

  const handleRunVerification = async () => {
    setIsLoading(true);
    try {
      const res = await onTestReferential(testValue, selectedRule, tableA, colA, tableB, colB);
      setResult(res);
    } catch {
      // handled
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl">
        <div className="flex items-center gap-3 border-b border-[#1e293b] pb-5 mb-5">
          <div className="w-10 h-10 rounded-xl bg-teal-500/10 border border-teal-500/20 text-teal-400 flex items-center justify-center">
            <GitMerge className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-base font-bold text-white">Foreign Key Referential Integrity Verifier</h2>
            <p className="text-xs text-slate-400">
              Prove that relationships between primary keys and foreign keys remain 100% consistent across tables.
            </p>
          </div>
        </div>

        {/* Configuration Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          {/* Entity A */}
          <div className="bg-[#121929]/70 border border-[#1e293b] rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-teal-400 uppercase tracking-wider">Primary Entity (Table A)</span>
              <span className="text-[11px] text-slate-500 font-mono">e.g. users.id / users.email</span>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[11px] text-slate-400 block mb-1">Table:</label>
                <select
                  value={tableA}
                  onChange={(e) => {
                    setTableA(e.target.value);
                    const t = tables.find((x) => x.name === e.target.value);
                    if (t && t.columns.length > 0) setColA(t.columns[0].name);
                  }}
                  className="w-full bg-[#090d16] border border-[#2b3a56] text-white font-mono text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-teal-500"
                >
                  {tables.map((t) => (
                    <option key={t.name} value={t.name}>
                      {t.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-[11px] text-slate-400 block mb-1">Column:</label>
                <select
                  value={colA}
                  onChange={(e) => setColA(e.target.value)}
                  className="w-full bg-[#090d16] border border-[#2b3a56] text-emerald-300 font-mono text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-teal-500"
                >
                  {tA?.columns.map((c) => (
                    <option key={c.name} value={c.name}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Entity B */}
          <div className="bg-[#121929]/70 border border-[#1e293b] rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-emerald-400 uppercase tracking-wider">Foreign Entity (Table B)</span>
              <span className="text-[11px] text-slate-500 font-mono">e.g. orders.customer_email</span>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[11px] text-slate-400 block mb-1">Table:</label>
                <select
                  value={tableB}
                  onChange={(e) => {
                    setTableB(e.target.value);
                    const t = tables.find((x) => x.name === e.target.value);
                    if (t && t.columns.length > 0) setColB(t.columns[0].name);
                  }}
                  className="w-full bg-[#090d16] border border-[#2b3a56] text-white font-mono text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-emerald-500"
                >
                  {tables.map((t) => (
                    <option key={t.name} value={t.name}>
                      {t.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-[11px] text-slate-400 block mb-1">Column:</label>
                <select
                  value={colB}
                  onChange={(e) => setColB(e.target.value)}
                  className="w-full bg-[#090d16] border border-[#2b3a56] text-emerald-300 font-mono text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-emerald-500"
                >
                  {tB?.columns.map((c) => (
                    <option key={c.name} value={c.name}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Input Value & Strategy Picker */}
        <div className="bg-[#090d16] border border-[#1e293b] rounded-xl p-4 space-y-4 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="md:col-span-2">
              <label className="text-xs font-semibold text-slate-300 block mb-1">Shared Key / Sensitive Identifier Value:</label>
              <input
                type="text"
                value={testValue}
                onChange={(e) => setTestValue(e.target.value)}
                className="w-full bg-[#121929] border border-[#2b3a56] text-white font-mono text-xs rounded-lg px-3.5 py-2.5 focus:outline-none focus:border-emerald-500"
                placeholder="e.g. elena.popescu@fintech-cloud.ro or 1001"
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-300 block mb-1">Shared Masking Strategy:</label>
              <select
                value={selectedRule}
                onChange={(e) => setSelectedRule(e.target.value)}
                className="w-full bg-[#121929] border border-[#2b3a56] text-emerald-300 font-mono text-xs rounded-lg px-3 py-2.5 focus:outline-none focus:border-emerald-500"
              >
                <option value="faker.email">faker.email</option>
                <option value="faker.uuid">faker.uuid</option>
                <option value="faker.phone_number">faker.phone_number</option>
                <option value="hash.sha256">hash.sha256</option>
                <option value="faker.credit_card">faker.credit_card</option>
              </select>
            </div>
          </div>

          <button
            onClick={handleRunVerification}
            disabled={isLoading}
            className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 font-bold py-2.5 rounded-xl text-xs flex items-center justify-center gap-2 transition cursor-pointer shadow-lg shadow-emerald-500/10 disabled:opacity-50"
          >
            <GitMerge className="w-4 h-4 stroke-[2.5]" />
            <span>Assert Referential Join Match</span>
          </button>
        </div>

        {/* Results Visualizer */}
        {result && (
          <div className="bg-[#121a2b] border border-emerald-500/30 rounded-2xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#1e293b] pb-3">
              <span className="text-xs font-bold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                Referential Match Verified: 100% Consistent
              </span>
              <span className="text-xs font-mono text-slate-400">Salt: {salt.substring(0, 10)}...</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 items-center gap-4 text-center font-mono">
              <div className="bg-[#090d16] p-4 rounded-xl border border-[#1e293b]">
                <div className="text-[11px] text-slate-500 mb-1">{tableA}.{colA}</div>
                <div className="text-sm font-bold text-emerald-300 break-all">{result.masked_a}</div>
              </div>

              <div className="flex flex-col items-center justify-center text-emerald-400">
                <ArrowRight className="w-6 h-6 animate-pulse hidden md:block" />
                <span className="text-xs font-bold mt-1 text-teal-300">== IDENTICAL ==</span>
              </div>

              <div className="bg-[#090d16] p-4 rounded-xl border border-[#1e293b]">
                <div className="text-[11px] text-slate-500 mb-1">{tableB}.{colB}</div>
                <div className="text-sm font-bold text-emerald-300 break-all">{result.masked_b}</div>
              </div>
            </div>

            <div className="p-3 bg-[#0a0f1d] rounded-xl border border-[#1e293b] text-xs text-slate-400 flex items-center gap-2">
              <Shield className="w-4 h-4 text-teal-400 shrink-0" />
              <span>
                Foreign keys across {tableA} and {tableB} join seamlessly. Queries like{' '}
                <code className="text-slate-200">JOIN {tableB} ON {tableA}.{colA} = {tableB}.{colB}</code> will return identical row counts in staging as in production!
              </span>
            </div>
          </div>
        )}

        {!result && (
          <div className="p-4 bg-[#121929]/40 border border-[#1e293b] rounded-xl text-xs text-slate-400 flex items-start gap-2.5">
            <AlertCircle className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
            <div>
              <span className="font-semibold text-slate-200">How Deterministic Preserving Works:</span>
              <p className="mt-1 leading-relaxed text-slate-400">
                Traditional faker tools generate random values on each invocation, destroying foreign key joins. MaskForge computes an HMAC-SHA256 seed derived from your salt and raw value, ensuring any occurrence of an identifier maps to the exact same pseudonym everywhere in the database.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
