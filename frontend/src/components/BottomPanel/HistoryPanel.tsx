import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useUIState } from '@/contexts/UIStateContext';
import { api } from '@/services/api';
import type { CommitEntry } from '@/types';
import { GitCommit, Loader2, User, Calendar, ExternalLink, Clock, RotateCcw } from 'lucide-react';
import { track } from '@/services/tracker';

const DEBOUNCE_MS = 800;

export function HistoryPanel() {
  const [commits, setCommits] = useState<CommitEntry[]>([]);
  const [repoURL, setRepoURL] = useState('');
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const { repoName, setTimelineGlowFiles, setTimelineGraph } = useUIState();

  // Slider position: index into chronoCommits (oldest=0, latest=max). Default = max (latest).
  const [sliderPos, setSliderPos] = useState(-1); // -1 = pending init

  // Refs for debounce timer and AbortController
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  // Commits in chronological order (oldest first) for timeline
  const chronoCommits = useMemo(() => [...commits].reverse(), [commits]);

  // Fetch commit history when repo changes
  useEffect(() => {
    if (!repoName) return;
    setLoading(true);
    setSliderPos(-1);
    setTimelineGraph(null);
    setTimelineGlowFiles(new Set());
    api.getHistory(0)
      .then((resp) => {
        const c = resp.commits || [];
        setCommits(c);
        setRepoURL(resp.repo_url || '');
        // Default slider to latest commit (rightmost)
        setSliderPos(c.length > 0 ? c.length - 1 : 0);
      })
      .catch(() => setCommits([]))
      .finally(() => setLoading(false));
  }, [repoName, setTimelineGraph, setTimelineGlowFiles]);

  // Debounced backend call for timeline graph
  const fetchTimelineGraph = useCallback(
    (pos: number) => {
      // Cancel pending debounce
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
        debounceRef.current = null;
      }

      // At latest commit or invalid: show full graph (no timeline filtering)
      if (pos < 0 || pos >= chronoCommits.length - 1 || chronoCommits.length === 0) {
        // Cancel in-flight request
        abortRef.current?.abort();
        abortRef.current = null;
        setTimelineGraph(null);
        setTimelineGlowFiles(new Set());
        setFetching(false);
        return;
      }

      // Compute cumulative file set and glow files immediately (cheap)
      const cumulative: string[] = [];
      const seen = new Set<string>();
      for (let i = 0; i <= pos && i < chronoCommits.length; i++) {
        for (const f of chronoCommits[i]?.files_changed || []) {
          if (!seen.has(f)) {
            seen.add(f);
            cumulative.push(f);
          }
        }
      }
      const glowing = new Set<string>(chronoCommits[pos]?.files_changed || []);
      setTimelineGlowFiles(glowing);

      // Debounce the backend call
      debounceRef.current = setTimeout(() => {
        // Cancel previous in-flight request
        abortRef.current?.abort();
        const controller = new AbortController();
        abortRef.current = controller;

        setFetching(true);
        api.getTimelineGraph(cumulative, controller.signal)
          .then((graph) => {
            if (!controller.signal.aborted) {
              setTimelineGraph(graph);
            }
          })
          .catch((err) => {
            if (err instanceof DOMException && err.name === 'AbortError') return;
            console.warn('timeline graph fetch failed', err);
          })
          .finally(() => {
            if (!controller.signal.aborted) {
              setFetching(false);
            }
          });
      }, DEBOUNCE_MS);
    },
    [chronoCommits, setTimelineGraph, setTimelineGlowFiles],
  );

  // Handle slider change
  const handleSliderChange = useCallback(
    (pos: number) => {
      setSliderPos(pos);
      track('timeline_slide', 'history', `${pos}/${chronoCommits.length}`);
      fetchTimelineGraph(pos);
    },
    [fetchTimelineGraph, chronoCommits.length],
  );

  // Keyboard: left/right arrows move one commit
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
        e.preventDefault();
        const delta = e.key === 'ArrowRight' ? 1 : -1;
        setSliderPos((prev) => {
          const cur = prev >= 0 ? prev : chronoCommits.length - 1;
          const next = Math.max(0, Math.min(chronoCommits.length - 1, cur + delta));
          fetchTimelineGraph(next);
          return next;
        });
      }
    },
    [chronoCommits.length, fetchTimelineGraph],
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
      abortRef.current?.abort();
      setTimelineGraph(null);
      setTimelineGlowFiles(new Set());
    };
  }, [setTimelineGraph, setTimelineGlowFiles]);

  const effectivePos = sliderPos >= 0 ? sliderPos : chronoCommits.length - 1;
  const currentCommit = chronoCommits[effectivePos] ?? null;

  if (!repoName) {
    return (
      <div className="flex items-center justify-center h-full text-dark-text-secondary text-sm">
        Load a repository to view commit history
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 size={20} className="animate-spin text-accent-primary" />
      </div>
    );
  }

  if (commits.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-dark-text-secondary text-sm">
        No commits found
      </div>
    );
  }

  const commitURL = (hash: string) =>
    repoURL ? `${repoURL}/commit/${hash}` : '';

  return (
    <div className="flex flex-col h-full">
      {/* Timeline slider */}
      <div className="px-4 py-2 border-b border-dark-border bg-dark-bg shrink-0">
        <div className="flex items-center gap-3">
          <Clock size={14} className="text-accent-primary shrink-0" />
          <span className="text-xs text-dark-text-secondary shrink-0">Time Travel</span>
          <input
            type="range"
            min={0}
            max={chronoCommits.length - 1}
            value={sliderPos >= 0 ? sliderPos : chronoCommits.length - 1}
            onChange={(e) => handleSliderChange(Number(e.target.value))}
            onKeyDown={handleKeyDown}
            className="flex-1 accent-accent-primary h-1 cursor-pointer"
          />
          <span className="text-xs text-dark-text-secondary shrink-0 w-24 text-right flex items-center gap-1 justify-end">
            {fetching && <Loader2 size={10} className="animate-spin" />}
            {`${(sliderPos >= 0 ? sliderPos : chronoCommits.length - 1) + 1} / ${chronoCommits.length}`}
          </span>
          <button
            onClick={() => handleSliderChange(chronoCommits.length - 1)}
            disabled={sliderPos >= chronoCommits.length - 1}
            className="shrink-0 text-xs px-2 py-0.5 rounded bg-dark-bg-tertiary text-dark-text-secondary hover:text-dark-text hover:bg-dark-border disabled:opacity-30 disabled:cursor-not-allowed transition-colors flex items-center gap-1"
            title="Reset to latest"
          >
            <RotateCcw size={10} />
            Reset
          </button>
        </div>
        {currentCommit && (
          <div className="mt-1 text-xs text-dark-text-secondary truncate pl-7">
            <a
              href={commitURL(currentCommit.hash) || undefined}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-accent-primary hover:underline"
              onClick={(e) => e.stopPropagation()}
            >{currentCommit.short}</a>
            {' '}{currentCommit.message}
            {' '}— <span className="text-yellow-300">{currentCommit.files_changed?.length || 0} files changed</span>
            {' '}· <span>{currentCommit.date}</span>
          </div>
        )}
      </div>

      {/* Commit list */}
      <div className="flex-1 overflow-auto">
        {commits.map((c, i) => {
          const url = commitURL(c.hash);
          const isActive = currentCommit?.hash === c.hash;
          const chronoIdx = commits.length - 1 - i;
          return (
            <a
              key={c.hash}
              href={url || undefined}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => {
                if (!e.metaKey && !e.ctrlKey) {
                  e.preventDefault();
                  handleSliderChange(chronoIdx);
                }
              }}
              className={`flex items-start gap-3 px-4 py-2 border-b border-dark-border hover:bg-dark-bg-tertiary transition-colors cursor-pointer group ${
                isActive ? 'bg-accent-primary/10' : ''
              }`}
            >
              <GitCommit size={14} className={`mt-0.5 shrink-0 ${isActive ? 'text-yellow-300' : 'text-accent-primary'}`} />
              <div className="flex-1 min-w-0">
                <div className={`text-sm truncate transition-colors ${isActive ? 'text-yellow-300' : 'text-dark-text group-hover:text-accent-primary'}`}>
                  {c.message}
                </div>
                <div className="flex items-center gap-3 text-dark-text-secondary text-xs mt-0.5">
                  <span className="font-mono text-accent-primary">{c.short}</span>
                  <span className="flex items-center gap-1"><User size={10} />{c.author}</span>
                  <span className="flex items-center gap-1"><Calendar size={10} />{c.date}</span>
                  {c.files_changed && (
                    <span className="text-dark-text-secondary">{c.files_changed.length} files</span>
                  )}
                </div>
              </div>
              <ExternalLink size={12} className="text-dark-text-secondary opacity-0 group-hover:opacity-100 transition-opacity mt-1 shrink-0" />
            </a>
          );
        })}
      </div>
    </div>
  );
}
