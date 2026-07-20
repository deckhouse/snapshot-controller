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
`golang.org/x/net` v0.52.0 -> v0.56.0 (CVE-2026-46600, a panic in
`golang.org/x/net/dns/dnsmessage` on an invalid SVCB/HTTPS RR) and
`golang.org/x/text` v0.35.0 -> v0.39.0 (CVE-2026-56852, an infinite loop
on invalid input). This pulls the transitive `golang.org/x/crypto`
v0.53.0, `golang.org/x/sys` v0.46.0, `golang.org/x/sync` v0.21.0 and
`golang.org/x/term` v0.44.0 bumps.
