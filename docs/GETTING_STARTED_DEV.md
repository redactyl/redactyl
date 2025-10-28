# Getting Started with Redactyl Development

**For developers working on the Gitleaks integration and artifact scanning pivot**

---

## Quick Context

**Read these first:**
1. `/CLAUDE.md` - Strategic context, decisions, and project direction
2. `/ROADMAP.md` - Product roadmap with quarterly milestones
3. `/docs/IMPLEMENTATION_PLAN.md` - Detailed technical implementation plan

**TL;DR:** We're pivoting from custom secret detection to being a specialized deep artifact scanner powered by Gitleaks. Focus: container images, Helm charts, K8s manifests, and complex nested artifacts.

---

## Current State (Before You Start)

### What Works Today
- ✅ Basic secret scanning with 80+ custom detectors
- ✅ Archive streaming (zip, tar, tgz) without disk extraction
- ✅ Container image scanning (Docker save format)
- ✅ Virtual paths for nested artifacts
- ✅ SARIF output, JSON output, remediation commands

### What We're Changing
- 🔄 Replacing custom detectors with Gitleaks integration
- 🔄 Adding Helm chart scanning
- 🔄 Adding Kubernetes manifest detection
- 🔄 Enhancing container scanning (OCI format, layer context)

### What We're Keeping
- ✅ Artifact streaming engine (`internal/artifacts/`)
- ✅ Virtual path system
- ✅ Remediation suite (`fix`, `purge` commands)
- ✅ SARIF/JSON output formats

---

## Development Setup

### Prerequisites
```bash
# Go 1.25+
go version

# Gitleaks binary (for testing integration)
brew install gitleaks  # macOS
# or download from https://github.com/gitleaks/gitleaks/releases

# Make
make --version
```

### Clone and Build
```bash
git clone https://github.com/redactyl/redactyl.git
cd redactyl

# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Build and run
./bin/redactyl scan --help
```

### Project Structure
```
redactyl/
├── cmd/redactyl/          # CLI commands
│   ├── scan.go           # Main scan command
│   ├── fix.go            # Remediation commands
│   └── purge.go          # History rewriting
├── internal/
│   ├── artifacts/        # 🌟 Core artifact streaming (keep & enhance)
│   │   ├── artifacts.go  # Archive/container scanning
│   │   └── ...
│   ├── detectors/        # ❌ Custom detectors (will be removed)
│   ├── scanner/          # ✨ NEW: Scanner interface & Gitleaks integration
│   ├── engine/           # Scan orchestration
│   ├── config/           # Configuration management
│   └── types/            # Shared types
├── pkg/core/             # Public Go API
├── docs/                 # Documentation
│   ├── IMPLEMENTATION_PLAN.md  # Detailed dev plan
│   └── deep-scanning.md        # Artifact scanning guide
├── CLAUDE.md             # 🎯 Project context (read this!)
└── ROADMAP.md            # Product roadmap
```

---

## Current Development Focus: Gitleaks Integration

**Goal:** Replace custom detectors with Gitleaks binary integration (12 weeks)

### Phase 1: Core Integration (Weeks 1-4) - **START HERE**

#### Week 1: Scanner Interface
**Branch:** `feature/gitleaks-integration`

**Tasks:**
1. Create `internal/scanner/scanner.go` interface
2. Create `internal/scanner/gitleaks/binary.go` (binary detection/download)
3. Update `internal/config/config.go` to add Gitleaks config section
4. Write unit tests

**How to start:**
```bash
git checkout -b feature/gitleaks-integration

# Create scanner interface
mkdir -p internal/scanner/gitleaks
touch internal/scanner/scanner.go
touch internal/scanner/gitleaks/binary.go
touch internal/scanner/gitleaks/scanner.go

# Run tests as you go
go test ./internal/scanner/...
```

**Reference implementation:** See `/docs/IMPLEMENTATION_PLAN.md` Phase 1, Week 1

