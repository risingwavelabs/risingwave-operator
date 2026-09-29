# MinIO image

The persistent example and the MinIO e2e fixtures use the publicly pullable
[Chainguard MinIO image](https://images.chainguard.dev/directory/image/minio/overview).
The upstream MinIO community repository is archived and its old image references
are no longer a maintained source.

All six manifests pin the same multi-platform digest, supporting Linux amd64 and
arm64. The image includes `/bin/sh`, `mkdir`, `/usr/bin/minio`, and `/usr/bin/mc`;
it does not use the upstream `/usr/bin/docker-entrypoint.sh`. The manifests create
the bucket directory before starting MinIO and run the server as UID/GID 65532.
`fsGroup` makes the data volume writable by that user. The e2e fixtures use
`emptyDir`; the persistent example retains its PVC.

The example credentials are for demonstration. Replace them before exposing the
service or storing sensitive data. When upgrading an existing deployment, back up
its data and confirm the storage driver supports `fsGroup` volume permissions.

## Verification and updates

The digest selected on 2026-09-29 was built on 2026-09-28 and contains MinIO
`0.20260922.192518-r0`. Its image signature, architecture-specific SPDX SBOM
signatures, and amd64 SLSA build provenance were verified with Cosign 3.1.3 against
the publisher identity documented on the
[MinIO provenance page](https://images.chainguard.dev/directory/image/minio/provenance).

To verify the pinned image and scan its installed packages without Docker:

```sh
image=$(sed -n 's/^        image: \(cgr.dev\/chainguard\/minio@sha256:.*\)$/\1/p' \
  docs/manifests/stable/persistent/minio/risingwave.yaml)
cosign verify \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  --certificate-identity=https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main \
  "$image"
grype "registry:$image" --platform linux/amd64
grype "registry:$image" --platform linux/arm64
```

The amd64 filesystem scan with Grype 0.119.0 and its 2026-09-29 database found no
high or critical vulnerabilities. It reported four medium findings in glibc:

| Finding | Relevant behavior |
| --- | --- |
| [CVE-2026-77117](https://explore.alas.aws.amazon.com/CVE-2026-77117.html) | Crafted SHIFT_JISX0213 conversion can hang a caller. |
| [CVE-2026-80489](https://ubuntu.com/security/CVE-2026-80489) | Crafted EUC_JISX0213 conversion can hang a caller. |
| [CVE-2026-8674](https://ubuntu.com/security/CVE-2026-8674) | An unusually long DNS search domain can abort a process using the glibc resolver. |
| [CVE-2026-89092](https://access.redhat.com/security/cve/cve-2026-89092) | A malicious DNS response can crash the `nscd` service. |

These manifests do not invoke character conversion or start `nscd`; cluster DNS
configuration remains an environmental consideration. This scan is a dated
assessment, not a guarantee that the image has no vulnerabilities. Review the
[publisher's current vulnerability report](https://images.chainguard.dev/directory/image/minio/vulnerabilities)
when updating or deploying it.

A digest pin does not receive security updates automatically. To update, resolve
the public `latest` tag to its new multi-platform digest, verify the publisher
signature and SBOMs, scan both architectures, and replace the digest in all six
manifests together. Run `bash ci/scripts/test-minio-images.sh` against a disposable
Kubernetes cluster with the RisingWave CRD installed. The PR workflow runs this
test to check readiness, bucket access, S3 writes and reads, and PVC persistence
across pod replacement.
