# 6. Development guide

## Adding a feature
1. Router API logic → `internal/mikrotik/client.go`
2. Data models → `internal/models/`
3. Web routes & roles → `internal/handlers/handlers.go`
4. UI strings → `internal/i18n/strings.go` (FA + EN)
5. Templates → `templates/`

Keep `main.go` thin.

## Local testing without a physical router
Use CHR (Cloud Hosted Router) or a lab device with API enabled, then point env vars at it.

## Quick rebuild
```bash
go build -trimpath -ldflags="-s -w" -o internet-manager .
```

## i18n
- Language detection: query `?lang=`, cookie `lang`, then `Accept-Language`
- All user-visible UI strings live in `internal/i18n/strings.go`
- Templates use `{{index .T "key"}}`

## Suggested next features
- Real block via address-list / filter
- Multi-router support
- Built-in TLS or documented reverse-proxy setup
- Admin action audit log
- Auth beyond shared passwords (user file / better accounts)
