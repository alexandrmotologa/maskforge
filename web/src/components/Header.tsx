import React from 'react';
import { Shield, Database, KeyRound, Download, RefreshCw, AlertTriangle } from 'lucide-react';
import { ScanReport } from '../types';

interface HeaderProps {
  report: ScanReport | null;
  salt: string;
  onUpdateSalt: (newSalt: string) => void;
  onRescan: () => void;
  onExport: () => void;
  onDownloadReport?: () => void;
  isLoading: boolean;
}

export const Header: React.FC<HeaderProps> = ({
  report,
  salt,
  onUpdateSalt,
  onRescan,
  onExport,
  onDownloadReport,
  isLoading,
}) => {
  const getRiskColor = (risk?: string) => {
    switch (risk) {
      case 'CRITICAL':
        return 'bg-rose-500/15 text-rose-400 border-rose-500/30';
      case 'HIGH':
        return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
      case 'MEDIUM':
        return 'bg-sky-500/15 text-sky-400 border-sky-500/30';
      default:
        return 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30';
    }
  };

  return (
    <header className="border-b border-[#1e293b] bg-[#0c1220]/90 backdrop-blur-md sticky top-0 z-40 px-6 py-4">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-emerald-500 to-teal-400 flex items-center justify-center shadow-lg shadow-emerald-500/20 text-[#090d16]">
            <Shield className="w-5 h-5 stroke-[2.5]" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
                MaskForge <span className="text-xs uppercase tracking-widest px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 font-semibold">Studio</span>
              </h1>
            </div>
            <p className="text-xs text-slate-400">
              Streaming database anonymizer & deterministic foreign-key preserver
            </p>
          </div>
        </div>

        {/* Global Controls & Metrics */}
        <div className="flex flex-wrap items-center gap-3 w-full md:w-auto justify-end">
          {report && (
            <div className="hidden lg:flex items-center gap-2 bg-[#121929] border border-[#1e293b] rounded-lg px-3 py-1.5 text-xs text-slate-300">
              <Database className="w-3.5 h-3.5 text-emerald-400" />
              <span>{report.total_tables} tables</span>
              <span className="text-slate-600">|</span>
              <span className="text-amber-400 font-medium">{report.pii_columns} PII columns</span>
              <span className="text-slate-600">|</span>
              <span className={`px-2 py-0.5 rounded border text-[11px] font-semibold flex items-center gap-1 ${getRiskColor(report.estimated_risk)}`}>
                <AlertTriangle className="w-3 h-3" />
                {report.estimated_risk}
              </span>
            </div>
          )}

          {/* Salt Pill */}
          <div className="flex items-center gap-2 bg-[#121929] border border-[#1e293b] rounded-lg px-3 py-1.5 text-xs text-slate-300 group hover:border-slate-700 transition">
            <KeyRound className="w-3.5 h-3.5 text-teal-400" />
            <span className="text-slate-400 font-mono text-[11px]">salt:</span>
            <input
              type="text"
              value={salt}
              onChange={(e) => onUpdateSalt(e.target.value)}
              className="bg-transparent text-emerald-300 font-mono text-xs focus:outline-none w-28 text-ellipsis"
              title="Deterministic HMAC salt"
            />
          </div>

          {/* Rescan Button */}
          <button
            onClick={onRescan}
            disabled={isLoading}
            className="flex items-center gap-1.5 bg-[#172033] hover:bg-[#1e2b44] text-slate-200 border border-[#2b3a56] rounded-lg px-3.5 py-1.5 text-xs font-medium transition cursor-pointer disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-emerald-400' : ''}`} />
            <span>Audit PII</span>
          </button>

          {/* Export Report Button */}
          {onDownloadReport && (
            <button
              onClick={onDownloadReport}
              className="flex items-center gap-1.5 bg-[#172033] hover:bg-[#1e2b44] text-teal-300 border border-teal-500/30 rounded-lg px-3 py-1.5 text-xs font-medium transition cursor-pointer"
            >
              <Download className="w-3.5 h-3.5" />
              <span>Audit Report</span>
            </button>
          )}

          {/* Export YAML Button */}
          <button
            onClick={onExport}
            className="flex items-center gap-1.5 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 font-semibold rounded-lg px-3.5 py-1.5 text-xs shadow-md shadow-emerald-500/20 transition cursor-pointer"
          >
            <Download className="w-3.5 h-3.5 stroke-[2.5]" />
            <span>Export mask.yaml</span>
          </button>
        </div>
      </div>
    </header>
  );
};