#### Week 2: Gitleaks Scanner Implementation
**Tasks:**
1. Implement `internal/scanner/gitleaks/scanner.go`
2. Write tests with real Gitleaks binary
3. Add virtual path remapping logic
4. Test JSON parsing from Gitleaks output

**Testing:**
```bash
# Create test file with known secret
echo 'aws_access_key_id = AKIAIOSFODNN7EXAMPLE' > /tmp/test-secret.txt

# Test Gitleaks directly
gitleaks detect --no-git --source /tmp/test-secret.txt

# Test your scanner implementation
go test ./internal/scanner/gitleaks -v -run TestScanWithGitleaks
```

#### Week 3: Engine Integration
**Tasks:**
1. Update `internal/engine/engine.go` to use scanner interface
2. Update `internal/artifacts/artifacts.go` to pass scanner to functions
3. Replace all `detectors.RunAll()` calls with `scanner.Scan()`
4. Update tests

**Before:**
```go
findings := detectors.RunAll(path, data)
```

**After:**
```go
findings, err := e.scanner.Scan(path, data)
```

#### Week 4: CLI Updates
**Tasks:**
1. Remove `--enable`, `--disable`, `--min-confidence` flags
2. Remove `cmd/redactyl/detectors.go` (detectors command)
3. Update help text and examples
4. Write migration guide

---

## Testing Guidelines

### Unit Tests
```bash
# Test specific package
go test ./internal/scanner/gitleaks -v

# Test with coverage
go test ./internal/scanner/... -cover

# Run all tests
make test
```

### Integration Tests
```bash
# E2E CLI tests (requires built binary)
make build
go test ./cmd/redactyl -v -tags=integration

# Test with real artifacts
./bin/redactyl scan --archives testdata/sample.zip
./bin/redactyl scan --containers testdata/sample.tar
```

### Manual Testing Checklist
Create this test suite as you develop:

```bash
# 1. Basic scan
echo 'GITHUB_TOKEN=ghp_1234567890' > /tmp/test.txt
./bin/redactyl scan /tmp/test.txt

# 2. Archive scanning
zip /tmp/test.zip /tmp/test.txt
./bin/redactyl scan --archives /tmp/test.zip

# 3. Container scanning
# (requires Docker)
docker pull alpine:latest
docker save alpine:latest > /tmp/alpine.tar
./bin/redactyl scan --containers /tmp/alpine.tar

# 4. Nested archives
# Create zip inside tar
./bin/redactyl scan --archives /tmp/nested.tar

# 5. JSON output
./bin/redactyl scan /tmp/test.txt --json

# 6. SARIF output
./bin/redactyl scan /tmp/test.txt --sarif
```

---

## Common Development Tasks

### Adding a New Feature
1. Check if it aligns with `/ROADMAP.md` priorities
2. Create feature branch: `feature/short-description`
3. Update `/CLAUDE.md` if it affects strategy
4. Write tests first (TDD preferred)
5. Implement feature
6. Update documentation
7. Open PR with clear description

### Debugging
```bash
# Enable verbose logging (TODO: add this flag)
./bin/redactyl scan --verbose /tmp/test.txt

# Use delve debugger
dlv debug . -- scan /tmp/test.txt

# Print intermediate values
# Add log.Printf() statements liberally during development
```

### Performance Profiling
```bash
# CPU profile
go test ./internal/artifacts -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profile
go test ./internal/artifacts -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Benchmark specific function
go test ./internal/scanner/gitleaks -bench=BenchmarkScan -benchmem
```

---

## Code Style Guidelines

### Go Best Practices
- Follow standard Go formatting (`gofmt`, `goimports`)
- Use meaningful variable names (no single-letter except loops)
- Keep functions small (< 50 lines ideal)
- Document all exported functions and types
- Handle errors explicitly (no `_` unless justified)

### Project-Specific Conventions

