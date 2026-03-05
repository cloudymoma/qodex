import { useState, useEffect } from 'react';
import { useUIState } from '@/contexts/UIStateContext';
import { api } from '@/services/api';
import type { CommitEntry } from '@/types';
import { GitCommit, Loader2, User, Calendar, ExternalLink } from 'lucide-react';

export function HistoryPanel() {
  const [commits, setCommits] = useState<CommitEntry[]>([]);
  const [repoURL, setRepoURL] = useState('');
  const [loading, setLoading] = useState(false);
  const { repoName } = useUIState();

  useEffect(() => {
    if (!repoName) return;
    setLoading(true);
    api.getHistory(50)
      .then((resp) => {
        setCommits(resp.commits || []);
        setRepoURL(resp.repo_url || '');
      })
      .catch(() => setCommits([]))
      .finally(() => setLoading(false));
  }, [repoName]);

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
    <div className="h-full overflow-auto">
      {commits.map((c) => {
        const url = commitURL(c.hash);
        return (
          <a
            key={c.hash}
            href={url || undefined}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-start gap-3 px-4 py-2 border-b border-dark-border hover:bg-dark-bg-tertiary transition-colors cursor-pointer group"
          >
            <GitCommit size={14} className="text-accent-primary mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="text-dark-text text-sm truncate group-hover:text-accent-primary transition-colors">
                {c.message}
              </div>
              <div className="flex items-center gap-3 text-dark-text-secondary text-xs mt-0.5">
                <span className="font-mono text-accent-primary">{c.short}</span>
                <span className="flex items-center gap-1"><User size={10} />{c.author}</span>
                <span className="flex items-center gap-1"><Calendar size={10} />{c.date}</span>
              </div>
            </div>
            <ExternalLink size={12} className="text-dark-text-secondary opacity-0 group-hover:opacity-100 transition-opacity mt-1 shrink-0" />
          </a>
        );
      })}
    </div>
  );
}
