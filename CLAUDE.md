# Redactyl - Project Context for Claude

**Last Updated:** 2025-10-27
**Status:** Pre-launch, pivoting to deep artifact scanning specialization

## Executive Summary

Redactyl is pivoting from a general-purpose secret scanner to a **specialized deep artifact scanner** that focuses on scanning complex artifacts (archives, containers, IaC files, Kubernetes manifests) where secrets hide. We're integrating Gitleaks for pattern detection and focusing our differentiation on artifact intelligence.

## Strategic Pivot Decision

### Why We're Pivoting
- **Market Reality:** Gitleaks dominates the general secret scanning space (19K stars, 20M+ downloads)
- **Competitive Analysis:** Our 80 detectors vs Gitleaks' 700+ rules = we can't win on detection breadth
- **Our Advantage:** Deep artifact scanning (streaming archives/containers without disk extraction) is genuinely unique
- **Pre-launch Status:** No v1.0 released yet, so we can pivot without breaking users

### New Positioning
**Before:** "Find secrets in your repo with low noise"
**After:** "Deep artifact scanner for cloud-native environments - powered by Gitleaks"

**Target Market:** DevSecOps teams at cloud-native companies dealing with:
- Container images in registries
- Helm charts and Kubernetes manifests
- Terraform state files and IaC artifacts
- Build artifacts (JARs, wheels, Docker images)
- CI/CD pipeline outputs

## Current Architecture

### What We Keep (Differentiators)
1. **Deep Artifact Scanning** (`internal/artifacts/`) - CROWN JEWEL
   - Streams archives (zip, tar, tgz) without extracting to disk
   - Scans container images with layer awareness
   - IaC hotspot detection (Terraform state, kubeconfigs)
   - Virtual paths: `archive.zip::layer-abc123::etc/app.yaml`
   - Guardrails: bytes, entries, depth, time budgets

2. **Structured Parsing** (`internal/ctxparse/`)
   - JSON/YAML parsing with line mapping
   - Adds context to findings inside complex artifacts

3. **Remediation Suite**
   - `fix` commands (forward-only: remove, redact, dotenv)
   - `purge` commands (history rewrite: path, pattern, replace)

4. **Enterprise Features**
   - SARIF output for GitHub Code Scanning
   - JSON upload capability
   - Public Go API (`pkg/core`)

### What We're Replacing
- **All 82 custom detectors** (4,066 LOC in `/internal/detectors/`)
  - Will integrate Gitleaks binary instead
  - Reduces maintenance, gets 700+ rules instantly
  - Positions us as complementary, not competitive

### What We're Adding (New Focus Areas)

#### Phase 1: Core Integration (Months 1-3)
- [ ] Gitleaks binary integration
- [ ] Virtual path remapping (so Gitleaks findings show artifact context)
- [ ] Helm chart scanning
- [ ] Kubernetes manifest secret detection
- [ ] Enhanced container scanning (OCI format, multi-arch)

#### Phase 2: Registry Integration (Months 4-6)
- [ ] Docker registry scanning (Docker Hub, GCR, ECR, ACR)
- [ ] Scan-on-push webhooks
- [ ] Image vulnerability + secret correlation
- [ ] CI/CD artifact scanning (scan the build outputs)

#### Phase 3: Advanced Features (Months 7-9)
- [ ] VM image scanning (AMI, VMDK, QCOW2)
- [ ] Dependency scanning (node_modules, vendor/)
- [ ] Simple web UI for artifact visualization
- [ ] Policy enforcement framework

## Technical Implementation Plan

### Gitleaks Integration Architecture

```go
// New package: internal/scanner/gitleaks.go
type GitleaksScanner struct {
    binaryPath string  // Auto-detect or download
    configPath string  // .gitleaks.toml
}

func (g *GitleaksScanner) ScanContent(virtualPath string, data []byte) ([]Finding, error) {
    // 1. Write content to temp file
    // 2. Invoke gitleaks detect --no-git --report-format json
    // 3. Parse JSON output
    // 4. Remap file paths to virtual paths
    // 5. Return findings with artifact context
}
```

### Configuration Evolution

```yaml
# .redactyl.yml (new format)
gitleaks:
  config: .gitleaks.toml          # Standard Gitleaks config
  binary: /usr/local/bin/gitleaks # Optional: custom path
  auto_download: true              # Download if missing

artifacts:
  archives: true
  containers: true
  helm: true                       # NEW
  kubernetes: true                 # NEW
  terraform: true
  dependencies: false              # Future
  max_archive_bytes: 67108864
  scan_time_budget: 30s
  global_artifact_budget: 5m

output:
  format: sarif
  include_artifact_metadata: true
  preserve_virtual_paths: true
```

### Virtual Path Enhancement

Current: `archive.zip::inner.tar::file.txt`

Enhanced:
```
myapp-image.tar::layer-sha256:abc123def::etc/secrets/config.yaml
  Context:
    - Image: myapp:v1.2.3
    - Layer: 5/12 (COPY --from=builder /app .)
    - Size: 2.1 KB
    - Detector: aws-access-key-id
    - Confidence: high
```

## Product Roadmap

### Milestone 1: Core Pivot (Q1 2025)
- Gitleaks integration complete
- Enhanced container scanning
- Helm + K8s manifest support
- Updated documentation
- **Launch blog:** "Why we're building on Gitleaks"

### Milestone 2: Registry Integration (Q2 2025)
- Docker registry connectors
- Scan-on-push webhooks
- CI/CD pipeline artifact scanning
- First 10 paying customers

### Milestone 3: Enterprise Features (Q3 2025)
- Simple web UI (artifact visualization)
- Policy engine (block deployments with secrets)
- SSO integration
- Multi-tenant support

