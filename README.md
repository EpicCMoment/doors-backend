# doors-backend

Go library that talks to a background DOORS 9.7 instance running an embedded
DXL TCP server. It is the core of the DOORS MCP server.

## Layout

- `config.go` — JSON `Config` + defaults (port 61610)
- `runner.go`, `runner_windows.go`, `runner_unix.go` — spawn `doors -batch` with the extracted DXL server script
- `doors_controller.go` — singleton `DoorsController`, created via `backend.Init(cfg)`
- `dxl/` — `DxlController` (TCP client, one conn/request, serialized), template cache, typed `DxlLiteral` injection, embedded DXL TCP server (`dxl/server/dxl_server.dxl`)
- `module/` — `ModuleController`: typed API + baseline-aware cache for modules, requirements, baselines, flat hierarchy
- `scripts/` — embedded utility DXL scripts (`ping`, `get_module`, `list_modules`, `get_requirements`, `get_baselines`, `traverse_hierarchy`), each producing a JSON envelope assigned to `return_`

## Status

Unit-tested with a fake transport; **not yet validated against a real DOORS
instance**. See the repo-root README for the shared TODO list.

## Commands

```sh
go vet ./...
go test ./...
GOOS=windows go build ./...
```
