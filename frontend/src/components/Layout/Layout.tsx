import { useUIState } from '@/contexts/UIStateContext';
import { RepoInput } from '@/components/RepoInput/RepoInput';
import { LeftSidebar } from '@/components/LeftSidebar/LeftSidebar';
import { MainCanvas } from '@/components/MainCanvas/MainCanvas';
import { BottomPanel } from '@/components/BottomPanel/BottomPanel';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import clsx from 'clsx';
import { track } from '@/services/tracker';

export function Layout() {
  const { bottomPanelMode, setBottomPanelMode, sidebarCollapsed, setSidebarCollapsed } = useUIState();

  return (
    <div className="flex flex-col h-screen bg-dark-bg">
      {/* Top bar: Repo URL input */}
      <RepoInput />

      {/* Main content area */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Left sidebar + toggle wrapper */}
        <div
          className={clsx(
            'shrink-0 border-r border-dark-border bg-dark-bg-secondary transition-[width] duration-300 overflow-hidden relative',
            sidebarCollapsed ? 'w-0 border-r-0' : 'w-72',
          )}
        >
          <div className="w-72 h-full">
            <LeftSidebar />
          </div>
        </div>

        {/* Sidebar toggle — sits right after the sidebar in flow */}
        <button
          onClick={() => { track('toggle_sidebar', 'layout', sidebarCollapsed ? 'expand' : 'collapse'); setSidebarCollapsed(!sidebarCollapsed); }}
          className="absolute top-1/2 -translate-y-1/2 z-40 bg-dark-bg-secondary border border-dark-border rounded-r-md px-0.5 py-2 text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg-tertiary transition-all duration-300"
          style={{ left: sidebarCollapsed ? 0 : 'calc(18rem - 1px)' }}
          title={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {sidebarCollapsed ? <PanelLeftOpen size={14} /> : <PanelLeftClose size={14} />}
        </button>

        {/* 3D Graph canvas */}
        <div className="flex-1 relative overflow-hidden">
          <MainCanvas />
        </div>
      </div>

      {/* Bottom panel */}
      <BottomPanel />

      {/* Floating toggle button when panel is hidden */}
      {bottomPanelMode === 'hidden' && (
        <button
          onClick={() => { track('open_search', 'floating_button'); setBottomPanelMode('search'); }}
          className="fixed bottom-4 right-4 z-40 bg-[#2a2a2a] rounded-full p-2 shadow-lg hover:bg-[#3a3a3a] transition-colors border border-[#444]"
          title="Open Search"
        >
          <img src="/qodex_logo.png" alt="Search" className="w-7 h-7 object-contain" />
        </button>
      )}
    </div>
  );
}
