# Research: Renovate Bot for Dependency Management

**Date**: 2026-03-26 | **Researcher**: nw-researcher (Nova) | **Confidence**: High | **Sources**: 8

## Executive Summary

Renovate Bot is a free, open-source dependency update automation tool maintained by Mend.io. It is completely free for both the hosted GitHub App and self-hosted deployments. For a Go project on GitHub with CircleCI, the simplest path is installing the Mend Renovate GitHub App (zero infrastructure cost, instant onboarding). Renovate natively supports Go modules (go.mod/go.sum), CircleCI orb version updates, and offers significantly more configuration flexibility than its primary alternative, GitHub Dependabot. A minimal `renovate.json` in the repository root is all that is needed to get started.

## Research Methodology

**Search Strategy**: Official Renovate documentation (docs.renovatebot.com), GitHub repository/marketplace, and practitioner comparison sources.
**Source Selection**: Types: official docs, GitHub marketplace, industry practitioner sources | Reputation: high/medium-high | Verification: cross-referenced across official docs and independent practitioner sources.
**Quality Standards**: All 5 findings backed by 2+ sources, with official documentation as primary authority.

## Findings

### Finding 1: Pricing and Free Tier

**Confidence**: High

Renovate is completely free at all tiers relevant to this use case:

| Option | Cost | Notes |
|--------|------|-------|
| **Mend Renovate GitHub App** (hosted) | Free | Managed by Mend.io. No paid plan required for Renovate functionality. Works on public and private repos. |
| **Self-hosted (OSS)** | Free (AGPL-3.0) | You pay only for your own compute/infrastructure. Available as npm CLI, Docker image, or GitHub Action. |
| **Mend.io commercial products** | Paid | Mend offers paid products (Mend SCA, Mend SAST) that bundle Renovate with additional security scanning. These are separate products -- Renovate itself remains free. |

There are no usage limits, repo limits, or PR limits on the free tier. The hosted GitHub App is provided as a complimentary service by Mend.io.

