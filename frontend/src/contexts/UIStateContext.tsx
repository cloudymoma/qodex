import {
  createContext,
  useContext,
  useState,
  useCallback,
  useMemo,
  type ReactNode,
} from 'react';
import type { BottomPanelMode, TreeNode } from '@/types';

interface UIStateContextValue {
  bottomPanelMode: BottomPanelMode;
  setBottomPanelMode: (mode: BottomPanelMode) => void;
  toggleBottomPanel: () => void;
  sidebarCollapsed: boolean;
  setSidebarCollapsed: (collapsed: boolean) => void;
  treeData: TreeNode[];
  setTreeData: (data: TreeNode[]) => void;
  loading: boolean;
  setLoading: (loading: boolean) => void;
  repoName: string;
  setRepoName: (name: string) => void;
  codeViewPath: string | null;
  setCodeViewPath: (path: string | null) => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  timelineGlowFiles: ReadonlySet<string>;
  setTimelineGlowFiles: (files: ReadonlySet<string>) => void;
  timelineGraph: import('@/types').GraphData | null;
  setTimelineGraph: (data: import('@/types').GraphData | null) => void;
}

const UIStateContext = createContext<UIStateContextValue | null>(null);

export function UIStateProvider({ children }: { children: ReactNode }) {
  const [bottomPanelMode, setBottomPanelMode] = useState<BottomPanelMode>('hidden');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [repoName, setRepoName] = useState('');
  const [codeViewPath, setCodeViewPath] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [timelineGlowFiles, setTimelineGlowFiles] = useState<ReadonlySet<string>>(new Set());
  const [timelineGraph, setTimelineGraph] = useState<import('@/types').GraphData | null>(null);

  const toggleBottomPanel = useCallback(() => {
    setBottomPanelMode((prev) => (prev === 'hidden' ? 'search' : 'hidden'));
  }, []);

  const value = useMemo(
    () => ({
      bottomPanelMode,
      setBottomPanelMode,
      toggleBottomPanel,
      sidebarCollapsed,
      setSidebarCollapsed,
      treeData,
      setTreeData,
      loading,
      setLoading,
      repoName,
      setRepoName,
      codeViewPath,
      setCodeViewPath,
      searchQuery,
      setSearchQuery,
      timelineGlowFiles,
      setTimelineGlowFiles,
      timelineGraph,
      setTimelineGraph,
    }),
    [bottomPanelMode, toggleBottomPanel, sidebarCollapsed, treeData, loading, repoName, codeViewPath, searchQuery, timelineGlowFiles, timelineGraph],
  );

  return (
    <UIStateContext.Provider value={value}>{children}</UIStateContext.Provider>
  );
}

export function useUIState() {
  const ctx = useContext(UIStateContext);
  if (!ctx) throw new Error('useUIState must be used within UIStateProvider');
  return ctx;
}
