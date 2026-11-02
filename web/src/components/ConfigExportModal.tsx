import React, { useState } from 'react';
import { Copy, Check, Download, Terminal, X, FileCode } from 'lucide-react';
import { Config } from '../types';

interface ConfigExportModalProps {
  config: Config;
  isOpen: boolean;
  onClose: () => void;
}

export const ConfigExportModal: React.FC<ConfigExportModalProps> = ({
  config,
  isOpen,
  onClose,
}) => {
  const [copiedYAML, setCopiedYAML] = useState(false);
  const [copiedCmd, setCopiedCmd] = useState(false);

  if (!isOpen) return null;

  // Generate YAML string
  const generateYAML = (): string => {
    let yaml = `# MaskForge Declarative Privacy Masking Configuration\n`;
    yaml += `version: "${config.version || '1.0'}"\n`;
    yaml += `salt: "${config.salt}"\n\n`;
    yaml += `global:\n`;
    yaml += `  preserve_nulls: true\n`;
    if (config.global?.ignore_columns && config.global.ignore_columns.length > 0) {
      yaml += `  ignore_columns: [${config.global.ignore_columns.map((c) => `"${c}"`).join(', ')}]\n`;
    }
    yaml += `\ntables:\n`;

    Object.entries(config.tables).forEach(([tbl, tcfg]) => {
      yaml += `  ${tbl}:\n`;
      if (tcfg.columns && Object.keys(tcfg.columns).length > 0) {
        yaml += `    columns:\n`;
        Object.entries(tcfg.columns).forEach(([col, rule]) => {
          yaml += `      ${col}: "${rule}"\n`;
        });
      }
      if (tcfg.ignore_columns && tcfg.ignore_columns.length > 0) {
        yaml += `    ignore_columns: [${tcfg.ignore_columns.map((c) => `"${c}"`).join(', ')}]\n`;
      }
    });

    return yaml;
  };

  const yamlContent = generateYAML();
  const cliCommand = `pg_dump -Fc prod_db | pg_restore | maskforge run --config mask.yaml | psql staging_db`;

  const handleCopyYAML = () => {
    navigator.clipboard.writeText(yamlContent);
    setCopiedYAML(true);
    setTimeout(() => setCopiedYAML(false), 2000);
  };

  const handleCopyCmd = () => {
    navigator.clipboard.writeText(cliCommand);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
  };

  const handleDownload = () => {
    const blob = new Blob([yamlContent], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'mask.yaml';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
      <div className="bg-[#0f172a] border border-[#1e293b] rounded-2xl max-w-3xl w-full overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
        {/* Modal Header */}
        <div className="p-5 border-b border-[#1e293b] bg-[#121c2e]/80 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center">
              <FileCode className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Export mask.yaml & CLI Command</h3>
              <p className="text-xs text-slate-400">Declarative rules ready for production CI/CD or staging pipes.</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-[#1e293b] transition cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Content */}
        <div className="p-6 space-y-5 overflow-y-auto flex-1">
          {/* Production CLI Command */}
          <div>
            <label className="text-xs font-semibold text-slate-300 block mb-1.5 flex items-center gap-1.5">
              <Terminal className="w-3.5 h-3.5 text-teal-400" />
              Direct Streaming Pipe Command (Zero Disk Exposure):
            </label>
            <div className="bg-[#090d16] border border-[#1e293b] rounded-xl p-3 flex items-center justify-between font-mono text-xs text-emerald-300">
              <span className="truncate">{cliCommand}</span>
              <button
                onClick={handleCopyCmd}
                className="ml-3 px-3 py-1 bg-[#172033] hover:bg-[#1e2b44] text-slate-200 border border-[#2b3a56] rounded-lg text-xs flex items-center gap-1 shrink-0 transition cursor-pointer"
              >
                {copiedCmd ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                {copiedCmd ? 'Copied' : 'Copy'}
              </button>
            </div>
          </div>

          {/* YAML Viewer */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-xs font-semibold text-slate-300">mask.yaml Configuration:</label>
              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopyYAML}
                  className="text-xs text-slate-400 hover:text-white flex items-center gap-1 transition cursor-pointer"
                >
                  {copiedYAML ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                  {copiedYAML ? 'Copied' : 'Copy YAML'}
                </button>
                <span className="text-slate-600">|</span>
                <button
                  onClick={handleDownload}
                  className="text-xs text-emerald-400 hover:text-emerald-300 font-semibold flex items-center gap-1 transition cursor-pointer"
                >
                  <Download className="w-3 h-3" />
                  Download mask.yaml
                </button>
              </div>
            </div>

            <pre className="bg-[#090d16] border border-[#1e293b] rounded-xl p-4 font-mono text-xs text-slate-300 overflow-x-auto max-h-[300px] leading-relaxed selection:bg-emerald-500/20">
              {yamlContent}
            </pre>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="p-4 bg-[#0a0f1d] border-t border-[#1e293b] flex items-center justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-[#172033] hover:bg-[#1e2b44] text-slate-300 rounded-xl text-xs font-medium transition cursor-pointer"
          >
            Close
          </button>
          <button
            onClick={handleDownload}
            className="px-4 py-2 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 font-bold rounded-xl text-xs flex items-center gap-1.5 transition cursor-pointer shadow-lg shadow-emerald-500/10"
          >
            <Download className="w-4 h-4 stroke-[2.5]" />
            <span>Download Configuration</span>
          </button>
        </div>
      </div>
    </div>
  );
};
