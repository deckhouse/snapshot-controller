# Patches

## 001-fix-cve.patch

Fix CVE reported by trivy against the `snapshot-controller` binary built from
external-snapshotter v8.6.0:

- `google.golang.org/grpc` v1.81.1 -> **v1.83.2** for GHSA-hrxh-6v49-42gf
  (fixed in 1.82.1), **CVE-2026-84303** and **CVE-2026-84304** (fixed in
  1.83.1) and **CVE-2026-84445** (fixed in 1.83.2);
- `golang.org/x/crypto` v0.52.0 -> **v0.56.0** for **CVE-2026-56854**,
  **CVE-2026-56855** and **CVE-2026-78662**;
- `golang.org/x/text` v0.37.0 -> v0.39.0 or later for **CVE-2026-56852**
  (resolved to v0.41.0 by the grpc requirements);
- `golang.org/x/net` v0.54.0 -> **v0.58.0**, the minimum grpc v1.83.2
  requires, which also covers CVE-2026-25680, CVE-2026-25681, CVE-2026-27136,
  CVE-2026-39821, CVE-2026-42502, CVE-2026-42506 and CVE-2026-46600;
- `github.com/google/cel-go` v0.28.1 -> **v0.29.0** for GHSA-gcjh-h69q-9w9g;
- the six `go.opentelemetry.io/otel*` modules (core, `metric`, `sdk`,
  `trace` and the two OTLP trace exporters) v1.43.0 -> **v1.45.0** for
  **CVE-2026-81870**. OpenTelemetry-Go releases the whole set together, so
  all six move to the same tag.

`go.mod`/`go.sum` only -- the build drops upstream `vendor/` and resolves from
the proxy. `golang.org/x/sys`, `x/sync`, `x/term`, `genproto`,
`go.opentelemetry.io/proto/otlp` and `go-logr/logr` follow as transitive
requirements. The image's OpenVEX document names the resulting
`golang.org/x/crypto` version.
