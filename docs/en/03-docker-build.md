# 3. Binary & Docker image build

## Build binary (no Docker)

```bash
cd internet-manager-go
go mod tidy

# amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-amd64

# arm64 (RB5009, CCR, …)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-arm64

# armv7 (many hAP devices)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-arm
```

Typical size after strip: about 8–12 MB.

## Build image with project Dockerfile

```bash
docker build -t internet-manager:latest .
```

Cross-build example (arm64 on amd64 host):

```bash
docker buildx build --platform linux/arm64 -t internet-manager:arm64 --load .
```

## Export tar for MikroTik upload

```bash
docker save internet-manager:arm64 -o internet-manager-arm64.tar
# or compressed:
docker save internet-manager:arm64 | gzip > internet-manager-arm64.tar.gz
```

Copy the tar to the router (e.g. `disk1/`) and use `/container/add file=...`.

## Local test run

```bash
docker run --rm -p 5000:5000 \
  -e API_HOST=192.168.88.1 \
  -e API_USER=admin \
  -e API_PASS=pass \
  -e DATA_DIR=/data \
  -v $(pwd)/data:/data \
  internet-manager:latest
```

## Image notes
The final image is based on `scratch` (binary + CA certs only) for minimal size and attack surface.
