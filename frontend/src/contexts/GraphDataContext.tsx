import {
  createContext,
  useContext,
  useState,
  useMemo,
  useCallback,
  type ReactNode,
} from 'react';
import type { GraphData, GraphNode } from '@/types';

interface GraphDataContextValue {
  fullGraphData: GraphData;
  displayGraphData: GraphData;
  focusedNodeIds: ReadonlySet<string>;
  setFullGraphData: (data: GraphData) => void;
  focusNode: (nodeId: string) => void;
  resetView: () => void;
}

const GraphDataContext = createContext<GraphDataContextValue | null>(null);

const emptyGraph: GraphData = { nodes: [], links: [] };
const emptySet = new Set<string>();

// Helper to get ID from link source/target which D3-force mutates from string to object
function getLinkId(linkEnd: string | GraphNode): string {
  return typeof linkEnd === 'string' ? linkEnd : linkEnd.id;
}

// Deep clone graph data to prevent D3-force from mutating our state
function cloneGraphData(data: GraphData): GraphData {
  return {
    nodes: data.nodes.map(n => ({ ...n })),
    links: data.links.map(l => ({ ...l })),
  };
}

export function GraphDataProvider({ children }: { children: ReactNode }) {
  const [fullGraphData, setFullGraphData] = useState<GraphData>(emptyGraph);
  const [focusedNodeIds, setFocusedNodeIds] = useState<ReadonlySet<string>>(emptySet);

  const displayGraphData = useMemo(() => {
    if (focusedNodeIds.size === 0) return cloneGraphData(fullGraphData);

    // Visible = all focused nodes + their direct neighbors (one level)
    const visible = new Set<string>(focusedNodeIds);
    fullGraphData.links.forEach((link) => {
      const sourceId = getLinkId(link.source);
      const targetId = getLinkId(link.target);
      if (focusedNodeIds.has(sourceId)) visible.add(targetId);
      if (focusedNodeIds.has(targetId)) visible.add(sourceId);
    });

    return cloneGraphData({
      nodes: fullGraphData.nodes.filter((n) => visible.has(n.id)),
      links: fullGraphData.links.filter((l) => {
        const sourceId = getLinkId(l.source);
        const targetId = getLinkId(l.target);
        return visible.has(sourceId) && visible.has(targetId);
      }),
    });
  }, [fullGraphData, focusedNodeIds]);

  // Toggle: add if not focused, remove if already focused
  const focusNode = useCallback((nodeId: string) => {
    setFocusedNodeIds((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }, []);

  const resetView = useCallback(() => {
    setFocusedNodeIds(emptySet);
  }, []);

  const value = useMemo(
    () => ({
      fullGraphData,
      displayGraphData,
      focusedNodeIds,
      setFullGraphData,
      focusNode,
      resetView,
    }),
    [fullGraphData, displayGraphData, focusedNodeIds, focusNode, resetView],
  );

  return (
    <GraphDataContext.Provider value={value}>
      {children}
    </GraphDataContext.Provider>
  );
}

export function useGraphData() {
  const ctx = useContext(GraphDataContext);
  if (!ctx) throw new Error('useGraphData must be used within GraphDataProvider');
  return ctx;
}
