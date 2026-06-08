# nano-dataset-manager

A minimal Go-based dataset intake service with disk storage, SQLite metadata, zip upload support, and Hugging Face download support using user access tokens.

## Quick Start

```bash
go build -o nano-dataset-manager .
./nano-dataset-manager
```

Server starts on `:8080` by default.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DATASETS_DIR` | `./data/datasets` | Root directory for dataset storage |
| `DATABASE_PATH` | `./data/metadata.db` | SQLite database file path |

## API

### Upload a dataset (zip file)

```bash
curl -X POST http://localhost:8080/datasets/upload \
  -F "file=@dataset.zip" \
  -F "name=my-dataset"
```

### Download from Hugging Face

```bash
curl -X POST http://localhost:8080/datasets/from-hf \
  -H "Content-Type: application/json" \
  -d '{"repo_id":"username/dataset-name","token":"hf_...","name":"my-dataset"}'
```

The token is used only for the download and is never stored.

### List all datasets

```bash
curl http://localhost:8080/datasets
```

### Get a dataset by ID

```bash
curl http://localhost:8080/datasets/<id>
```

### Delete a dataset

```bash
curl -X DELETE http://localhost:8080/datasets/<id>
```

### Health check

```bash
curl http://localhost:8080/health
```

## Storage Layout

```
data/datasets/
└── <uuid>/
    ├── raw.zip       # original upload (upload source only)
    └── data/         # extracted / downloaded files
```

SQLite stores metadata only — IDs, names, sources, paths, sizes, statuses, and timestamps. Dataset contents live on disk.

## Dependencies

Only external dependency is `modernc.org/sqlite` (pure Go SQLite driver). Everything else uses the Go standard library (net/http, archive/zip, encoding/json, crypto/rand).
