# Audit Logging

Redactyl automatically maintains an audit log of all scans for compliance and reporting purposes.

Audit logs are **redacted by default** (no raw match/secret values). To opt in to storing raw values, toggle `R` in the TUI and re-run scans.

## Log Location

- `.git/redactyl_audit.jsonl` (if in a Git repository)
- `.redactyl_audit.jsonl` (otherwise)

## Log Format

JSON Lines format (one JSON object per line) for easy parsing:

```json
{
  "timestamp": "2025-11-24T15:41:43.103433-07:00",
  "scan_id": "scan_1764024103",
  "root": "/Users/user/project",
  "total_findings": 130,
  "new_findings": 12,
  "baselined_count": 118,
  "severity_counts": {
    "high": 116,
    "medium": 13,
    "low": 1
  },
  "files_scanned": 163,
  "duration": "302ms",
  "baseline_file": "redactyl.baseline.json",
  "top_findings": [
    {
      "path": "cmd/app/main.go",
      "detector": "generic-api-key",
      "severity": "high",
      "line": 42
    }
  ]
}
```

## Usage Examples

View all audit logs:

```sh
cat .git/redactyl_audit.jsonl | jq .
```

Count total scans:

```sh
wc -l .git/redactyl_audit.jsonl
```

Filter scans with high-severity findings:

```sh
cat .git/redactyl_audit.jsonl | jq 'select(.severity_counts.high > 0)'
```

Export for compliance report:

```sh
cp .git/redactyl_audit.jsonl audit_trail_$(date +%Y%m%d).jsonl
```

Generate summary report:

```sh
cat .git/redactyl_audit.jsonl | jq -r '[.timestamp, .total_findings, .new_findings] | @csv'
```

## Benefits

- **Immutable trail** - Append-only log ensures scan history is preserved
- **Compliance ready** - Structured format suitable for SOC2, ISO 27001, and other audits
- **Timestamped** - Every scan recorded with precise timestamp
- **Severity tracking** - Monitor high/medium/low findings over time
- **Baseline tracking** - Shows which findings are accepted vs new
- **Sample findings** - Top 10 new findings included for quick reference
- **Performance metrics** - Duration and files scanned tracked
