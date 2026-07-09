# Patches

## 001-fix-cve.patch

Fix CVE

## 002-fix-cve-otel-sdk.patch

Fix CVE-2026-39883 in `go.opentelemetry.io/otel/sdk` by upgrading
OpenTelemetry-Go modules from v1.40.0 to v1.43.0. The vulnerability is
a PATH hijacking flaw on BSD/Solaris caused by the `kenv` command not
using an absolute path.

## 003-fix-cve-golang-x.patch

Fix CVE in the `golang.org/x` modules reported by trivy by upgrading
`golang.org/x/net` v0.52.0 -> v0.55.0 (x/net/html and x/net/idna CVEs),
`golang.org/x/crypto` v0.49.0 -> v0.52.0 (x/crypto/ssh CVEs) and
`golang.org/x/sys` v0.42.0 -> v0.45.0 (x/sys/windows CVE-2026-39824).
