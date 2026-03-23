# blogroll-md

`blogroll-md` is a markdown blogroll generator for Jekyll pages.

## Building

### Generating mocks

```bash
docker run -v "$PWD":/src -w /src vektra/mockery:3
```

### Docker image

```bash
docker build -t brunolebon/blogroll-md .
```

### Go

```bash
go build -o build/blogroll-md ./cmd/...
```

## Task input

A task is configured with a YAML file. Example:

```yaml
header:
  layout: page
  title: "Mauritian blogs"
  permalink: /mauritian-blogs/
sqlite:
  filename: mbl.db
  tablename: mbl
  cutoff: 1970-01-01
  update: true
yaml:
  filename: mbl.yml
  cutoff: 1970-01-01
  input: true
  update: true
output:
  filename: mauritian-blogs.md
  cutoff: 2010-01-01
  timezone: "Indian/Mauritius"
worker:
  poolSizeMultiplier: 1
```

### Fields

#### `header`

Written verbatim into the front matter of the generated Markdown file.

| Field | Description |
|---|---|
| `layout` | Jekyll layout name |
| `title` | Page title |
| `permalink` | Page permalink |

#### `sqlite`

Optional SQLite database. When both `filename` and `tablename` are set and `input: true`, the database is used as the blog list source.

| Field | Description |
|---|---|
| `filename` | Path to the `.db` file |
| `tablename` | Table to read from / write to |
| `cutoff` | Blogs with a last-post date before this date are excluded when reading |
| `input` | Set `true` to use this database as the input source |
| `update` | Set `true` to write updated post data back to the database after fetching |

#### `yaml`

Optional YAML database. When `filename` is set and `input: true`, the file is used as the blog list source.

| Field | Description |
|---|---|
| `filename` | Path to the `.yml` file |
| `cutoff` | Blogs with a last-post date before this date are excluded when reading |
| `input` | Set `true` to use this file as the input source |
| `update` | Set `true` to write updated post data back to the file after fetching |

Exactly one of `sqlite` or `yaml` must have `input: true`.

#### `output`

Controls the generated Markdown file.

| Field | Description |
|---|---|
| `filename` | Path to the output `.md` file |
| `cutoff` | Blogs with a last-post date before this date are excluded from the output |
| `timezone` | IANA timezone name used when formatting post timestamps (e.g. `"Indian/Mauritius"`) |

#### `worker`

Controls the concurrency of feed fetching.

| Field | Default | Description |
|---|---|---|
| `poolSizeMultiplier` | `4` | Worker pool size = `poolSizeMultiplier × NumCPU` |

## Running

### Docker image

```bash
docker run --rm \
  -v "$(pwd)/example:/data" \
  -w /data \
  --user "$(id -u):$(id -g)" \
  brunolebon/blogroll-md blogroll-md -input task.yml
```
