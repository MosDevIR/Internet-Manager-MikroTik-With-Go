# ۳. ساخت ایمیج Docker و باینری

## ساخت باینری (بدون Docker)

```bash
cd internet-manager-go
go mod tidy

# amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-amd64

# arm64 (RB5009، CCR و ...)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-arm64

# armv7 (بسیاری از hAP)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -trimpath -ldflags="-s -w" -o internet-manager-arm
```

حجم معمول بعد از strip: حدود ۸–۱۲ مگابایت.

## ساخت ایمیج با Dockerfile پروژه

```bash
docker build -t internet-manager:latest .
```

برای معماری خاص (مثلاً arm64 روی سیستم amd64):

```bash
docker buildx build --platform linux/arm64 -t internet-manager:arm64 --load .
```

## خروجی tar برای آپلود به میکروتیک

```bash
docker save internet-manager:arm64 -o internet-manager-arm64.tar
# یا فشرده:
docker save internet-manager:arm64 | gzip > internet-manager-arm64.tar.gz
```

فایل tar را روی روتر (مثلاً `disk1/`) کپی کنید و با `/container/add file=...` استفاده کنید.

## اجرا محلی برای تست

```bash
docker run --rm -p 5000:5000 \
  -e API_HOST=192.168.88.1 \
  -e API_USER=admin \
  -e API_PASS=pass \
  -e DATA_DIR=/data \
  -v $(pwd)/data:/data \
  internet-manager:latest
```

## Dockerfile خلاصه
ایمیج نهایی از `scratch` ساخته می‌شود و فقط باینری + گواهی‌های CA را دارد → حداقل حجم و سطح حمله.
