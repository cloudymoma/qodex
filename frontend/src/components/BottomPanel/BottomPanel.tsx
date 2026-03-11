import { useUIState } from '@/contexts/UIStateContext';
import { SearchPanel } from './SearchPanel';
import { ChatPanel } from './ChatPanel';
import { HistoryPanel } from './HistoryPanel';
import { Search, MessageSquare, History, ChevronDown } from 'lucide-react';
import clsx from 'clsx';
import { track } from '@/services/tracker';

export function BottomPanel() {
  const { bottomPanelMode, setBottomPanelMode } = useUIState();

  if (bottomPanelMode === 'hidden') return null;

  return (
    <div className="border-t border-dark-border bg-dark-bg-secondary" style={{ height: '280px' }}>
      {/* Tab header */}
      <div className="flex items-center justify-between px-4 py-1.5 border-b border-dark-border">
        <div className="flex gap-1">
          <button
            onClick={() => { track('tab_switch', 'bottom_panel', 'search'); setBottomPanelMode('search'); }}
            className={clsx(
              'flex items-center gap-1.5 px-3 py-1 rounded text-sm font-medium transition-colors',
              bottomPanelMode === 'search'
                ? 'bg-accent-primary text-white'
                : 'text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg-tertiary',
            )}
          >
            <Search size={14} />
            Search
          </button>
          <button
            onClick={() => { track('tab_switch', 'bottom_panel', 'history'); setBottomPanelMode('history'); }}
            className={clsx(
              'flex items-center gap-1.5 px-3 py-1 rounded text-sm font-medium transition-colors',
              bottomPanelMode === 'history'
                ? 'bg-accent-primary text-white'
                : 'text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg-tertiary',
            )}
          >
            <History size={14} />
            History
          </button>
          <button
            onClick={() => { track('tab_switch', 'bottom_panel', 'chat'); setBottomPanelMode('chat'); }}
            className={clsx(
              'flex items-center gap-1.5 px-3 py-1 rounded text-sm font-medium transition-colors',
              bottomPanelMode === 'chat'
                ? 'bg-accent-primary text-white'
                : 'text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg-tertiary',
            )}
          >
            <MessageSquare size={14} />
            Chat
          </button>
        </div>

        <button
          onClick={() => { track('tab_switch', 'bottom_panel', 'hidden'); setBottomPanelMode('hidden'); }}
          className="p-1 rounded text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg-tertiary"
        >
          <ChevronDown size={16} />
        </button>
      </div>

      {/* Panel content */}
      <div className="h-[calc(100%-36px)] overflow-hidden">
        {bottomPanelMode === 'search' && <SearchPanel />}
        {bottomPanelMode === 'history' && <HistoryPanel />}
        {bottomPanelMode === 'chat' && <ChatPanel />}
      </div>
    </div>
  );
}