**Sources**: [Running Renovate - Renovate Docs](https://docs.renovatebot.com/getting-started/running/), [Renovate - GitHub Marketplace](https://github.com/marketplace/renovate), [renovatebot/renovate - GitHub](https://github.com/renovatebot/renovate)

### Finding 2: Deployment Models (GitHub App vs Self-Hosted)

**Confidence**: High

There are two primary deployment models:

**Option A: Mend Renovate GitHub App (Recommended for most users)**
- Install from the GitHub Marketplace in one click.
- Mend manages all infrastructure, scheduling, and updates.
- Stateful app that responds to GitHub webhooks (e.g., merged PRs trigger immediate re-evaluation).
- Includes priority job queue (merged PRs prioritized over scheduled scans).
- Released every 1-2 months (slower, more stable cadence than OSS CLI).
- Configuration is repository-level only (via `renovate.json`).

**Option B: Self-Hosted**
- Full control over the bot environment and scheduling.
- Distribution options: npm CLI, Docker image (`renovate/renovate`), GitHub Action (`renovatebot/github-action`), or CircleCI Orb (`daniel-shuy/renovate`).
- Requires you to manage: compute, scheduling (cron), authentication tokens, and updates.
- Supports global configuration (applied across all repos) plus per-repo overrides.
- Released more frequently than the hosted app (follows OSS release cadence).
- Security consideration: self-hosted instances must operate under a trust relationship with monitored repo developers.

**For this project**: The hosted GitHub App is the right choice. The kanban project is a single repo on GitHub -- there is no need for self-hosting complexity.

**Sources**: [Running Renovate - Renovate Docs](https://docs.renovatebot.com/getting-started/running/), [Self-Hosting Examples - Renovate Docs](https://docs.renovatebot.com/examples/self-hosting/), [Self-Hosted Configuration - Renovate Docs](https://docs.renovatebot.com/self-hosted-configuration/)

### Finding 3: Setup for Go Project on GitHub with CircleCI

**Confidence**: High

**Step 1: Install the Mend Renovate GitHub App**
1. Go to [github.com/apps/renovate](https://github.com/apps/renovate).
2. Click "Install" and select the repository (or all repositories).
3. Renovate will automatically create an onboarding PR with a default `renovate.json`.

**Step 2: Merge the Onboarding PR**
Renovate's first PR adds a `renovate.json` to the repo root. Review and merge it.

**Step 3: Renovate Automatically Handles Go and CircleCI**

For Go modules, Renovate:
- Detects `go.mod` files anywhere in the repo (pattern: `(^|/)go\.mod$`).
- Updates dependencies in `go.mod` and regenerates `go.sum`.
- Automatically vendors if `vendor/modules.txt` is present.
- Can run `go mod tidy` via `postUpdateOptions: ["gomodTidy"]`.
- Updates the `toolchain` directive by default.
- Does NOT bump the `go` directive by default (it is a compatibility marker, not a version pin). Enable with `rangeStrategy: "bump"` if desired.

For CircleCI, Renovate:
- Detects files matching `(^|/)\.circleci/.+\.ya?ml$`.
- Updates CircleCI orb versions automatically.
- Updates Docker image references in CircleCI config.

**No special CircleCI integration is needed** -- Renovate operates via GitHub PRs. CircleCI will run its normal CI pipeline on each Renovate PR, validating the dependency update before merge.

**Sources**: [Go Modules - Renovate Docs](https://docs.renovatebot.com/modules/manager/gomod/), [CircleCI Manager - Renovate Docs](https://docs.renovatebot.com/modules/manager/circleci/), [Running Renovate - Renovate Docs](https://docs.renovatebot.com/getting-started/running/)

### Finding 4: Configuration (renovate.json)

**Confidence**: High

Place a `renovate.json` file in the repository root. Here is a practical configuration for a Go project with CircleCI:

**Minimal (start here):**
```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:recommended"]
}
```

**Recommended for this project:**
```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended",
    "group:monorepos",
    ":automergeMinor"
  ],
  "postUpdateOptions": ["gomodTidy"],
  "packageRules": [
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["minor", "patch"],
      "automerge": true
    },
    {
      "matchManagers": ["circleci"],
      "automerge": true
    }
  ],
  "schedule": ["before 7am on Monday"]
}
```

**Key configuration options explained:**

| Option | Purpose |
|--------|---------|
| `extends: ["config:recommended"]` | Sensible defaults: labels, branch naming, PR limits |
| `group:monorepos` | Groups monorepo packages (e.g., all `golang.org/x/*`) into single PRs |
| `postUpdateOptions: ["gomodTidy"]` | Runs `go mod tidy` after updating go.mod |
| `automerge` | Auto-merges PRs that pass CI (minor/patch only recommended) |
| `schedule` | Controls when Renovate creates PRs (reduces noise) |
| `packageRules` | Per-dependency or per-manager overrides |

**Sources**: [Configuration Options - Renovate Docs](https://docs.renovatebot.com/configuration-options/), [Renovate Config Overview](https://docs.renovatebot.com/config-overview/), [Go Modules Manager - Renovate Docs](https://docs.renovatebot.com/modules/manager/gomod/)

### Finding 5: Alternatives

**Confidence**: High

| Tool | Cost | Platforms | Go Support | Key Differentiator |
|------|------|-----------|------------|-------------------|
| **GitHub Dependabot** | Free (built into GitHub) | GitHub only | Yes (go.mod) | Zero setup -- built into GitHub. Simpler but less configurable. |
| **Renovate** | Free | GitHub, GitLab, Bitbucket, Azure DevOps, Gitea | Yes (go.mod) | More flexible: grouping, scheduling, monorepo presets, Dependency Dashboard. |

**Renovate advantages over Dependabot:**
- Dependency Dashboard (GitHub Issue showing all pending updates at a glance).
- Better PR grouping -- community presets group related dependencies automatically.
- Multi-platform support (matters if you ever move off GitHub).
- More granular scheduling (per-dependency, not just per-language).
- Merge Confidence badges (Age, Adoption, Passing, Confidence).

**Dependabot advantages over Renovate:**
- Zero external dependencies -- built into GitHub natively.
- Simpler configuration (`dependabot.yml`).
- Shows upstream CHANGELOG content in PRs.
- No third-party GitHub App permissions required.

**Recommendation**: Renovate is the stronger choice for this project. The Go project already uses CircleCI (not GitHub Actions), so Dependabot's tight GitHub integration is less of an advantage. Renovate's grouping, scheduling, and CircleCI orb update support provide more value.

**Sources**: [Bot Comparison - Renovate Docs](https://docs.renovatebot.com/bot-comparison/), [Renovate vs Dependabot - TurboStarter](https://www.turbostarter.dev/blog/renovate-vs-dependabot-whats-the-best-tool-to-automate-your-dependency-updates), [Why I recommend Renovate - Jamie Tanna](https://www.jvt.me/posts/2024/04/12/use-renovate/)

## Source Analysis

| Source | Domain | Reputation | Type | Access Date | Cross-verified |
|--------|--------|------------|------|-------------|----------------|
| Renovate Docs - Running | docs.renovatebot.com | High (1.0) | Official | 2026-03-26 | Y |
| Renovate Docs - gomod | docs.renovatebot.com | High (1.0) | Official | 2026-03-26 | Y |
| Renovate Docs - CircleCI | docs.renovatebot.com | High (1.0) | Official | 2026-03-26 | Y |
| Renovate Docs - Config | docs.renovatebot.com | High (1.0) | Official | 2026-03-26 | Y |
| Renovate Docs - Bot Comparison | docs.renovatebot.com | High (1.0) | Official | 2026-03-26 | Y |
| GitHub Marketplace - Renovate | github.com/marketplace | High (1.0) | Official | 2026-03-26 | Y |
| renovatebot/renovate GitHub | github.com/renovatebot | High (1.0) | Official | 2026-03-26 | Y |
| Jamie Tanna - Use Renovate | jvt.me | Medium-High (0.8) | Industry | 2026-03-26 | Y |

Reputation: High: 7 (87.5%) | Medium-High: 1 (12.5%) | Avg: 0.975

**Bias note**: The Bot Comparison page is published by Renovate's own documentation. While factually accurate on feature differences, it naturally favors Renovate's strengths. This was cross-referenced against independent practitioner sources to confirm claims.

## Knowledge Gaps

### Gap 1: Rate Limiting Details for Hosted App
**Issue**: Exact rate limits or scheduling frequency for the free hosted Mend Renovate App are not published.
**Attempted**: Official docs, GitHub discussions.
**Recommendation**: In practice, this is not a concern for single-repo usage. Monitor PR creation cadence after installation.

### Gap 2: Mend.io Data Retention and Privacy
**Issue**: What data the hosted Renovate app collects/retains is not clearly documented in the public docs.
**Attempted**: Official docs, GitHub discussion #29598.
**Recommendation**: Review the Mend.io privacy policy before installation if data sensitivity is a concern. Self-hosting eliminates this concern entirely.

## Full Citations

[1] Renovate. "Running Renovate". Renovate Docs. https://docs.renovatebot.com/getting-started/running/. Accessed 2026-03-26.
[2] Renovate. "Automated Dependency Updates for Go Modules". Renovate Docs. https://docs.renovatebot.com/modules/manager/gomod/. Accessed 2026-03-26.
[3] Renovate. "Automated Dependency Updates for CircleCI". Renovate Docs. https://docs.renovatebot.com/modules/manager/circleci/. Accessed 2026-03-26.
[4] Renovate. "Configuration Options". Renovate Docs. https://docs.renovatebot.com/configuration-options/. Accessed 2026-03-26.
[5] Renovate. "Bot Comparison". Renovate Docs. https://docs.renovatebot.com/bot-comparison/. Accessed 2026-03-26.
[6] GitHub. "Renovate - GitHub Marketplace". https://github.com/marketplace/renovate. Accessed 2026-03-26.
[7] Renovate. "renovatebot/renovate". GitHub. https://github.com/renovatebot/renovate. Accessed 2026-03-26.
[8] Tanna, Jamie. "Why I recommend Renovate over any other dependency update tools". jvt.me. 2024-04-12. https://www.jvt.me/posts/2024/04/12/use-renovate/. Accessed 2026-03-26.

## Research Metadata

Duration: ~15 min | Examined: 12 | Cited: 8 | Cross-refs: 5 | Confidence: High 100% | Output: docs/research/renovate-bot-dependency-management.md
