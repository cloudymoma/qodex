import { useState, useRef, useCallback, useEffect } from 'react';
import { useGraphData } from '@/contexts/GraphDataContext';
import { useUIState } from '@/contexts/UIStateContext';
import { api } from '@/services/api';
import type { RepoEntry } from '@/types';
import { GitBranch, Loader2, AlertCircle, History } from 'lucide-react';

export function RepoInput() {
  const [url, setUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [error, setError] = useState<string | null>(null);
  const [repos, setRepos] = useState<RepoEntry[]>([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const { setFullGraphData } = useGraphData();
  const { setTreeData, setLoading, loading, setRepoName } = useUIState();
  const dropdownRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const fetchRepos = useCallback(async () => {
    try {
      const list = await api.listRepos();
      setRepos(list);
    } catch {
      // silently ignore — dropdown just won't show
    }
  }, []);

  const doIngest = useCallback(async (ingestUrl: string, ingestBranch: string) => {
    if (!ingestUrl.trim() || loading) return;

    setLoading(true);
    setError(null);

    try {
      const ingestResp = await api.ingest({ url: ingestUrl.trim(), branch: ingestBranch });

      if (ingestResp.status === 'error') {
        setError(ingestResp.message || 'Ingest failed');
        return;
      }

      setRepoName(ingestResp.repo_name);

      const [graphData, treeData] = await Promise.all([
        api.getGraph(),
        api.getTree(),
      ]);

      setFullGraphData(graphData);
      setTreeData(treeData);

      // Refresh repo list after successful ingest
      fetchRepos();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Unknown error';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [loading, setLoading, setRepoName, setFullGraphData, setTreeData, fetchRepos]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setShowDropdown(false);
    doIngest(url, branch);
  };

  const handleSelect = (repo: RepoEntry) => {
    setUrl(repo.url);
    setBranch(repo.branch);
    setShowDropdown(false);
    setError(null);
    doIngest(repo.url, repo.branch);
  };

  const handleFocus = () => {
    fetchRepos();
    setShowDropdown(true);
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const filteredRepos = url.trim()
    ? repos.filter((r) => r.url.toLowerCase().includes(url.toLowerCase()) || r.repo_name.toLowerCase().includes(url.toLowerCase()))
    : repos;

  return (
    <div className="bg-dark-bg-secondary border-b border-dark-border relative">
      <form
        onSubmit={handleSubmit}
        className="flex items-center gap-3 px-4 py-2"
      >
        <img
          src="/qodex_txt.png"
          alt="Qodex"
          className="h-7 shrink-0"
        />

        <GitBranch size={18} className="text-accent-primary shrink-0" />

        <div className="flex-1 relative">
          <input
            ref={inputRef}
            type="text"
            value={url}
            onChange={(e) => { setUrl(e.target.value); setError(null); setShowDropdown(true); }}
            onFocus={handleFocus}
            placeholder="Enter GitHub URL (e.g. https://github.com/owner/repo)"
            className="w-full bg-dark-bg-tertiary text-dark-text px-3 py-1.5 rounded border border-dark-border focus:border-accent-primary focus:outline-none text-sm"
          />

          {/* Repo history dropdown */}
          {showDropdown && filteredRepos.length > 0 && !loading && (
            <div
              ref={dropdownRef}
              className="absolute top-full left-0 right-0 mt-1 bg-dark-bg-tertiary border border-dark-border rounded shadow-lg z-50 max-h-60 overflow-auto"
            >
              <div className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-dark-text-secondary border-b border-dark-border">
                <History size={12} />
                Previously explored
              </div>
              {filteredRepos.map((repo) => (
                <button
                  key={repo.repo_name}
                  type="button"
                  onClick={() => handleSelect(repo)}
                  className="w-full text-left px-3 py-2 hover:bg-dark-bg-secondary transition-colors text-sm flex items-center justify-between gap-2"
                >
                  <span className="text-dark-text truncate">{repo.url}</span>
                  <span className="text-dark-text-secondary text-xs shrink-0">{repo.branch}</span>
                </button>
              ))}
            </div>
          )}
        </div>

        <input
          type="text"
          value={branch}
          onChange={(e) => setBranch(e.target.value)}
          placeholder="branch"
          className="w-24 bg-dark-bg-tertiary text-dark-text px-3 py-1.5 rounded border border-dark-border focus:border-accent-primary focus:outline-none text-sm"
        />

        <button
          type="submit"
          disabled={loading || !url.trim()}
          className="px-4 py-1.5 bg-accent-primary text-white rounded text-sm font-medium hover:bg-accent-primary/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
        >
          {loading ? (
            <>
              <Loader2 size={14} className="animate-spin" />
              Loading...
            </>
          ) : (
            'Explore'
          )}
        </button>
      </form>

      {/* Error banner */}
      {error && (
        <div className="flex items-center gap-2 px-4 py-1.5 bg-accent-error/10 border-t border-accent-error/30 text-accent-error text-sm">
          <AlertCircle size={14} className="shrink-0" />
          <span className="truncate">{error}</span>
          <button
            onClick={() => setError(null)}
            className="ml-auto text-xs hover:underline shrink-0"
          >
            Dismiss
          </button>
        </div>
      )}
    </div>
  );
}
