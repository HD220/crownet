# External Dependencies Status

This document records the status of key external dependencies used in the CrowNet project.

## 1. `github.com/mattn/go-sqlite3`

-   **Purpose:** SQLite3 driver for Go, enabling database logging features.
-   **Version in `go.mod`:** `v1.14.28` (as of 2025-06-26)
-   **License:** MIT License (Compatible with project)
-   **Repository:** [https://github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

### Status Review (2025-06-26)

-   **Maintenance:** The library appears to be actively maintained. The version `v1.14.28` was tagged on April 16, 2025, indicating recent updates. Commit history on the repository generally shows ongoing activity.
-   **Latest Version:** The version used (`v1.14.28`) is the latest tagged release according to the repository's tag list.
-   **Vulnerabilities:** A search of the GitHub Advisory Database for "go-sqlite3" did not reveal any direct, reviewed vulnerabilities affecting `github.com/mattn/go-sqlite3` version `v1.14.28`.
    -   An advisory GHSA-9r4c-jwx3-3j76 (CVE-2025-24786) was found related to "Sqlite3 database" but was specific to the `github.com/clidey/whodb/core` package and its handling of file paths, not a vulnerability within `mattn/go-sqlite3` itself.
-   **CGO Requirement:** This is a CGO package and requires a C compiler (like GCC) to be available in the build environment. This is an important consideration for setting up development and deployment environments.

### Recommendation (as of 2025-06-26)

The current version (`v1.14.28`) of `github.com/mattn/go-sqlite3` appears to be suitable for use. It is up-to-date with the latest tagged release, actively maintained, has a compatible license, and no direct vulnerabilities were identified for this version.

Future checks should be performed periodically, especially before major releases of CrowNet or if new security advisories emerge.

---
*(This document should be updated periodically or as new dependencies are added/reviewed.)*
