# doors-backend

Backend library that bridges Go to a DOORS 9.7 instance over its DXL TCP server.

## Non-obvious facts (would bite an agent)

- **`eval_` protocol**: `DxlController` sends a fully-rendered DXL script per request (one short-lived TCP connection). The script produces output by assigning to the implicit DOORS variable `return_` — `eval_`'s return value is NOT the payload. If `return_` is empty, the server returns a `{"status":"error",...}` envelope.
- **DXL string concat**: no `+` operator — adjacent operands concatenate (`a b` / `result = result ","`). Always assign via intermediate string for concatenated arguments in calls (`jerr(...)`, `baselineName(...)`).
- **Envelope**: every utility script ends with `return_ = result` where `result` is `{"status":"ok"|"error","data"|"message"}`. `module.decodeEnvelope` validates it.
- **Module identity is split**: `Module.ID = uniqueID(m)` (DOORS unique ID), `ModulePath = fullName(m)` (what `open()` accepts). Cache module keys and `moduleId`-style params are paths; renaming `moduleId -> modulePath` was deliberate — keep it when touching new code.
- **Baseline dimension**: all content structs carry `BaselineID`; `""` means the current view. Cache keys are baseline-aware (`path@current`, `modulePath/reqID@baseline`). Reading a baseline uses `load(m, b, false)` → returns a `Module` to iterate, never the older two-arg form.
- **Baseline name**: preamble helper `baselineName(b) = "<major>.<minor><suffix> (<annotation>)"` for both id and name.
- **Template cache lives in `dxl.DxlController`** (`TemplateCache`), seeded in `DoorsController.Init` with `scripts.Load(name)`. Preamble (JSON helpers: `jsonEscape`, `jstr`, `intToStr`, `jerr`, `baselineName`) is prepended by `scripts.Load`.
- **DoorsController is a singleton** built by `backend.Init(cfg)`; failed init keeps `Default()` returning an error.
- **Runner** spawns `doors -batch <extracted dxl_server.dxl>` and polls the port; embedded server script lives at `dxl/server/dxl_server.dxl`, port injected via `RenderServerScript`.
- **Syntax correctness of `scripts/*.dxl` is only partially validated**: no DOORS binary on dev machines. Smoke-test with `ping.dxl` first; object-model builtins (`open`, `o."Object Number"`, `o."Object Heading"`, `o."Level"`, `load`, `Baseline b in m`) were written against the 9.7 reference and may need small tweaks on a real host.

## Commands

```sh
go vet ./...
go test ./...
GOOS=windows go build ./...   # runner_windows.go path
```
