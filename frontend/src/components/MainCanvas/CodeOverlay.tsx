import { useEffect, useState, useMemo, createElement, type ReactNode } from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { X, Loader2 } from 'lucide-react';
import { api } from '@/services/api';
import { useUIState } from '@/contexts/UIStateContext';

interface CodeOverlayProps {
  path: string;
  onClose: () => void;
}

// --- Search highlight helpers for react-syntax-highlighter custom renderer ---

/** Split text around case-insensitive matches of `queryLower`, wrapping hits in <mark>. */
function highlightText(text: string, queryLower: string, key: number): ReactNode {
  const lower = text.toLowerCase();
  const first = lower.indexOf(queryLower);
  if (first === -1) return text;

  const parts: ReactNode[] = [];
  let last = 0;
  let idx = first;
  let pk = 0;

  while (idx !== -1) {
    if (idx > last) parts.push(text.slice(last, idx));
    parts.push(
      createElement('mark', {
        key: `m${key}-${pk++}`,
        style: {
          backgroundColor: 'rgba(255,200,0,0.45)',
          color: 'inherit',
          borderRadius: '2px',
          padding: '0 1px',
        },
      }, text.slice(idx, idx + queryLower.length)),
    );
    last = idx + queryLower.length;
    idx = lower.indexOf(queryLower, last);
  }

  if (last < text.length) parts.push(text.slice(last));
  return createElement('span', { key: `w${key}` }, ...parts);
}

/** Resolve HAST node classNames to inline styles via the syntax theme stylesheet. */
function resolveStyle(
  properties: Record<string, any> | undefined,
  stylesheet: Record<string, React.CSSProperties>,
): React.CSSProperties {
  let style: Record<string, any> = {};
  const classNames: any[] = properties?.className || [];
  for (const cn of classNames) {
    if (typeof cn === 'string' && stylesheet[cn]) {
      style = { ...style, ...stylesheet[cn] };
    }
  }
  if (properties?.style) {
    style = { ...style, ...properties.style };
  }
  return style;
}

/** Recursively render a HAST node, injecting keyword highlights into text nodes
 *  while preserving syntax theme colors from the stylesheet. */
function renderNode(
  node: { type: string; value?: string | number; tagName?: string; properties?: Record<string, any>; children?: any[] },
  key: number,
  queryLower: string,
  stylesheet: Record<string, React.CSSProperties>,
): ReactNode {
  if (node.type === 'text') {
    const text = String(node.value ?? '');
    return highlightText(text, queryLower, key);
  }
  const children = (node.children || []).map((c: any, i: number) =>
    renderNode(c, i, queryLower, stylesheet),
  );
  return createElement(
    node.tagName || 'span',
    { key, style: resolveStyle(node.properties, stylesheet) },
    ...children,
  );
}

/** Collect plain-text content of a HAST subtree. */
function textContent(node: { type: string; value?: string | number; children?: any[] }): string {
  if (node.type === 'text') return String(node.value ?? '');
  return (node.children || []).map(textContent).join('');
}

// --- Component ---

export function CodeOverlay({ path, onClose }: CodeOverlayProps) {
  const [content, setContent] = useState<string>('');
  const [language, setLanguage] = useState<string>('text');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { searchQuery } = useUIState();

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    api.getFile(path)
      .then((resp) => {
        if (cancelled) return;
        setContent(resp.content);
        setLanguage(resp.language);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [path]);

  // Close on Escape key
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

  // Count matches for the header badge
  const matchCount = useMemo(() => {
    if (!searchQuery || !content) return 0;
    const lower = content.toLowerCase();
    const qLower = searchQuery.toLowerCase();
    let count = 0;
    let idx = lower.indexOf(qLower);
    while (idx !== -1) {
      count++;
      idx = lower.indexOf(qLower, idx + qLower.length);
    }
    return count;
  }, [searchQuery, content]);

  // Custom renderer: highlight search keywords inside syntax-highlighted tokens
  const renderer = useMemo(() => {
    if (!searchQuery) return undefined;
    const qLower = searchQuery.toLowerCase();

    return ({ rows, stylesheet }: { rows: any[]; stylesheet: Record<string, React.CSSProperties> }) => {
      return rows.map((row: any, i: number) => {
        const lineText = textContent(row);
        const hasMatch = lineText.toLowerCase().includes(qLower);
        const children = (row.children || []).map((c: any, j: number) =>
          renderNode(c, j, qLower, stylesheet),
        );
        const rowStyle = resolveStyle(row.properties, stylesheet);
        return createElement(
          'span',
          {
            key: i,
            style: {
              ...rowStyle,
              display: 'block',
              ...(hasMatch ? { backgroundColor: 'rgba(255,200,0,0.07)' } : {}),
            },
          },
          ...children,
        );
      });
    };
  }, [searchQuery]);

  return (
    <div className="absolute inset-0 z-30 flex flex-col bg-dark-bg/95 backdrop-blur-sm">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-dark-surface border-b border-dark-border shrink-0">
        <div className="flex items-center gap-2 min-w-0 mr-4">
          <span className="text-sm text-dark-text font-mono truncate">{path}</span>
          {matchCount > 0 && (
            <span className="shrink-0 text-xs px-1.5 py-0.5 rounded bg-yellow-500/20 text-yellow-300">
              {matchCount} match{matchCount !== 1 ? 'es' : ''}
            </span>
          )}
        </div>
        <button
          onClick={onClose}
          className="p-1 rounded hover:bg-dark-border text-dark-text-secondary hover:text-dark-text transition-colors"
          aria-label="Close code viewer"
        >
          <X size={18} />
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {loading && (
          <div className="flex items-center justify-center h-full">
            <Loader2 size={24} className="animate-spin text-accent-primary" />
          </div>
        )}

        {error && (
          <div className="flex items-center justify-center h-full">
            <p className="text-red-400 text-sm">{error}</p>
          </div>
        )}

        {!loading && !error && (
          <SyntaxHighlighter
            language={language}
            style={vscDarkPlus}
            showLineNumbers
            wrapLines
            renderer={renderer}
            customStyle={{
              margin: 0,
              padding: '1rem',
              background: 'transparent',
              fontSize: '13px',
              minHeight: '100%',
            }}
          >
            {content}
          </SyntaxHighlighter>
        )}
      </div>
    </div>
  );
}
