import { useRef, useCallback, useEffect, useState, useMemo, lazy, Suspense } from 'react';
import * as THREE from 'three';
import SpriteText from 'three-spritetext';
import ForceGraph2D, { type ForceGraphMethods as ForceGraphMethods2D } from 'react-force-graph-2d';
import { useGraphData } from '@/contexts/GraphDataContext';
import { useUIState } from '@/contexts/UIStateContext';
import type { GraphNode } from '@/types';
import { Loader2, X, Box, Square } from 'lucide-react';
import { CodeOverlay } from './CodeOverlay';

// Lazy-load ForceGraph3D so the page doesn't crash when WebGL is unavailable
const ForceGraph3D = lazy(() => import('react-force-graph-3d'));

// Detect WebGL support once at module load
function detectWebGL(): boolean {
  try {
    const canvas = document.createElement('canvas');
    const gl = canvas.getContext('webgl2') || canvas.getContext('webgl');
    return gl !== null;
  } catch {
    return false;
  }
}

const webglSupported = detectWebGL();

// Scale line count to a reasonable node size (1-8 range)
function scaleNodeVal(val: number | undefined): number {
  if (!val || val <= 0) return 1;
  return Math.max(1, Math.min(8, Math.log2(val)));
}

// --- Extension-based coloring (persistent per session, shuffled per load) ---

// Node colors must not conflict with size-indicator center dot colors:
// blue=#0000FF, tiffany=#0ABAB5, green=#00FF00, yellow=#FFD700, red=#FF0000
const COLOR_PALETTE = [
  '#00ADD8', '#DEA584', '#3178C6', '#3776AB',
  '#B07219', '#E34C26', '#563D7C', '#A97BFF',
  '#FF6F61', '#88C0D0', '#EBCB8B', '#D08770',
  '#B48EAD', '#5E81AC', '#81A1C1', '#8FBCBB',
  '#E06C75', '#C678DD', '#D19A66', '#56B6C2',
];

// Fisher-Yates shuffle (runs once per page load → different order each session)
function shuffle(arr: string[]): string[] {
  const a = [...arr];
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    const tmp = a[i]!;
    a[i] = a[j]!;
    a[j] = tmp;
  }
  return a;
}

const shuffledPalette = shuffle(COLOR_PALETTE);
const extColorMap = new Map<string, string>();
const NO_EXT_COLOR = '#666666';
const HIGHLIGHT_COLOR = '#00ff88';

function getExtColor(name: string): string {
  const dot = name.lastIndexOf('.');
  const ext = dot > 0 ? name.substring(dot).toLowerCase() : '';
  if (!ext) return NO_EXT_COLOR;

  const existing = extColorMap.get(ext);
  if (existing) return existing;

  const color = shuffledPalette[extColorMap.size % shuffledPalette.length]!;
  extColorMap.set(ext, color);
  return color;
}

// --- File size categories (by line count) ---

const SIZE_CATEGORIES = [
  { label: '0–50',    name: 'Tiny',   color: '#0000FF' },  // blue
  { label: '51–200',  name: 'Small',  color: '#0ABAB5' },  // tiffany blue
  { label: '201–500', name: 'Medium', color: '#00FF00' },  // green
  { label: '501–1k',  name: 'Large',  color: '#FFD700' },  // yellow
  { label: '>1,000',  name: 'Giant',  color: '#FF0000' },  // red
] as const;

function getSizeColor(lines: number | undefined): string {
  if (!lines || lines <= 50) return SIZE_CATEGORIES[0].color;
  if (lines <= 200) return SIZE_CATEGORIES[1].color;
  if (lines <= 500) return SIZE_CATEGORIES[2].color;
  if (lines <= 1000) return SIZE_CATEGORIES[3].color;
  return SIZE_CATEGORIES[4].color;
}

// --- Component ---

