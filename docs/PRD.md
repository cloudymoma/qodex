Product Requirements Document (PRD): Qodex
1. Overview
Qodex is a web-based, interactive 3D codebase visualizer and search tool. It allows developers to input a public GitHub repository, automatically parses the code structure, and renders a 3D dependency graph alongside an advanced search and AI chat interface.

2. Goals & Scope (MVP)
Target: Google gHack Submission (Solo Project).

Goal: Demonstrate a "fancy", highly visual, and disruptive way to explore and understand codebases.

Scope: Public GitHub repositories only. No user authentication. Plain HTTP local deployment. Focus on visual impact and core search functionality, with scaffolding for local LLM integration.

3. User Interface Layout (Dark Theme)
Main Canvas (Center/Right): 3D WebGL Graph. Nodes represent files; edges represent import/usage dependencies. Supports mouse drag to rotate, and scroll to zoom.

Left Sidebar: Traditional hierarchical file explorer tree.

Bottom Panel: A toggleable dock containing:

Mode Switch: [Search] | [Chat (Phase 2)]

Search Mode: Text input for traditional keyword search, displaying highlighted results from the inverted index.

Chat Mode: Interface to communicate with a local Ollama instance (UI included for MVP, backend integration deferred if time is tight).

4. Core User Flows
Ingestion: User enters a GitHub URL and optional branch. System clones it to $HOME/.qodex/<repo_name> and builds an inverted index and dependency graph.

3D Exploration:

Default view: Full file dependency graph. Orphans float independently.

Clicking a Node: Focuses the node, highlights it, and filters the graph to show only its immediate (1st level) connected nodes.

Expanding: Click an "expand" action on a 1st-level node to reveal its immediate children (1 level at a time).

Deselect: Clicking empty space resets to the default global view.

Tree Navigation: Clicking a file in the left sidebar triggers the same focus/highlight behavior on the 3D graph as clicking a node directly.

Code Search: User types keywords (function names, variables). System queries the inverted index and returns matching files and code snippets with highlights.

Technical Design Document
1. Architecture Stack
Backend: Golang (Standard net/http or lightweight framework like Fiber/Gin).

Frontend: React (via Vite for fast compilation) + Tailwind CSS (for native Dark Mode and sleek UI).

3D Rendering: react-force-graph-3d (A powerful wrapper around Three.js and d3-force-3d, perfect for this exact use case).

Search Index: bleve (A robust, pure Go text indexing library).

LLM Integration: Ollama (running locally, exposed via REST API).

2. Backend Implementation Details (Golang)
2.1. Repository Management
Use the go-git/go-git library to clone public repositories in-memory or directly to the file system ($HOME/.qodex/...).

Endpoint: POST /api/repo/load -> Payload: { "url": "...", "branch": "main" }

2.2. Parsing & Indexing
Text Indexing (bleve): Traverse the downloaded directory. Index file paths, file names, and raw file contents. bleve will automatically tokenize and allow for fast searching of functions, variables, and comments.

Dependency Parsing (The tricky part for MVP):

Recommendation: To keep it simple for the hackathon, restrict the MVP dependency parsing to a specific language you know well (like Go or Rust), using regex to find import statements.

Advanced (if time permits): Use Go bindings for tree-sitter to parse the Abstract Syntax Tree (AST) of multiple languages to accurately map file-to-file dependencies.

2.3. API Endpoints
POST /api/ingest: Triggers clone and indexing. Returns success status.

GET /api/graph: Returns the JSON structure required by the frontend 3D graph.

Format: { "nodes": [{"id": "fileA", "group": 1}], "links": [{"source": "fileA", "target": "fileB"}] }

GET /api/tree: Returns the hierarchical directory structure for the left sidebar.

GET /api/search?q=keyword: Queries bleve and returns an array of matched files, line numbers, and text snippets.

3. Frontend Implementation Details (React)
3.1. UI & Theming
Use Tailwind CSS to enforce a strict dark theme (bg-gray-900, text-gray-100).

Use shadcn/ui or Radix UI for unstyled, accessible components (dropdowns, inputs, scroll areas) that you can easily make look "hacker-chic".

3.2. The 3D Graph (react-force-graph-3d)
Map the /api/graph payload directly to the <ForceGraph3D graphData={data} /> component.

Interaction Logic: Maintain a React state focusedNode.

onNodeClick: Update focusedNode. Filter the graphData passed to the component to only include the clicked node and links where source == focusedNode or target == focusedNode.

onBackgroundClick: Clear focusedNode state, restoring the full graph.

3.3. State Management
Since it's a single-page MVP, standard React useState and useContext will suffice to synchronize the left Tree Explorer and the center 3D Graph (e.g., clicking the tree updates the focusedNode state of the graph).

4. Execution Plan (Hackathon Timeline)
Hour 1-2: Bootstrap React frontend with dark theme and placeholder 3D graph. Bootstrap Go backend.

Hour 3-5: Implement GitHub cloning and basic file directory traversal in Go. Wire up the left sidebar tree.

Hour 5-8: Implement simple Regex/AST parsing to find dependencies. Feed JSON to the 3D graph. Implement node-click filtering logic.

Hour 8-10: Implement bleve indexing and wire up the bottom Search UI.

Hour 10+: Polish UI, add the "Chat" toggle UI (even if it just sends a hardcoded mock response to/from Ollama for the demo).

Extra info:
User a `Makefile` to control the build and run, run on local port 10080 http, make a stop target which will ps and grep the process ID of the running app then kill it. let the server side app can handle graceful shutdown. centralize all configurations into a conf.yaml file, no hardcode thing.