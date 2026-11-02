import React, { useState } from 'react';
import { ShieldCheck, ShieldAlert, CheckCircle, Search, FileText } from 'lucide-react';
import { RiskLevel } from '../types';

interface LeakViolation {
  line_number: number;
  violation_type: string;
  severity: RiskLevel;
  snippet: string;
  description: string;
}

interface VerificationReport {
  total_lines_scanned: number;
  violations: LeakViolation[];
  is_clean: boolean;
  leak_score: number;
}

interface LeakVerifierProps {
  initialSQL: string;
}

export const LeakVerifier: React.FC<LeakVerifierProps> = ({ initialSQL }) => {
  const [sqlContent, setSqlContent] = useState<string>(initialSQL);
  const [allowTestCards, setAllowTestCards] = useState<boolean>(true);
  const [report, setReport] = useState<VerificationReport | null>(null);
  const [isAuditing, setIsAuditing] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleRunAudit = async () => {
    setIsAuditing(true);
    setError(null);
    try {
      const res = await fetch('/api/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sql: sqlContent,
          allow_synthetic_cards: allowTestCards,
        }),
      });
      if (!res.ok) {
        throw new Error(`Verification failed: ${res.statusText}`);
      }
      const data = await res.json();
      setReport(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Audit error');
    } finally {
      setIsAuditing(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-teal-500/10 border border-teal-500/20 text-teal-400 flex items-center justify-center">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-base font-bold text-white">Post-Masking Zero-Leak Verifier</h2>
            <p className="text-xs text-slate-400">
              Adversarial scanner auditing masked dumps to prove zero lingering PII leaks for SOC 2 and GDPR auditors.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-slate-300 bg-[#090d16] px-3 py-1.5 rounded-lg border border-[#1e293b] cursor-pointer">
            <input
              type="checkbox"
              checked={allowTestCards}
              onChange={(e) => setAllowTestCards(e.target.checked)}
              className="rounded accent-emerald-500"
            />
            <span>Permit Synthetic Test Cards</span>
          </label>
          <button
            onClick={handleRunAudit}
            disabled={isAuditing || !sqlContent.trim()}
            className="bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 font-bold px-4 py-2 rounded-xl text-xs flex items-center gap-1.5 transition cursor-pointer shadow-lg shadow-emerald-500/10 disabled:opacity-50"
          >
            <Search className={`w-3.5 h-3.5 ${isAuditing ? 'animate-spin' : ''}`} />
            <span>{isAuditing ? 'Auditing Dump...' : 'Run Security Audit'}</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-rose-500/10 border border-rose-500/30 rounded-xl p-4 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Editor & Report View */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* SQL Content to verify */}
        <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl overflow-hidden shadow-xl flex flex-col">
          <div className="p-3.5 bg-[#121c2e]/70 border-b border-[#1e293b] flex items-center justify-between text-xs">
            <span className="font-semibold text-slate-300 flex items-center gap-1.5">
              <FileText className="w-3.5 h-3.5 text-teal-400" />
              SQL Dump to Audit
            </span>
            <span className="text-slate-500 font-mono text-[11px]">{sqlContent.split('\n').length} lines</span>
          </div>

          <textarea
            value={sqlContent}
            onChange={(e) => setSqlContent(e.target.value)}
            className="w-full flex-1 min-h-[360px] p-4 bg-[#090d16] text-slate-200 font-mono text-xs focus:outline-none resize-none leading-relaxed"
            placeholder="Paste masked or unmasked SQL dump to verify for PII leaks..."
          />
        </div>

        {/* Audit Results */}
        <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between border-b border-[#1e293b] pb-4 mb-4">
              <h3 className="text-sm font-bold text-white">Compliance Audit Verdict</h3>
              {report && (
                <span className="text-xs font-mono text-slate-400">
                  {report.total_lines_scanned} lines scanned
                </span>
              )}
            </div>

            {report ? (
              <div className="space-y-4">
                {/* Score Pill */}
                <div className={`p-4 rounded-xl border flex items-center justify-between ${
                  report.is_clean
                    ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'
                    : 'bg-rose-500/10 border-rose-500/30 text-rose-300'
                }`}>
                  <div className="flex items-center gap-3">
                    {report.is_clean ? (
                      <CheckCircle className="w-8 h-8 text-emerald-400" />
                    ) : (
                      <ShieldAlert className="w-8 h-8 text-rose-400" />
                    )}
                    <div>
                      <div className="font-bold text-sm">
                        {report.is_clean ? 'Zero Sensitive PII Leaks Detected' : 'Potential Sensitive Leaks Detected'}
                      </div>
                      <div className="text-xs text-slate-400">
                        {report.is_clean
                          ? 'This dump is certified safe for staging and local development.'
                          : `${report.violations.length} lingering sensitive values detected.`}
                      </div>
                    </div>
                  </div>
                  <div className="text-right font-mono">
                    <div className="text-2xl font-extrabold">{report.leak_score} / 100</div>
                    <div className="text-[10px] uppercase tracking-wider text-slate-400">Score</div>
                  </div>
                </div>

                {/* Violations List */}
                {report.violations.length > 0 && (
                  <div className="space-y-2 max-h-[260px] overflow-y-auto pr-1">
                    {report.violations.map((v, idx) => (
                      <div key={idx} className="bg-[#090d16] border border-rose-500/20 rounded-xl p-3 text-xs space-y-1 font-mono">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-rose-400 flex items-center gap-1">
                            Line #{v.line_number} &bull; {v.violation_type}
                          </span>
                          <span className="text-[10px] bg-rose-500/20 text-rose-300 px-1.5 py-0.2 rounded">
                            {v.severity}
                          </span>
                        </div>
                        <div className="text-slate-400 text-[11px]">{v.description}</div>
                        <div className="text-emerald-300 bg-[#121929] px-2 py-1 rounded truncate">
                          {v.snippet}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <div className="p-8 text-center text-slate-500 text-xs">
                <Search className="w-10 h-10 mx-auto mb-2 text-slate-600 opacity-60" />
                <span>Click "Run Security Audit" to verify the SQL dump against all leak detection heuristics.</span>
              </div>
            )}
          </div>

          <div className="pt-4 border-t border-[#1e293b] text-xs text-slate-500 flex items-center justify-between">
            <span>Heuristics: RFC 5322 Emails, Luhn Cards, SSN, API Tokens</span>
            <span className="text-emerald-400 font-medium">SOC 2 / GDPR Verification</span>
          </div>
        </div>
      </div>
    </div>
  );
};
