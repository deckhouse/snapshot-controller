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

## 004-fix-cve-grpc-cel-go.patch

Fix CVE in `google.golang.org/grpc` and `github.com/google/cel-go` reported by
trivy against the `snapshot-controller` binary. `google.golang.org/grpc`
v1.80.0 -> v1.82.1 (GHSA-hrxh-6v49-42gf, HIGH: xDS RBAC and HTTP/2
vulnerabilities). `github.com/google/cel-go` is vulnerable in
`>= 0.22.0, <= 0.28.1` (GHSA-gcjh-h69q-9w9g, MEDIUM: private JSON fields
exposed via `NativeTypes` and `ParseStructTag`) and is fixed only in 0.29.0,
whose `ext.TwoVarComprehensions` signature no `k8s.io/apiserver` release below
v0.36 compiles against, so the k8s libraries move v0.35.0 -> v0.36.3 in the
same patch.
