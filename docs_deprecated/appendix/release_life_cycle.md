# OpenGIN Release Lifecycle

This document defines the **release stages**, **versioning scheme**, and **naming conventions** used in the OpenGIN platform.  
It ensures consistency, clarity, and traceability across all releases and environments.

---

## Versioning Scheme

OpenGIN follows **Semantic Versioning (SemVer)** with optional pre-release identifiers and calendar tags.

```
MAJOR.MINOR.PATCH[-STAGE.NUMBER]
```

**Examples:**
- `1.0.0-alpha.1` — First internal alpha build  
- `1.0.0-beta.2` — Second beta release  
- `1.0.0-RC.1` — First release candidate  
- `1.0.0` — General Availability (Stable)  

---

## Alpha Release

**Definition:**  
> Early internal version used for architecture validation, core integration, and early-stage testing.

**Purpose:**  
- Validate service-to-service communication (e.g., Database layer, backing up, integration tests, etc.).  
- Test architecture, schema design, and API functionality.  
- Identify critical issues before public testing.

**Stability:** ❌ Not stable  
**Audience:** Internal developers and system engineers  
**Tag Example:** `1.0.0-alpha.1`

---

## Beta Release

**Definition:**  
> Feature-complete version intended for broader testing and validation from selected external users.

**Purpose:**  
- Ensure all modules (Graph Engine, Read API, Extractors, Frontend Studio) function correctly together.  
- Gather usability and performance feedback.  
- Identify and fix known issues before production readiness.

**Stability:** ⚙️ Moderate  
**Audience:** Internal + selected external testers  
**Tag Example:** `2.0.0-beta.1`

---

## Release Candidate (RC)

**Definition:**  
> A near-final version that could become the official release if no significant issues are found.

**Purpose:**  
- Verify stability under production-like conditions.  
- Validate end-to-end integrations and ensure all bugs are resolved.  
- Prepare deployment artifacts (Docker images, Helm charts, documentation).

**Stability:** 🧩 High  
**Audience:** QA and staging environments  
**Tag Example:** `2.0.0-RC.1`

---

## General Availability (GA)

**Definition:**  
> The official, production-ready release validated through testing, review, and documentation.

**Purpose:**  
- Fully tested and validated for production deployment.  
- Represents a stable baseline for users and partners.  
- Supported under maintenance and patch cycles.

**Stability:** ✅ Stable  
**Audience:** All users, partners, and production environments  
**Tag Example:** `2.0.0`

---

## Optional Calendar Tagging

Each major or minor release can optionally include a **calendar tag** to align with milestones.

**Examples:**
- `OpenGIN 1.0.0 (2025.09)` — September 2025 major release  
- `OpenGIN 1.1.0 (2026.Q1)` — First quarter 2026 update  

This helps track releases alongside development cycles or roadmap milestones.

---

## Branch Naming Convention

To maintain a consistent workflow across teams and CI/CD pipelines:

| Branch Type | Format | Description |
|--------------|--------|-------------|
| **Main** | `main` | Always points to the latest stable (GA) release |
| **Development** | `dev` | Used for integration and pre-release work |
| **Feature** | `feat-<feature-name>` | New functionality under development |
| **Release** | `release-<version>` | Prepares an upcoming release (alpha, beta, rc) |
| **Hotfix** | `hotfix-<version>` | Urgent fixes for production issues |

**Examples:**
- `release-2.0.0-beta.1`  
- `feature-graph-link-expansion`  
- `hotfix-v1.3.2`

---

## Summary Table

| Stage | Tag Example | Audience | Stability | Purpose |
|--------|--------------|-----------|------------|----------|
| **Alpha** | `2.0.0-alpha.1` | Internal | 🚧 Low | Architecture validation |
| **Beta** | `2.0.0-beta.1` | Testers | ⚙️ Medium | Feature validation |
| **RC** | `2.0.0-RC.1` | QA | 🧩 High | Final verification |
| **GA** | `2.0.0` | Public | ✅ Stable | Production deployment |

---

## Notes

- Increment numbers (`.1`, `.2`, `.3`, etc.) indicate iterations within a stage.  
- Only **GA** releases are considered official and stable for production.  
- Patch versions (e.g., `2.0.1`, `2.0.2`) are reserved for minor fixes.  
- Each release should include changelogs summarizing new features, improvements, and bug fixes.

---

_This standard ensures OpenGIN releases remain consistent, transparent, and traceable across all development and deployment environments._
