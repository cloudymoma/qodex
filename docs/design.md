# Qodex - Technical Design Document

## 1. Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                     Browser (React SPA)                      │
│  ┌──────────┐  ┌─────────────────┐  ┌─────────────────────┐ │
│  │ FileTree │  │ ForceGraph3D    │  │ BottomPanel         │ │
│  │ (Left)   │  │ (Center/Right)  │  │ Search | Chat       │ │
│  └──────────┘  └─────────────────┘  └─────────────────────┘ │
└──────────────────────────┬───────────────────────────────────┘
                           │ HTTP (port 1983)
┌──────────────────────────▼───────────────────────────────────┐
│                     Go HTTP Server                           │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Middleware: Recovery → Logger → CORS                 │    │
│  └──────────────────────────────────────────────────────┘    │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐   │
│  │ /ingest │ │ /graph  │ │ /tree   │ │ /search         │   │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────────┬────────┘   │
│       │           │           │                │             │
│  ┌────▼────────────▼───────────▼────────────────▼────────┐   │
│  │              IngestService (Orchestrator)              │   │
│  │  ┌──────────┐ ┌──────────┐ ┌────────────┐            │   │
│  │  │ go-git   │ │ Parser   │ │ Bleve      │            │   │
│  │  │ (Clone)  │ │ (Deps)   │ │ (Index)    │            │   │
│  │  └──────────┘ └──────────┘ └────────────┘            │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                              │
│  Storage: $HOME/.qodex/<repo>/                       │
│  Indexes: $HOME/.qodex/.indexes/<repo>/              │
└──────────────────────────────────────────────────────────────┘
```

## 2. Technology Stack

| Layer      | Technology                  | Purpose                              |
|------------|-----------------------------|--------------------------------------|
| Backend    | Go 1.23+ (net/http)         | API server, static file serving      |
| Frontend   | React 19 + Vite + TypeScript| SPA with 3D visualization            |
| 3D Engine  | react-force-graph-3d        | WebGL force-directed graph           |
| CSS        | Tailwind CSS v4             | Dark theme, utility-first styling    |
| Search     | Bleve v2                    | Full-text indexing and search        |
| Git        | go-git/go-git v5            | Repository cloning                   |
| Config     | gopkg.in/yaml.v3            | YAML configuration                   |

## 3. Backend Design

### 3.1 Package Structure

```
cmd/server/main.go          — Entry point, graceful shutdown
internal/
  config/config.go           — YAML config loader with env expansion
  api/
    router.go                — HTTP router and middleware wiring
    handler/
      ingest.go              — POST /api/ingest
      graph.go               — GET /api/graph
      tree.go                — GET /api/tree
      search.go              — GET /api/search
    middleware/
      cors.go                — CORS handling
      logger.go              — Request logging (slog)
      recovery.go            — Panic recovery
  repository/
    repository.go            — Repository interface
    git.go                   — go-git clone/pull implementation
  parser/
    parser.go                — Parser interface + file types
    golang.go                — Go import regex parser
    registry.go              — Language parser registry
  indexer/
    indexer.go               — Indexer interface
    bleve.go                 — Bleve implementation
  graph/
    types.go                 — Graph data structures
    builder.go               — Build graph from parsed dependencies
  service/
    ingest.go                — Orchestrate clone→parse→index pipeline
pkg/models/
  graph.go                   — GraphData, Node, Link (JSON API models)
  tree.go                    — TreeNode (JSON API model)
  search.go                  — SearchResult (JSON API model)
  ingest.go                  — IngestRequest/Response
```

### 3.2 API Endpoints

| Method | Path           | Description                              | Request              | Response            |
|--------|----------------|------------------------------------------|----------------------|---------------------|
| POST   | /api/ingest    | Clone repo, build index + dependency graph | `IngestRequest`     | `IngestResponse`    |
| GET    | /api/graph     | Get graph nodes and links                | -                    | `GraphData`         |
| GET    | /api/tree      | Get file tree hierarchy                  | -                    | `[]TreeNode`        |
| GET    | /api/search    | Search indexed code                      | `?q=keyword`         | `SearchResponse`    |

### 3.3 Ingest Pipeline

```
POST /api/ingest { url, branch }
  │
  ├─ 1. Validate URL, extract owner/repo
  ├─ 2. Clone via go-git to ~/.qodex/<owner>-<repo>/
  ├─ 3. Walk directory, skip ignore patterns
  ├─ 4. Parse imports per language (regex MVP)
  ├─ 5. Build graph: files → nodes, imports → links
  ├─ 6. Index file contents via Bleve
  └─ 7. Store graph in memory, return response
