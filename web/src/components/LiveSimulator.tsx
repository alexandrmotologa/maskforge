import React, { useState } from 'react';
import { Terminal, Play, Zap, CheckCircle2, RotateCcw, Copy, Check } from 'lucide-react';
import { PipelineMetrics } from '../types';
import { DEMO_SQL_DUMP } from '../data/demoSQL';

const MYSQL_SAMPLE = `-- MySQL Extended INSERT Dump Sample
CREATE TABLE \`users\` (
  \`id\` int(11) NOT NULL,
  \`email\` varchar(255) NOT NULL,
  \`full_name\` varchar(100) DEFAULT NULL,
  \`phone\` varchar(50) DEFAULT NULL,
  PRIMARY KEY (\`id\`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO \`users\` (\`id\`, \`email\`, \`full_name\`, \`phone\`) VALUES
(1, 'ceo@techstartup.io', 'Marc Andreessen', '+14155550199'),
(2, 'cto@cloudbank.com', 'Ada Lovelace', '+447700900077'),
(3, 'security@finshield.de', 'Carl Friedrich Gauss', '+49301234567');
`;

interface LiveSimulatorProps {
  initialSQL: string;
  configYAML: string;
  onSimulate: (sql: string, configYAML: string) => Promise<{ maskedSQL: string; metrics: PipelineMetrics }>;
}

