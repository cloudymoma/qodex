import { useState } from 'react';
import { api } from '@/services/api';
import { useGraphData } from '@/contexts/GraphDataContext';
import { useUIState } from '@/contexts/UIStateContext';
import type { SearchResult } from '@/types';
import { Search, File } from 'lucide-react';

export function SearchPanel() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const { focusNode } = useGraphData();
  const { setSearchQuery } = useUIState();

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setSearching(true);
    setSearchQuery(query.trim());
    try {
      const resp = await api.search(query.trim());
      setResults(resp.results || []);
    } catch (err) {
      console.error('Search failed:', err);
    } finally {
      setSearching(false);
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Search input */}
      <form onSubmit={handleSearch} className="flex items-center gap-2 px-4 py-2">
        <Search size={14} className="text-dark-text-secondary shrink-0" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search code (functions, variables, comments...)"
          className="flex-1 bg-dark-bg-tertiary text-dark-text px-3 py-1 rounded border border-dark-border focus:border-accent-primary focus:outline-none text-sm"
        />
        <button
          type="submit"
          disabled={searching}
          className="px-3 py-1 bg-accent-primary text-white rounded text-sm hover:bg-accent-primary/80 disabled:opacity-50"
        >
          {searching ? '...' : 'Search'}
        </button>
      </form>

      {/* Results */}
      <div className="flex-1 overflow-auto px-4">
        {results.length === 0 && !searching && (
          <p className="text-dark-text-secondary text-sm py-2">
            {query ? 'No results found.' : 'Type a query to search the codebase.'}
          </p>
        )}
        {results.map((result, i) => (
          <button
            key={i}
            onClick={() => focusNode(result.file_path)}
            className="w-full text-left flex items-start gap-2 px-2 py-1.5 rounded hover:bg-dark-bg-tertiary transition-colors"
          >
            <File size={14} className="text-dark-text-secondary mt-0.5 shrink-0" />
            <div className="min-w-0">
              <div className="text-sm text-dark-text truncate">{result.file_path}</div>
              {result.matches?.map((m, j) => (
                <div key={j} className="text-xs text-dark-text-secondary font-mono truncate">
                  L{m.line_number}: {m.line}
                </div>
              ))}
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
