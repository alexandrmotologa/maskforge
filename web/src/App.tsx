import { useState } from 'react';
import { Table as TableIcon, Sliders, Columns, GitMerge, Terminal, Shield, Lock, Cpu, Sparkles } from 'lucide-react';
import { useMaskSession } from './hooks/useMaskSession';
import { Header } from './components/Header';
import { SchemaTableView } from './components/SchemaTableView';
import { MaskRuleEditor } from './components/MaskRuleEditor';
import { RowDiffPreview } from './components/RowDiffPreview';
import { ReferentialTest } from './components/ReferentialTest';
import { LiveSimulator } from './components/LiveSimulator';
import { ConfigExportModal } from './components/ConfigExportModal';

export function App() {
  const {
    report,
    config,
    transformers,
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
    sqlInput,
  } = useMaskSession();

  const [isExportOpen, setIsExportOpen] = useState(false);

  // Transform config into rule map for easy lookup
  const columnRules: Record<string, Record<string, string>> = {};
  Object.entries(config.tables).forEach(([tbl, tcfg]) => {
    columnRules[tbl] = tcfg.columns || {};
  });

  interface TabItem {
    id: 'tables' | 'editor' | 'diff' | 'referential' | 'simulator';
    label: string;
    icon: typeof TableIcon;
    count?: number;
    badge?: string;
  }

  const tabs: TabItem[] = [
    { id: 'tables', label: 'Schema & PII Matrix', icon: TableIcon, count: report?.total_tables },
    { id: 'editor', label: 'Rule Configurator', icon: Sliders },
    { id: 'diff', label: 'Before / After Diff', icon: Columns },
    { id: 'referential', label: 'Referential Integrity', icon: GitMerge, badge: 'FK Match' },
    { id: 'simulator', label: 'Stream Simulator', icon: Terminal, badge: 'Live Pipe' },
  ];

  return (
    <div className="min-h-screen bg-[#090d16] text-[#e2e8f0] flex flex-col selection:bg-emerald-500/20 selection:text-emerald-300">
      {/* Top Header */}
      <Header
        report={report}
        salt={config.salt}
        onUpdateSalt={updateSalt}
        onRescan={() => runScan()}
        onExport={() => setIsExportOpen(true)}
        isLoading={isLoading}
      />

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">
        {/* Navigation Tabs Bar */}
        <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#1e293b] pb-2">
          <div className="flex flex-wrap items-center gap-2">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold transition cursor-pointer ${
                    isActive
                      ? 'bg-[#172033] text-emerald-300 border border-emerald-500/40 shadow-md shadow-emerald-500/5'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-[#121929] border border-transparent'
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isActive ? 'text-emerald-400' : 'text-slate-500'}`} />
                  <span>{tab.label}</span>
                  {tab.count !== undefined && (
                    <span className="text-[10px] bg-[#090d16] text-slate-400 px-1.5 py-0.5 rounded-full border border-[#1e293b]">
                      {tab.count}
                    </span>
                  )}
                  {tab.badge && (
                    <span className="text-[9px] uppercase tracking-wider font-bold bg-teal-500/20 text-teal-300 border border-teal-500/30 px-1.5 py-0.2 rounded-full">
                      {tab.badge}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          <div className="hidden sm:flex items-center gap-4 text-xs text-slate-500">
            <span className="flex items-center gap-1.5">
              <Lock className="w-3.5 h-3.5 text-emerald-400" /> Zero Disk Writes
            </span>
            <span className="flex items-center gap-1.5">
              <Cpu className="w-3.5 h-3.5 text-teal-400" /> &gt;100k rows/s
            </span>
          </div>
        </div>

        {/* Global Error Banner */}
        {error && (
          <div className="bg-rose-500/10 border border-rose-500/30 rounded-xl p-4 text-rose-300 text-xs flex items-center justify-between">
            <span>{error}</span>
            <button onClick={() => runScan()} className="text-white underline font-semibold cursor-pointer">
              Retry
            </button>
          </div>
        )}

        {/* Active Tab View */}
        <div className="transition-all duration-150">
          {activeTab === 'tables' && (
            <SchemaTableView
              tables={report?.tables || []}
              selectedTable={selectedTable}
              onSelectTable={setSelectedTable}
              columnRules={columnRules}
              onUpdateRule={updateColumnRule}
              transformers={transformers}
            />
          )}

          {activeTab === 'editor' && (
            <MaskRuleEditor
              tables={report?.tables || []}
              selectedTable={selectedTable}
              onSelectTable={setSelectedTable}
              columnRules={columnRules}
              onUpdateRule={updateColumnRule}
              transformers={transformers}
              salt={config.salt}
            />
          )}

          {activeTab === 'diff' && (
            <RowDiffPreview
              tables={report?.tables || []}
              selectedTable={selectedTable}
              onSelectTable={setSelectedTable}
              columnRules={columnRules}
            />
          )}

          {activeTab === 'referential' && (
            <ReferentialTest
              tables={report?.tables || []}
              salt={config.salt}
              onTestReferential={testReferentialMatch}
            />
          )}

          {activeTab === 'simulator' && (
            <LiveSimulator
              initialSQL={sqlInput}
              configYAML={`version: "1.0"\nsalt: "${config.salt}"\n`}
              onSimulate={simulateStream}
            />
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-[#1e293b] py-6 px-6 bg-[#0a0f1d] text-xs text-slate-500 mt-12">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Shield className="w-4 h-4 text-emerald-400" />
            <span className="font-semibold text-slate-300">MaskForge</span>
            <span>&bull; Production Database Anonymizer for PostgreSQL & MySQL</span>
          </div>
          <div className="flex items-center gap-4 text-slate-400">
            <span>Compliance: GDPR Art. 4 & 87, HIPAA § 164.514, PCI-DSS 3.4, SOC 2</span>
            <span>&bull;</span>
            <span className="flex items-center gap-1 text-emerald-400">
              <Sparkles className="w-3.5 h-3.5" /> High-speed streaming engine
            </span>
          </div>
        </div>
      </footer>

      {/* Export YAML Modal */}
      <ConfigExportModal
        config={config}
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
      />
    </div>
  );
}