export function MainCanvas() {
  const graphRef3D = useRef<any>(undefined);
  const graphRef2D = useRef<ForceGraphMethods2D>(undefined);
  const containerRef = useRef<HTMLDivElement>(null);
  const { displayGraphData, focusNode, resetView, fullGraphData, focusedNodeIds } = useGraphData();
  const hasFocus = focusedNodeIds.size > 0;
  const { loading, codeViewPath, setCodeViewPath, timelineGlowFiles, timelineGraph } = useUIState();
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const [use3D, setUse3D] = useState(webglSupported);

  // When timeline is active, use the backend-computed graph; otherwise use focus-filtered graph
  const visibleGraphData = useMemo(() => {
    if (!timelineGraph) return displayGraphData;
    return {
      nodes: timelineGraph.nodes.map(n => ({ ...n })),
      links: timelineGraph.links.map(l => ({ ...l })),
    };
  }, [displayGraphData, timelineGraph]);

  // Double-click detection: defer single-click so double-click can cancel it
  const clickTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastClickRef = useRef<{ nodeId: string; time: number }>({ nodeId: '', time: 0 });

  // Track container size with ResizeObserver
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        if (width > 0 && height > 0) {
          setDimensions({ width: Math.floor(width), height: Math.floor(height) });
        }
      }
    });

    observer.observe(el);
    const rect = el.getBoundingClientRect();
    if (rect.width > 0 && rect.height > 0) {
      setDimensions({ width: Math.floor(rect.width), height: Math.floor(rect.height) });
    }

    return () => observer.disconnect();
  }, []);

  const handleNodeClick = useCallback(
    (node: GraphNode) => {
      const now = Date.now();
      const last = lastClickRef.current;

      if (last.nodeId === node.id && now - last.time < 300) {
        // Double-click: cancel pending single-click and open code viewer
        if (clickTimerRef.current) {
          clearTimeout(clickTimerRef.current);
          clickTimerRef.current = null;
        }
        setCodeViewPath(node.id);
        lastClickRef.current = { nodeId: '', time: 0 };
        return;
      }

      // Defer single-click action so double-click can cancel it
      lastClickRef.current = { nodeId: node.id, time: now };
      if (clickTimerRef.current) {
        clearTimeout(clickTimerRef.current);
      }
      clickTimerRef.current = setTimeout(() => {
        clickTimerRef.current = null;
        focusNode(node.id);
      }, 300);
    },
    [focusNode],
  );

  const handleBackgroundClick = useCallback(() => {
    resetView();
  }, [resetView]);

  // Auto-fit when data loads
  useEffect(() => {
    const ref = use3D ? graphRef3D : graphRef2D;
    if (fullGraphData.nodes.length > 0 && ref.current) {
      setTimeout(() => {
        ref.current?.zoomToFit(400, 50);
      }, 800);
    }
  }, [fullGraphData]);

  // Zoom to fit when focus changes
  useEffect(() => {
    const ref = use3D ? graphRef3D : graphRef2D;
    if (ref.current && visibleGraphData.nodes.length > 0) {
      setTimeout(() => {
        ref.current?.zoomToFit(400, 50);
      }, 100);
    }
  }, [focusedNodeIds, visibleGraphData.nodes.length]);

  // 2D glow animation: keep render loop alive so nodeCanvasObject redraws each frame
  const glowActive = !use3D && timelineGlowFiles.size > 0;
  useEffect(() => {
    if (!glowActive) return;
    // Kick the simulation so the render loop restarts with our Infinity cooldownTicks
    graphRef2D.current?.d3ReheatSimulation();
  }, [glowActive]);

  // Node color: highlight focused nodes, otherwise color by extension
  const nodeColor = useCallback(
    (node: object) => {
      const n = node as GraphNode;
      if (focusedNodeIds.has(n.id)) return HIGHLIGHT_COLOR;
      return getExtColor(n.name);
    },
    [focusedNodeIds],
  );

  // 2D: draw file name label + highlight ring after default circle
  const nodeCanvasObject2D = useCallback(
    (node: object, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const n = node as GraphNode;
      const x = n.x ?? 0;
      const y = n.y ?? 0;
      const fontSize = Math.max(10 / globalScale, 1.5);
      const nodeSize = Math.sqrt(scaleNodeVal(n.val)) * 4;
      const isFocused = focusedNodeIds.has(n.id);

      // Highlight ring for focused nodes
      if (isFocused) {
        ctx.beginPath();
        ctx.arc(x, y, nodeSize + 2, 0, 2 * Math.PI);
        ctx.strokeStyle = HIGHLIGHT_COLOR;
        ctx.lineWidth = 1.5 / globalScale;
        ctx.stroke();
      }

      // Breathing glow for timeline-changed files
      if (timelineGlowFiles.has(n.id)) {
        const pulse = 0.4 + 0.6 * Math.abs(Math.sin(Date.now() / 600));
        const glowRadius = nodeSize + 6;
        ctx.beginPath();
        ctx.arc(x, y, glowRadius, 0, 2 * Math.PI);
        ctx.fillStyle = `rgba(255, 255, 100, ${0.25 * pulse})`;
        ctx.fill();
        ctx.beginPath();
        ctx.arc(x, y, glowRadius, 0, 2 * Math.PI);
        ctx.strokeStyle = `rgba(255, 255, 100, ${0.7 * pulse})`;
        ctx.lineWidth = 1.5 / globalScale;
        ctx.stroke();
      }

      // Center dot: file size category indicator
      const dotRadius = Math.max(nodeSize * 0.35, 1.5);
      ctx.beginPath();
      ctx.arc(x, y, dotRadius, 0, 2 * Math.PI);
      ctx.fillStyle = getSizeColor(n.val);
      ctx.fill();

      // File name label below node
      ctx.font = `${fontSize}px Sans-Serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillStyle = isFocused ? HIGHLIGHT_COLOR : '#ccc';
      ctx.fillText(n.name, x, y + nodeSize + 2);
    },
    [focusedNodeIds, timelineGlowFiles],
  );

  // 3D: SpriteText label + center dot sphere + optional glow sphere
  const nodeThreeObject3D = useCallback(
    (node: object) => {
      const n = node as GraphNode;
      const group = new THREE.Group();

      // Center dot: small sphere with file-size color
      const dotRadius = scaleNodeVal(n.val) * 0.4;
      const dotGeom = new THREE.SphereGeometry(dotRadius, 12, 12);
      const dotMat = new THREE.MeshBasicMaterial({ color: getSizeColor(n.val) });
      const dotMesh = new THREE.Mesh(dotGeom, dotMat);
      group.add(dotMesh);

      // Glow sphere for timeline-changed files
      if (timelineGlowFiles.has(n.id)) {
        const glowRadius = scaleNodeVal(n.val) + 2;
        const glowGeom = new THREE.SphereGeometry(glowRadius, 16, 16);
        const glowMat = new THREE.MeshBasicMaterial({
          color: 0xffff64,
          transparent: true,
          opacity: 0.35,
        });
        const glowMesh = new THREE.Mesh(glowGeom, glowMat);
        glowMesh.userData.isGlow = true;
        group.add(glowMesh);
      }

      // Text label above
      const sprite = new SpriteText(n.name);
      sprite.color = focusedNodeIds.has(n.id) ? HIGHLIGHT_COLOR : '#ccc';
      sprite.textHeight = 2;
      sprite.position.y = scaleNodeVal(n.val) + 3;
      group.add(sprite);

      return group;
    },
    [focusedNodeIds, timelineGlowFiles],
  );

  // 3D: Animate glow meshes with breathing effect
  const onEngineTick3D = useCallback(() => {
    if (timelineGlowFiles.size === 0 || !graphRef3D.current) return;
    const scene = graphRef3D.current.scene?.();
    if (!scene) return;
    const pulse = 0.2 + 0.8 * Math.abs(Math.sin(Date.now() / 600));
    scene.traverse((obj: THREE.Object3D) => {
      if (obj.userData.isGlow && obj instanceof THREE.Mesh) {
        (obj.material as THREE.MeshBasicMaterial).opacity = 0.35 * pulse;
      }
    });
  }, [timelineGlowFiles]);

  const hasData = visibleGraphData.nodes.length > 0;

  // Build legend from current extension-color assignments
  const legend = useMemo(() => {
    if (!hasData) return [];
    const seen = new Set<string>();
    for (const node of visibleGraphData.nodes) {
      const dot = node.name.lastIndexOf('.');
      if (dot > 0) seen.add(node.name.substring(dot).toLowerCase());
    }
    return Array.from(seen)
      .sort()
      .map((ext) => ({ ext, color: getExtColor(`file${ext}`) }));
  }, [visibleGraphData, hasData]);

  return (
    <div ref={containerRef} className="w-full h-full bg-dark-bg relative">
      {/* Loading overlay */}
      {loading && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-dark-bg/80">
          <div className="flex flex-col items-center gap-3">
            <Loader2 size={32} className="animate-spin text-accent-primary" />
            <p className="text-dark-text-secondary text-sm">Cloning and analyzing repository...</p>
          </div>
        </div>
      )}

      {hasData ? (
        <>
          {use3D ? (
            <Suspense fallback={
              <div className="flex items-center justify-center h-full">
                <Loader2 size={24} className="animate-spin text-accent-primary" />
              </div>
            }>
              <ForceGraph3D
                ref={graphRef3D}
                width={dimensions.width}
                height={dimensions.height}
                graphData={visibleGraphData}
                nodeLabel={(node: object) => (node as GraphNode).name}
                nodeColor={nodeColor}
                nodeVal={(node: object) => scaleNodeVal((node as GraphNode).val)}
                nodeOpacity={0.9}
                nodeThreeObjectExtend={true}
                nodeThreeObject={nodeThreeObject3D}
                linkColor={() => '#555'}
                linkWidth={1}
                linkOpacity={0.4}
                linkDirectionalArrowLength={3}
                linkDirectionalArrowRelPos={1}
                onNodeClick={(node: object) => handleNodeClick(node as GraphNode)}
                onBackgroundClick={handleBackgroundClick}
                onEngineTick={onEngineTick3D}
                showNavInfo={false}
                backgroundColor="#0a0a0a"
              />
            </Suspense>
          ) : (
            <ForceGraph2D
              ref={graphRef2D}
              width={dimensions.width}
              height={dimensions.height}
              graphData={visibleGraphData}
              nodeLabel={(node: object) => (node as GraphNode).name}
              nodeColor={nodeColor}
              nodeVal={(node: object) => scaleNodeVal((node as GraphNode).val)}
              nodeCanvasObjectMode={() => 'after'}
              nodeCanvasObject={nodeCanvasObject2D}
              linkColor={() => '#555'}
              linkWidth={1}
              linkDirectionalArrowLength={4}
              linkDirectionalArrowRelPos={1}
              onNodeClick={(node: object) => handleNodeClick(node as GraphNode)}
              onBackgroundClick={handleBackgroundClick}
              backgroundColor="#0a0a0a"
              cooldownTicks={glowActive ? Infinity : undefined}
            />
          )}

          {/* Clear Selection button */}
          {hasFocus && (
            <button
              onClick={resetView}
              className="absolute top-3 left-3 z-10 flex items-center gap-1.5 px-2.5 py-1 text-xs rounded bg-dark-bg/70 text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg/90 border border-dark-border transition-colors"
              aria-label="Clear selection"
            >
              <X size={14} />
              Clear Selection
            </button>
          )}

          {/* Stats overlay + 2D/3D toggle */}
          <div className="absolute top-3 right-3 flex items-center gap-2">
            <div className="text-xs text-dark-text-secondary bg-dark-bg/70 px-2 py-1 rounded">
              {visibleGraphData.nodes.length} nodes / {visibleGraphData.links.length} edges
              {hasFocus && (
                <span className="ml-2 text-accent-primary">({focusedNodeIds.size} focused)</span>
              )}
            </div>
            <button
              onClick={() => setUse3D(!use3D)}
              disabled={!webglSupported}
              className={`flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors ${
                webglSupported
                  ? 'bg-dark-bg/70 text-dark-text-secondary hover:text-dark-text hover:bg-dark-bg/90 border border-dark-border cursor-pointer'
                  : 'bg-dark-bg/40 text-dark-text-secondary/40 border border-dark-border/40 cursor-not-allowed'
              }`}
              title={webglSupported ? `Switch to ${use3D ? '2D' : '3D'} mode` : 'WebGL not supported'}
            >
              {use3D ? <Square size={12} /> : <Box size={12} />}
              {use3D ? '2D' : '3D'}
            </button>
          </div>

          {/* Bottom-left legends: file types on top, lines on bottom */}
          {hasData && (
            <div className="absolute bottom-3 left-3 text-xs flex flex-col gap-2">
              {legend.length > 0 && (
                <div className="bg-dark-bg/80 px-3 py-2 rounded border border-dark-border max-h-48 overflow-auto">
                  {legend.map(({ ext, color }) => (
                    <div key={ext} className="flex items-center gap-2 py-0.5">
                      <span
                        className="w-2.5 h-2.5 rounded-full inline-block shrink-0"
                        style={{ backgroundColor: color }}
                      />
                      <span className="text-dark-text-secondary">{ext}</span>
                    </div>
                  ))}
                </div>
              )}
              <div className="bg-dark-bg/80 px-3 py-2 rounded border border-dark-border">
                <div className="text-dark-text-secondary mb-1 font-medium">Lines</div>
                {SIZE_CATEGORIES.map((cat) => (
                  <div key={cat.label} className="flex items-center gap-2 py-0.5">
                    <span
                      className="w-2.5 h-2.5 rounded-full inline-block shrink-0"
                      style={{ backgroundColor: cat.color }}
                    />
                    <span className="text-dark-text-secondary">{cat.label}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Code viewer overlay */}
          {codeViewPath && (
            <CodeOverlay
              path={codeViewPath}
              onClose={() => setCodeViewPath(null)}
            />
          )}
        </>
      ) : (
        <div className="flex items-center justify-center h-full text-dark-text-secondary">
          <div className="text-center max-w-md">
            <img src="/qodex_logo.png" alt="Qodex" className="w-48 h-48 mx-auto mb-6 object-contain opacity-80" />
            <img src="/qodex_txt.png" alt="Qodex" className="h-[51px] mx-auto mb-2" />
            <p className="text-sm leading-relaxed">
              Enter a public GitHub repository URL above to visualize its
              dependency graph in 3D. Click nodes to focus, click background to reset.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
