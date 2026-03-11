<img src="misc/qodex_txt.png" height="80" />

[![Build](https://github.com/cloudymoma/qodex/actions/workflows/go.yml/badge.svg)](https://github.com/cloudymoma/qodex/actions/workflows/go.yml)

Interactive codebase visualizer. Paste a public GitHub URL, explore its dependency graph in 3D/2D, search code, browse files, and time-travel through git history — all in a dark-themed web UI.

![Qodex Screenshot](misc/qodex_screenshot.png)

## Quick Start

```bash
make build-all   # build frontend + backend
make run          # start server at http://localhost:1983
```

Then open http://localhost:1983, paste a GitHub repo URL, and click **Explore**.

## Prerequisites

- Go 1.23+
- Node.js 20+ / npm
- Make

## Makefile Targets

| Target | Description |
|---|---|
| `make build-all` | Build frontend + Go backend (one step) |
| `make run` | Build and run the server |
| `make run-accesscode` | Run with access code protection |
| `make stop` | Stop the running server |
| `make test` | Run Go tests with race detection |
| `make frontend-dev` | Start Vite dev server (HMR on :5173) |
| `make clean` | Remove build artifacts |
| `make help` | Show all targets |

## Configuration

All settings in `conf.yaml` — port, storage paths, parser rules, CORS, logging. Defaults work out of the box.

## Access Code Protection

Optionally protect your Qodex instance with a simple access code:

```bash
make run-accesscode
```

- **First visit**: You'll be prompted to set an access code. It is hashed with bcrypt and stored in `.accesscode`.
- **Subsequent visits**: Enter the access code to continue. A session cookie is set so you only authenticate once.
- **Session timeout**: Sessions expire after 30 minutes of inactivity. User activity (mouse, keyboard, scroll) automatically extends the session.
- **In-memory sessions**: Sessions are not persisted to disk. Restarting the server invalidates all sessions.
- **Immutable**: Once set, the access code cannot be changed through the UI. Delete `.accesscode` from the server to reset.
- **Optional**: Without `--accesscode`, the app runs with no authentication (default).
