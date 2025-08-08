## Redactyl

Find secrets in your repo with low noise. Redactyl scans your working tree, staged changes, diffs, or history and reports likely credentials and tokens.

### Features
- Fast multi-threaded scanning with size limits and binary detection
- Worktree, staged, history (last N commits), or diff vs base branch
- Detector enable/disable controls
- Baseline file to suppress known findings
- Outputs: table (default), JSON, SARIF 2.1.0
- Gitignore-style path ignores via `.redactylignore`
- CI-friendly exit codes

### Install
- Build locally (repo root):
  ```sh
  make build
  ./bin/redactyl --help
  ```
- Or:
  ```sh
  go build -o bin/redactyl .
  go install .  # installs to $(go env GOBIN) or $(go env GOPATH)/bin
  ```

### Quick start
- Default scan:
  ```sh
  ./bin/redactyl scan
  ```
- JSON:
  ```sh
  ./bin/redactyl scan --json
  ```
- SARIF:
  ```sh
  ./bin/redactyl scan --sarif > redactyl.sarif.json
  ```
- Staged changes only:
  ```sh
  ./bin/redactyl scan --staged
  ```
- Last N commits:
  ```sh
  ./bin/redactyl scan --history 5
  ```
- Diff vs base branch:
  ```sh
  ./bin/redactyl scan --base main
  ```
- Performance:
  ```sh
  ./bin/redactyl scan --threads 4 --max-bytes 2097152
  ```

### Baseline
- Update the baseline from the current scan results:
  ```sh
  ./bin/redactyl baseline update
  ```
- Baseline file: `redactyl.baseline.json`
- The baseline suppresses previously recorded findings; only new findings are reported.

### Ignoring paths
- Create `.redactylignore` at your repo root (gitignore syntax). Example:
  ```
  # node artifacts
  node_modules/
  dist/
  *.min.js

  # test data
  testdata/**
  ```
- Paths matching this file are skipped.

### Detectors
- List available IDs:
  ```sh
  ./bin/redactyl detectors
  ```
- Enable only specific detectors:
  ```sh
  ./bin/redactyl scan --enable "twilio,github_token"
  ```
- Disable specific detectors:
  ```sh
  ./bin/redactyl scan --disable "entropy_context"
  ```
- Current detector IDs:
  - aws_access_key
  - aws_secret_key
  - github_token
  - slack_token
  - jwt
  - private_key_block
  - entropy_context
  - stripe_secret
  - twilio_account_sid
  - twilio_api_key_sid
  - twilio_auth_token
  - twilio_api_key_secret_like

### Output formats
- Table (default): human-friendly summary
- JSON: machine-readable; never returns null array
- SARIF 2.1.0: for code scanning dashboards

### CLI reference
- `./bin/redactyl --help`
- `./bin/redactyl scan --help`

### Common scan flags
- **--path, -p**: path to scan (default: .)
- **--staged**: scan staged changes
- **--history N**: scan last N commits
- **--base BRANCH**: scan diff vs base branch
- **--include / --exclude**: comma-separated globs
- **--max-bytes**: skip files larger than this (default: 1 MiB)
- **--threads**: worker count (default: GOMAXPROCS)
- **--enable / --disable**: comma-separated detector IDs
- **--json / --sarif**: select output format
- **--fail-on**: low | medium | high (default: medium)

### Exit codes
- 0: no findings or below threshold
- 1: findings at or above threshold (see `--fail-on`)
- 2: error while scanning

### CI usage (GitHub Actions)
```yaml
name: Redactyl Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - run: go build -o bin/redactyl .
      - run: ./bin/redactyl scan --sarif > redactyl.sarif.json
```

### Notes
- Redactyl respects `.redactylignore` for path filtering.
- Findings are deduplicated by `(path|detector|match)`.
- Baseline suppresses previously seen findings; update your baseline after intentionally introduced secrets are handled (for example, false positives).

### Version
```sh
./bin/redactyl version
```

License, contribution guidelines, and detailed examples can be added here if needed.