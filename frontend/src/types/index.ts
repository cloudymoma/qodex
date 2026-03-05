// Graph types matching backend API
export interface GraphNode {
  id: string;
  name: string;
  group: number;
  val?: number;
  // Force-graph positioning (managed internally)
  x?: number;
  y?: number;
  z?: number;
}

// Note: D3-force mutates source/target from strings to node objects after first render
// We handle this with getLinkId() helper in GraphDataContext
export interface GraphLink {
  source: string | GraphNode;
  target: string | GraphNode;
}

export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

// Tree types
export interface TreeNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  children?: TreeNode[];
}

// Search types
export interface SearchResult {
  file_path: string;
  file_name: string;
  score: number;
  matches: MatchFragment[];
}

export interface MatchFragment {
  line_number: number;
  line: string;
}

export interface SearchResponse {
  query: string;
  results: SearchResult[];
  total: number;
}

// Ingest types
export interface IngestRequest {
  url: string;
  branch?: string;
}

export interface IngestResponse {
  repo_name: string;
  status: 'success' | 'error';
  message?: string;
  files_indexed?: number;
}

// Repo history
export interface RepoEntry {
  url: string;
  branch: string;
  repo_name: string;
}

// File viewer types
export interface FileResponse {
  path: string;
  content: string;
  language: string;
}

// UI types
export type BottomPanelMode = 'search' | 'chat' | 'hidden';
