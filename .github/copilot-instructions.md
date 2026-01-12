# GitHub Copilot Instructions for Tinode Chat

## Project Overview
Tinode is a modern, open-source instant messaging platform (server & clients) written primarily in Go (backend) and supports various client platforms. It aims to be a replacement for XMPP.

## Codebase Structure
- **chat/**: Main server codebase (Go).
  - `server/`: Core server logic and entry point (`main.go`).
  - `tinode-db/`: Database initialization and management tool.
  - `pbx/`: Protobuf definitions (`model.proto`).
  - `rest-auth/`: Sample REST authentication server (Python).
- **pushgw/**: Push Gateway service for handling mobile push notifications (FCM, etc.).
- **py_grpc/**: Python gRPC client and library bindings.
- **exporter/**: Prometheus exporter for monitoring.
- **prod/**: Production deployment scripts, configuration, and Docker setups.

## Critical Developer Workflows

### 1. Building the Server
**Crucial**: The server **must** be built with a specific database tag. It does not support all databases simultaneously in a single binary by default.
- **Tags**: `mysql`, `postgres`, `mongodb`, `rethinkdb`.
- **Command**:
  ```bash
  # Example for PostgreSQL
  go build -tags postgres ./server
  ```
- **Scripts**:
  - `build-all.sh`: Builds binaries for all architectures and DBs.
  - `build-one.sh`: Build for a specific target.

### 2. Database Initialization
Use the `tinode-db` tool to initialize or upgrade the database schema.
- Located in `chat/tinode-db`.
- Needs the same build tags as the server.
- Usage: `./tinode-db -config=./tinode.conf -data=data.json`

### 3. Protobuf & gRPC Generation
Do **not** edit `*.pb.go` or `*_pb2.py` files manually. They are generated from `pbx/model.proto`.
- **Go**: Run `go generate` in the relevant directories or use:
  ```bash
  protoc --proto_path=../pbx --go_out=plugins=grpc:../pbx ../pbx/model.proto
  ```
- **Python**:
  ```bash
  python -m grpc_tools.protoc -I../pbx --python_out=. --grpc_python_out=. ../pbx/model.proto
  ```

### 4. Running Tests & Linting
- **Tests**: Standard Go testing.
  ```bash
  go test -tags postgres ./...
  ```
- **Linting**: Use `run-linter.sh` in the `chat` root. It uses `golangci-lint` with specific flags to disable certain noisy checks (e.g., `stylecheck`, `paralleltest`, `wsl`).

## Architecture & patterns

### Server (`chat/server`)
- **Entry Point**: `main.go` parses flags and config (`tinode.conf`), then starts the Hub.
- **Hub**: Central component managing topics and sessions (`hub.go`).
- **Plugins**: The server supports gRPC plugins for extensions (auth, macro-bot operations).
- **Handlers**:
  - `hdl_websock.go`: WebSocket handler (the primary transport).
  - `hdl_grpc.go`: gRPC handler.
  - `hdl_longpoll.go`: Fallback transport.

### Database Abstraction
The `store` package defines interfaces for DB operations. The implementations (e.g., `store/mysql`, `store/postgres`) are conditionally compiled using build tags.

### Configuration
Configuration is loaded from `tinode.conf` (JSON format). It controls listen addresses, database connections, and plugin settings.