export const LiveSimulator: React.FC<LiveSimulatorProps> = ({
  initialSQL,
  configYAML,
  onSimulate,
}) => {
  const [sqlInput, setSqlInput] = useState<string>(initialSQL || DEMO_SQL_DUMP);
  const [sqlOutput, setSqlOutput] = useState<string>('');
  const [metrics, setMetrics] = useState<PipelineMetrics | null>(null);
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [copied, setCopied] = useState<boolean>(false);

  const handleRunStream = async () => {
    setIsRunning(true);
    try {
      const res = await onSimulate(sqlInput, configYAML);
      setSqlOutput(res.maskedSQL);
      setMetrics(res.metrics);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Simulation failed';
      setSqlOutput(`-- ERROR: ${msg}`);
    } finally {
      setIsRunning(false);
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(sqlOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Top Banner & Presets */}
      <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl p-6 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center">
            <Terminal className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-base font-bold text-white">Live SQL Streaming Pipe Simulator</h2>
            <p className="text-xs text-slate-400">
              Simulates zero-disk pipe execution in-memory (<code className="text-emerald-300">pg_dump | maskforge | psql</code>).
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setSqlInput(DEMO_SQL_DUMP)}
            className="text-xs bg-[#121929] hover:bg-[#162136] text-slate-300 border border-[#1e293b] px-3 py-1.5 rounded-lg transition cursor-pointer"
          >
            PostgreSQL COPY Preset
          </button>
          <button
            onClick={() => setSqlInput(MYSQL_SAMPLE)}
            className="text-xs bg-[#121929] hover:bg-[#162136] text-slate-300 border border-[#1e293b] px-3 py-1.5 rounded-lg transition cursor-pointer"
          >
            MySQL INSERT Preset
          </button>
        </div>
      </div>

      {/* Editor Split View */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input SQL */}
        <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl overflow-hidden shadow-xl flex flex-col">
          <div className="p-3.5 bg-[#121c2e]/70 border-b border-[#1e293b] flex items-center justify-between text-xs">
            <span className="font-semibold text-slate-300 flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-rose-400"></span>
              Raw Input Stream (STDIN)
            </span>
            <span className="text-slate-500 font-mono text-[11px]">{sqlInput.split('\n').length} lines</span>
          </div>

          <textarea
            value={sqlInput}
            onChange={(e) => setSqlInput(e.target.value)}
            className="w-full flex-1 min-h-[380px] p-4 bg-[#090d16] text-slate-200 font-mono text-xs focus:outline-none resize-none leading-relaxed selection:bg-rose-500/20"
            spellCheck={false}
            placeholder="Paste pg_dump or mysqldump stream here..."
          />

          <div className="p-3 bg-[#0c1220] border-t border-[#1e293b] flex items-center justify-between">
            <button
              onClick={() => setSqlInput('')}
              className="text-xs text-slate-500 hover:text-slate-300 flex items-center gap-1 transition cursor-pointer"
            >
              <RotateCcw className="w-3 h-3" /> Clear
            </button>
            <button
              onClick={handleRunStream}
              disabled={isRunning || !sqlInput.trim()}
              className="bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 font-bold px-5 py-2 rounded-xl text-xs flex items-center gap-1.5 transition cursor-pointer shadow-lg shadow-emerald-500/10 disabled:opacity-50"
            >
              <Play className={`w-3.5 h-3.5 fill-current ${isRunning ? 'animate-spin' : ''}`} />
              <span>{isRunning ? 'Streaming...' : 'Run Streaming Pipe'}</span>
            </button>
          </div>
        </div>

        {/* Output SQL */}
        <div className="bg-[#0f172a]/80 border border-[#1e293b] rounded-2xl overflow-hidden shadow-xl flex flex-col">
          <div className="p-3.5 bg-[#121c2e]/70 border-b border-[#1e293b] flex items-center justify-between text-xs">
            <span className="font-semibold text-slate-300 flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
              Masked Output Stream (STDOUT)
            </span>
            {sqlOutput && (
              <button
                onClick={copyToClipboard}
                className="text-slate-400 hover:text-white flex items-center gap-1 text-[11px] transition cursor-pointer"
              >
                {copied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                {copied ? 'Copied' : 'Copy'}
              </button>
            )}
          </div>

          <textarea
            value={sqlOutput}
            readOnly
            className="w-full flex-1 min-h-[380px] p-4 bg-[#090d16] text-emerald-300 font-mono text-xs focus:outline-none resize-none leading-relaxed selection:bg-emerald-500/20"
            placeholder="Masked SQL output will stream here..."
          />

          <div className="p-3 bg-[#0c1220] border-t border-[#1e293b] flex items-center justify-between text-xs text-slate-400">
            <span className="flex items-center gap-1 text-emerald-400">
              <CheckCircle2 className="w-3.5 h-3.5" />
              Safe for staging & local test import
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              {sqlOutput ? `${sqlOutput.split('\n').length} lines` : 'Awaiting stream run'}
            </span>
          </div>
        </div>
      </div>

      {/* Metrics HUD */}
      {metrics && (
        <div className="bg-[#101726] border border-emerald-500/30 rounded-2xl p-5 shadow-xl grid grid-cols-2 md:grid-cols-5 gap-4 text-center font-mono">
          <div className="p-3 bg-[#090d16] rounded-xl border border-[#1e293b]">
            <div className="text-[11px] text-slate-400 uppercase tracking-wider mb-1">Rows Masked</div>
            <div className="text-lg font-bold text-white">{metrics.RowsMasked}</div>
          </div>

          <div className="p-3 bg-[#090d16] rounded-xl border border-[#1e293b]">
            <div className="text-[11px] text-slate-400 uppercase tracking-wider mb-1">Lines Processed</div>
            <div className="text-lg font-bold text-teal-300">{metrics.LinesProcessed}</div>
          </div>

          <div className="p-3 bg-[#090d16] rounded-xl border border-[#1e293b]">
            <div className="text-[11px] text-slate-400 uppercase tracking-wider mb-1">Duration</div>
            <div className="text-lg font-bold text-sky-400">{(metrics.Duration / 1000000).toFixed(1)} ms</div>
          </div>

          <div className="p-3 bg-[#090d16] rounded-xl border border-[#1e293b]">
            <div className="text-[11px] text-slate-400 uppercase tracking-wider mb-1">Throughput</div>
            <div className="text-lg font-bold text-emerald-400 flex items-center justify-center gap-1">
              <Zap className="w-4 h-4 fill-current text-amber-400" />
              {metrics.RowsPerSec.toFixed(0)} <span className="text-xs font-normal text-slate-400">r/s</span>
            </div>
          </div>

          <div className="p-3 bg-[#090d16] rounded-xl border border-[#1e293b] col-span-2 md:col-span-1">
            <div className="text-[11px] text-slate-400 uppercase tracking-wider mb-1">Stream Bandwidth</div>
            <div className="text-lg font-bold text-purple-400">{metrics.MBPerSec.toFixed(2)} MB/s</div>
          </div>
        </div>
      )}
    </div>
  );
};
