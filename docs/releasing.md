# Releasing

Releases are published by `.github/workflows/release.yml` after pushing a
semantic-version tag such as `v1.2.3`. Pull requests never run this workflow.
The workflow cross-builds `figma` with `CGO_ENABLED=0` for Linux, macOS, and
Windows on amd64 and arm64, then uploads six archives and `SHA256SUMS` to the
matching GitHub Release.

## Local verification

Before pushing a tag, verify each target with the same repository module and
build flags used by CI:

```sh
rm -rf /tmp/figma-release-check && mkdir -p /tmp/figma-release-check
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  suffix=""
  [ "$os" = windows ] && suffix=.exe
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -buildvcs=false -o "/tmp/figma-release-check/figma-${os}-${arch}${suffix}" ./cmd/figma
done
sha256sum /tmp/figma-release-check/figma-* > /tmp/figma-release-check/SHA256SUMS
sha256sum -c /tmp/figma-release-check/SHA256SUMS
```

The workflow additionally normalizes archive metadata and sorts archive names
before generating its checksum manifest. To publish, push the tag:

```sh
git tag v1.2.3
git push origin v1.2.3
```

Only the six generated archives and `SHA256SUMS` are attached to the release.