### Milestone 4: Advanced Scanning (Q4 2025)
- VM image support
- Dependency tree scanning
- Supply chain analysis
- 50+ paying customers

## Competitive Differentiation

### vs Gitleaks
- **Complementary, not competitive**
- "Gitleaks for cloud-native artifacts"
- Gitleaks scans repos, we scan everything else
- Can contribute deep-scanning upstream

### vs GitGuardian
- Privacy-first (self-hosted, zero telemetry)
- Specialized in artifact complexity
- Lower cost (OSS core + paid enterprise)

### vs TruffleHog
- Better artifact handling (streaming, not extraction)
- Kubernetes/Helm native understanding
- Policy enforcement built-in

## Business Model

### Open Source (Core)
- CLI tool (MIT license)
- Gitleaks integration
- Deep artifact scanning
- SARIF/JSON output
- Community support

### Enterprise (Paid)
- Web dashboard
- Registry integrations
- Policy enforcement
- SSO/RBAC
- SLA support
- Hosted option

**Pricing (draft):**
- Free: OSS CLI
- Team: $99/month (up to 10 users)
- Business: $499/month (unlimited users, registry integration)
- Enterprise: Custom (hosted, SSO, SLA)

## Key Metrics to Track

### Technical
- Artifact types supported: 5 → 10
- Average scan time: < 30s for typical container
- False positive rate: < 5%
- Registry integrations: 0 → 4

### Business
- GitHub stars: 0 → 1,000 (6 months)
- Weekly active scans: 0 → 10,000
- Paying customers: 0 → 50 (12 months)
- Community contributors: 0 → 10

## Marketing Strategy

### Content Marketing
1. **Technical blog series:**
   - "Why scanning Docker images is harder than it looks"
   - "How we stream 10GB archives without disk I/O"
   - "The hidden secrets in your Helm charts"

2. **Open source presence:**
   - Contribute to Gitleaks
   - Speak at KubeCon/CloudNativeCon
   - Create awesome-secret-scanning list

3. **Community building:**
   - Discord for DevSecOps engineers
   - Weekly artifact scanning tips
   - Bug bounty for new artifact formats

### Launch Plan
- Private beta: 20 design partners (cloud-native companies)
- Public launch: HackerNews, DevSecOps subreddit
- Partnership: Gitleaks blog post collaboration

## Development Priorities

### Now (Next 30 Days)
1. ✅ Create CLAUDE.md (this file)
2. ⬜ Update README.md with new positioning
3. ⬜ Create Gitleaks integration package
4. ⬜ Implement basic binary detection/download
5. ⬜ Add virtual path remapping

### Next (30-60 Days)
6. ⬜ Helm chart scanning
7. ⬜ Kubernetes manifest detection
8. ⬜ Enhanced container format support
9. ⬜ Write technical blog post
10. ⬜ Launch private beta

### Later (60-90 Days)
11. ⬜ Registry integration (start with Docker Hub)
12. ⬜ Policy engine prototype
13. ⬜ Simple web UI
14. ⬜ Public launch prep

## Technical Debt & Cleanup

### To Remove
- [ ] Delete `/internal/detectors/` once Gitleaks integration is stable
- [ ] Remove `--enable/--disable` detector flags
- [ ] Clean up `redactyl detectors` command
- [ ] Remove custom confidence scoring (use Gitleaks mechanism)

### To Refactor
- [ ] Simplify config system (merge with Gitleaks TOML)
- [ ] Extract artifact streaming into standalone library
- [ ] Improve test coverage for artifact scanning (currently ~60%)

## Questions & Decisions Log

### Open Questions
- **Q:** Should we vendor Gitleaks binary or require installation?
  - **Leaning:** Auto-download on first use (like terraform)

- **Q:** Do we need our own detector rules at all?
  - **Leaning:** No, use Gitleaks exclusively

- **Q:** How do we handle Gitleaks version compatibility?
  - **Leaning:** Pin to tested version, allow override

### Decisions Made
- ✅ Pivot to deep artifact scanning (2025-10-27)
- ✅ Integrate Gitleaks instead of competing (2025-10-27)
- ✅ Keep pre-launch, no v1.0 migration needed (2025-10-27)
- ✅ Target DevSecOps/cloud-native market (2025-10-27)

## Resources & Links

### Competition Analysis
- Gitleaks: https://github.com/gitleaks/gitleaks
- TruffleHog: https://github.com/trufflesecurity/trufflehog
- GitGuardian: https://www.gitguardian.com/

### Technical References
- Gitleaks config format: https://github.com/gitleaks/gitleaks#configuration
- SARIF spec: https://docs.oasis-open.org/sarif/sarif/v2.1.0/
- OCI image spec: https://github.com/opencontainers/image-spec

### Market Research
- DevSecOps market size: $7.1B (2024)
- Container security TAM: $1.5B
- Secret management market: $2.8B

## Communication Style

When working on this project:
- **Be direct** - No fluff, just facts
- **Technical depth** - This is for engineers
- **Pragmatic** - Focus on shipping, not perfection
- **Competitive** - We're building to win the artifact scanning niche

## Notes for Future Claude Sessions

**Before suggesting new features, check:**
1. Does this strengthen our artifact scanning differentiation?
2. Could Gitleaks already do this (if so, use Gitleaks)?
3. Does this fit the cloud-native DevSecOps target market?

**When in doubt:**
- Prioritize artifact scanning depth over breadth
- Prefer integration over reinvention
- Focus on enterprise pain points (compliance, scale, automation)

---

*This document is living context. Update it as strategy evolves.*