```

### 3.4 Key Data Structures

```go
// Node in the dependency graph
type Node struct {
    ID    string `json:"id"`    // relative file path
    Name  string `json:"name"`  // filename only
    Group int    `json:"group"` // language group for coloring
    Val   int    `json:"val"`   // node size (line count)
}

// Link between two nodes
type Link struct {
    Source string `json:"source"`
    Target string `json:"target"`
}

// Tree node for sidebar
type TreeNode struct {
    Name     string      `json:"name"`
    Path     string      `json:"path"`
    Type     string      `json:"type"` // "file" | "directory"
    Children []*TreeNode `json:"children,omitempty"`
}
```

### 3.5 Configuration (conf.yaml)

All runtime values are centralized in `conf.yaml` — zero hardcoded values:
- Server: port, host, timeouts, shutdown grace period
- Storage: base directory, index directory (supports `$HOME` expansion)
- Parser: max depth, ignore patterns
- Indexer: batch size, max file size
- CORS: allowed origins, methods, headers
- Logging: level, format

### 3.6 Graceful Shutdown

The server listens for `SIGINT`/`SIGTERM` and:
1. Stops accepting new connections
2. Drains in-flight requests (up to `shutdown_timeout`)
3. Closes Bleve indexes
4. Exits cleanly

## 4. Frontend Design

### 4.1 Component Tree

```
App
├── GraphDataProvider (context)
│   └── UIStateProvider (context)
│       └── Layout (CSS Grid: sidebar | canvas | bottom)
│           ├── RepoInput (URL input bar, top)
│           ├── LeftSidebar
│           │   └── FileTree (react-arborist)
│           ├── MainCanvas
│           │   └── ForceGraph3D (react-force-graph-3d)
│           └── BottomPanel (toggleable)
│               ├── SearchPanel
│               └── ChatPanel (Phase 2)
```

### 4.2 State Management

Two React Contexts, split by update frequency:

**GraphDataContext** (infrequent updates):
- `fullGraphData` — complete graph from API
- `displayGraphData` — filtered view based on focused node
- `focusNode(id)` / `resetView()` actions

**UIStateContext** (moderate updates):
- `bottomPanelMode` — 'search' | 'chat' | 'hidden'
- `sidebarCollapsed` — boolean
- `focusedNodeId` — synced from GraphDataContext

### 4.3 3D Graph Interaction

| Action               | Behavior                                           |
|----------------------|----------------------------------------------------|
| Click node           | Focus node, filter to show immediate neighbors only|
| Click background     | Reset to full graph view                           |
| Click file in tree   | Same as clicking the corresponding node            |
| Scroll               | Zoom in/out                                        |
| Drag                 | Rotate camera                                      |

### 4.4 Dark Theme

Enforced via Tailwind CSS:
- Background: `#0a0a0a` (near-black)
- Secondary BG: `#1a1a1a`
- Text: `#e5e5e5`
- Accent: `#3b82f6` (blue)
- Custom scrollbars, smooth transitions

## 5. Development Workflow

```
make frontend-dev   # Start Vite dev server (port 5173)
make run            # Build & run Go server (port 1983)
make frontend       # Build frontend → web/static/
make stop           # Kill running server
make test           # Run Go tests
make clean          # Remove build artifacts
```

In production mode, the Go server serves both the API and the frontend static files from `web/static/`.

During development, Vite dev server runs on port 5173 and proxies API calls to Go on port 1983.

## 6. Design Decisions

| Decision                        | Choice                  | Rationale                                 |
|---------------------------------|-------------------------|-------------------------------------------|
| HTTP framework                  | net/http (Go 1.22+)    | Method routing built-in, zero deps        |
| Dependency parsing              | Regex (MVP)             | Fast to implement, upgrade path to AST    |
| Graph storage                   | In-memory               | Simple, fast access, acceptable for MVP   |
| State management                | React Context           | Sufficient for single-page MVP            |
| Frontend bundler                | Vite                    | Fast HMR, TypeScript support              |
| Config format                   | YAML                    | Human-readable, supports env vars         |
| Search engine                   | Bleve                   | Pure Go, embedded, no external deps       |
