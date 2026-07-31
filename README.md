<div align="center">

# dataflute

**Expose local data files as read-only REST APIs.**

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-4CAF50?style=flat-square)](LICENSE)
[![Router: Chi](https://img.shields.io/badge/Router-Chi-2ECC71?style=flat-square)](https://github.com/go-chi/chi)
[![CLI: Cobra](https://img.shields.io/badge/CLI-Cobra-2C3E50?style=flat-square)](https://github.com/spf13/cobra)
[![Watcher: fsnotify](https://img.shields.io/badge/Watcher-fsnotify-2980B9?style=flat-square)](https://github.com/fsnotify/fsnotify)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat-square)]()

</div>

`dataflute` is a production-ready, read-only REST API server for local data files.
Point it at a directory (or a single file) and it exposes each supported dataset as a
JSON endpoint with filtering, sorting, pagination, field selection, and search — with
automatic hot-reload when files change.

Built with Go, Cobra, Chi, fsnotify, and slog. Follows Clean Architecture and is
designed for extensibility: new file formats are added by implementing a parser
interface — no HTTP server changes required.

## Features

- **Two serving modes** — serve a whole directory of datasets or a single file.
- **Format-agnostic** — datasets are always returned as JSON regardless of source format.
- **Seven formats built in** — JSON, CSV, TSV, XML, YAML, TOML, and INI, auto-selected by file extension.
- **Hot reload** — `fsnotify`-based watcher reloads only changed files; no restart needed.
- **Query features** — filter, sort, search, field selection, and pagination on array datasets.
- **Read-only** — `POST`/`PUT`/`PATCH`/`DELETE` return `405`.
- **Built-in endpoints** — `GET /` (metadata), `GET /health`, `GET /metrics`.
- **Middleware** — recovery, request logging, CORS, timeout, gzip compression, graceful shutdown.
- **Thread-safe in-memory storage** — `RWMutex`-protected, datasets never re-parsed per request.
- **Extensible parser registry** — add Excel, SQLite, Parquet, etc. by implementing one interface.

## Install & Build

Requires Go 1.26+.

```bash
make build          # builds bin/dataflute
make test           # runs all tests with coverage
make lint           # go vet
make fmt            # gofmt
make install        # go install
make docker         # docker build
```

Or directly:

```bash
go build -o dataflute ./cmd/dataflute
```

## Usage

### Directory mode

Serve every supported file in a directory. Each file becomes an endpoint named after
its filename (without extension).

```bash
dataflute serve -d ./data
```

Given:

```
data/
  users.json
  products.json
  employees.csv
```

You get:

```
GET /users
GET /products
GET /employees
```

### Single file mode

Serve one file under a user-defined endpoint.

```bash
dataflute serve -f ./users.json -e /employees
```

### Flags

| Flag            | Short | Description                                           | Default    |
| --------------- | ----- | ----------------------------------------------------- | ---------- |
| `--data`        | `-d`  | Directory containing supported data files             |            |
| `--file`        | `-f`  | Serve a single data file                              |            |
| `--endpoint`    | `-e`  | Endpoint for the single file (required with `--file`) |            |
| `--host`        | `-H`  | Host to bind the HTTP server to                       | `0.0.0.0`  |
| `--port`        | `-p`  | Port to bind the HTTP server to                       | `8080`     |
| `--pretty`      | `-P`  | Pretty-print JSON responses                           | `false`    |

Exactly one of `--data` or `--file` is required; they cannot be combined. Invalid
configurations produce clear CLI errors.

## API

All responses are JSON. Content-Type is always `application/json`, and responses are
gzip-compressed when the client sends `Accept-Encoding: gzip`.

### `GET /`

Returns API metadata.

```json
{
  "mode": "directory",
  "formats": [".csv", ".ini", ".json", ".toml", ".tsv", ".xml", ".yaml", ".yml"],
  "available": ["/app", "/config", "/countries", "/employees", "/mapping", "/products", "/report", "/settings", "/users"]
}
```

### `GET /health`

```json
{ "status": "ok" }
```

### `GET /metrics`

```json
{
  "uptime": "1h2m3.5s",
  "requests": 103,
  "reloads": 7,
  "datasets": 3
}
```

### `GET /{dataset}`

Returns the dataset as JSON. The root value may be an array, object, or a primitive
(string, number, boolean, null).

### Read-only

Only `GET` (and `OPTIONS` preflight) is supported. `POST`, `PUT`, `PATCH`, and `DELETE`
return `405 Method Not Allowed` with an `Allow` header.

## Query Features

When the dataset root is an array, the following query parameters apply. All can be
combined.

### Filtering — exact match on top-level fields

```
GET /users?name=John
GET /users?country=PH&active=true
```

### Sorting

```
GET /users?sort=name          # ascending
GET /users?sort=-created_at   # descending
GET /users?sort=name,-id      # multiple keys
```

### Pagination

```
GET /users?page=1&limit=20
```

### Field selection

```
GET /users?fields=id,name,email
```

### Search — case-insensitive substring over string fields

```
GET /users?q=john
```

### Errors

| Status | Body                                   |
| ------ | -------------------------------------- |
| `400`  | `{"error":"invalid query"}`            |
| `404`  | `{"error":"dataset not found"}`        |
| `405`  | `{"error":"method not allowed"}`       |
| `500`  | `{"error":"internal server error"}`    |
| `504`  | `{"error":"request timeout"}`          |

## Supported Formats

The parser is selected automatically by file extension:

| Format | Extensions        | Notes                                            |
| ------ | ----------------- | ------------------------------------------------ |
| JSON   | `.json`           | Numbers preserved as-is via `UseNumber`          |
| CSV    | `.csv`            | First row is the header; empty cells become null |
| TSV    | `.tsv`            | Same as CSV with a tab delimiter                 |
| XML    | `.xml`            | Repeated child elements become arrays of records |
| YAML   | `.yaml`, `.yml`   |                                                  |
| TOML   | `.toml`           |                                                  |
| INI    | `.ini`            | Sections become nested objects                   |

The architecture makes adding formats trivial: implement `Parser` and register it.

```go
type Parser interface {
    Extensions() []string // e.g. [".json"]
    Parse(path string) (any, error)
}
```

Registering a new format requires no changes to the HTTP server, storage, watcher, or
handlers:

```go
registry := parser.NewRegistry()
registry.Register(json.New())
registry.Register(csv.New())
registry.Register(xml.New())
registry.Register(yaml.New())
registry.Register(toml.New())
registry.Register(ini.New())
```

### Examples

CSV rows become objects keyed by header, with typed values:

```
GET /employees?department=Engineering&sort=-salary
```

```json
[
  { "active": true, "department": "Engineering", "id": 3, "name": "Carol", "salary": 110000 },
  { "active": true, "department": "Engineering", "id": 1, "name": "Alice", "salary": 95000 }
]
```

XML files with a single repeated child element become arrays of records:

```
GET /report
```

```json
[
  { "amount": 12000.5, "id": 1, "name": "Q1 Sales" },
  { "amount": 9800, "id": 2, "name": "Q2 Sales" }
]
```

YAML, TOML, and INI files become objects:

```
GET /mapping
```

```json
{
  "features": ["filter", "sort", "search", "pagination"],
  "limits": { "limit": 50, "page": 100 },
  "service": "dataflute",
  "version": "1.0.0"
}
```

Planned formats that slot in this way: Excel (`.xlsx`), SQLite (`.db`, `.sqlite`),
Parquet (`.parquet`), Feather (`.feather`). Parser-specific dependencies are only
allowed when implementing those formats.

## File Watching

The server watches the data source with `fsnotify`:

- **Directory mode** — watches the directory; added/modified files are loaded or
  reloaded, removed files are dropped.
- **Single file mode** — watches the file's parent directory and filters events to the
  file itself (robust against atomic replacement).

Events are debounced to batch bursts, and only changed files are re-parsed.

## Architecture

```
cmd/dataflute/          CLI entrypoint
  main.go
  serve/                serve command + application wiring (run.go)
internal/
  api/
    handlers/           HTTP handlers (dataset, metadata, health, metrics)
    response/           shared JSON response helpers
    router/             chi router + middleware stack
  config/               runtime configuration & validation
  middleware/           recovery, request logging, CORS, timeout, gzip
  loader/               loads files into datasets via the parser registry
  watcher/              fsnotify-based hot reload
  service/              application layer: datasets, reloads, metrics
  models/               Dataset model
  storage/              thread-safe in-memory dataset store (RWMutex)
  metrics/              runtime counters for /metrics
  utils/                query processing (filter, sort, search, fields, page)
  parser/               Parser interface + Registry
    scalar/             scalar value coercion for tabular formats
    json/               JSON parser
    csv/                CSV parser
    tsv/                TSV parser
    xml/                XML parser
    yaml/               YAML parser
    toml/               TOML parser
    ini/                INI parser
pkg/data/               sample datasets
tests/                  end-to-end integration tests
```

### Dependency flow

```
cmd/serve → service ─→ loader ─→ parser.Registry ─→ parser(json, csv, tsv, xml, yaml, toml, ini)
                   └→ storage
                   └→ watcher
                   └→ metrics
api/router → api/handlers → service
```

Each layer depends only on the layer beneath it, uses constructor-based dependency
injection, and keeps all state in injected instances (no global mutable state).

## Docker

Multi-stage build, exposes port `8080`, distroless runtime.

```bash
docker build -t dataflute .

# Directory mode (mount a directory of datasets)
docker run --rm -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  dataflute serve --data /data

# Single file mode (mount one file)
docker run --rm -p 8080:8080 \
  -v "$(pwd)/users.json:/users.json" \
  dataflute serve --file /users.json --endpoint /users
```

## Development

```bash
make run                 # serve ./pkg/data on :8080
curl http://localhost:8080/
curl "http://localhost:8080/users?country=PH&sort=-id&limit=5"
curl "http://localhost:8080/employees?department=Engineering&sort=-salary"
```

### Testing

`go test ./...` runs unit and integration tests across all packages (core packages have
90–100% coverage). File-watcher tests exercise real `fsnotify` events on temp
directories.

## License

MIT