**Virtual Paths:**
```go
// Always use scanner.BuildVirtualPath()
virtualPath := scanner.BuildVirtualPath("archive.zip", "inner.tar", "file.txt")
// Result: "archive.zip::inner.tar::file.txt"
```

**Error Handling:**
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to scan %s: %w", path, err)
}
```

**Testing:**
```go
// Use testify for assertions
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// require.* for fatal errors
require.NoError(t, err)

// assert.* for non-fatal assertions
assert.Equal(t, expected, actual)
```

---

## Documentation Standards

### When to Update Docs
- Adding a new CLI flag → update command help + README
- Adding a config option → update `docs/configuration.md`
- Changing behavior → update relevant doc + add migration note
- New feature → update README, add to CHANGELOG.md

### Doc Locations
- User-facing: `README.md`, `docs/*.md`
- Developer-facing: `CLAUDE.md`, `CONTRIBUTING.md`, inline code comments
- API: `pkg/core/` (godoc comments)

---

## Git Workflow

### Branch Strategy
- `main` - stable, always buildable
- `feature/*` - new features
- `fix/*` - bug fixes
- `docs/*` - documentation only

### Commit Messages
```
feat(scanner): add Gitleaks binary detection

- Implement BinaryManager for finding gitleaks binary
- Add auto-download from GitHub releases
- Add version checking
- Closes #123
```

Format: `<type>(<scope>): <subject>`

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`

### Pull Request Template
```markdown
## Description
Brief description of changes

## Related Issues
Closes #123

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Documentation
- [ ] README updated (if needed)
- [ ] CHANGELOG updated
- [ ] Inline docs added

## Checklist
- [ ] Code follows style guidelines
- [ ] No new linter warnings
- [ ] Backward compatible (or migration guide provided)
```

---

## Troubleshooting

### Gitleaks Binary Not Found
```bash
# Check if gitleaks is in PATH
which gitleaks

# Install locally for testing
brew install gitleaks  # macOS
# or download from GitHub releases

# For auto-download testing, remove cached binary
rm -rf ~/.redactyl/bin/gitleaks
```

### Tests Failing
```bash
# Run specific test with verbose output
go test ./internal/scanner/gitleaks -v -run TestSpecificTest

# Run tests without cache
go clean -testcache
go test ./...

# Check for race conditions
go test -race ./...
```

### Build Errors
```bash
# Clean and rebuild
make clean
make build

# Update dependencies
go mod tidy
go mod verify

# Check Go version
go version  # Should be 1.25+
```

---

## Resources

### Internal Docs
- `/CLAUDE.md` - Project strategy and context
- `/ROADMAP.md` - Product roadmap
- `/docs/IMPLEMENTATION_PLAN.md` - Technical implementation details
- `/docs/deep-scanning.md` - Artifact scanning guide

### External References
- [Gitleaks Documentation](https://github.com/gitleaks/gitleaks)
- [SARIF Spec](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)
- [OCI Image Spec](https://github.com/opencontainers/image-spec)
- [Helm Chart Structure](https://helm.sh/docs/topics/charts/)

### Community
- GitHub Discussions: (TODO: enable)
- Discord: (TODO: create)
- Email: (TODO: set up)

---

## Next Steps

1. **Read strategic context:** `/CLAUDE.md`
2. **Understand the plan:** `/ROADMAP.md` and `/docs/IMPLEMENTATION_PLAN.md`
3. **Set up environment:** Install Go, Gitleaks, build project
4. **Start coding:** Begin with Phase 1, Week 1 tasks
5. **Ask questions:** Open GitHub discussion or check existing docs

**Current Priority:** Gitleaks integration (Phase 1, Weeks 1-4)

**Ready to start?**
```bash
git checkout -b feature/gitleaks-integration
mkdir -p internal/scanner/gitleaks
# Start with scanner interface (see IMPLEMENTATION_PLAN.md)
```

---

**Welcome to the team! Let's build the best artifact scanner for cloud-native environments.**
