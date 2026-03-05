import type {
  GraphData,
  TreeNode,
  SearchResponse,
  IngestRequest,
  IngestResponse,
  RepoEntry,
  FileResponse,
} from '@/types';

const API_BASE = '';

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${url}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API error ${res.status}: ${text}`);
  }

  return res.json();
}

export const api = {
  ingest(req: IngestRequest): Promise<IngestResponse> {
    return fetchJSON('/api/ingest', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  getGraph(): Promise<GraphData> {
    return fetchJSON('/api/graph');
  },

  getTree(): Promise<TreeNode[]> {
    return fetchJSON('/api/tree');
  },

  search(query: string, limit = 20): Promise<SearchResponse> {
    return fetchJSON(`/api/search?q=${encodeURIComponent(query)}&limit=${limit}`);
  },

  listRepos(): Promise<RepoEntry[]> {
    return fetchJSON('/api/repos');
  },

  getFile(path: string): Promise<FileResponse> {
    return fetchJSON(`/api/file?path=${encodeURIComponent(path)}`);
  },
};
