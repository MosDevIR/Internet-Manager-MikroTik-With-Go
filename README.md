# Internet Manager for MikroTik (Go)

<div dir="rtl">

## فارسی

پنل وب سبک برای مدیریت اینترنت / VPN روی MikroTik — سازگار با **RouterOS 6 و 7**.

### ویژگی‌ها
- سه نقش: کاربر / ادمین / سوپرادمین
- تفکیک اینترنت با mangle + routing mark/table
- پشتیبانی VPN (WireGuard, OpenVPN, L2TP, ...)
- لاگ لایو با فیلتر و بهینه‌سازی
- **رابط دوزبانه فارسی / English**
- بکاپ خودکار روی `DATA_DIR`
- حجم باینری حدود ۸–۱۲ مگابایت

### مستندات (دوزبانه)
- فهرست: [`docs/README.md`](docs/README.md)
- فارسی: [`docs/fa/`](docs/fa/)
- English: [`docs/en/`](docs/en/)

### ساخت
```bash
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o internet-manager
```

### زبان
در هدر پنل **FA** یا **EN** را بزنید (در کوکی ذخیره می‌شود).

</div>

---

## English

Lightweight web panel for MikroTik multi-WAN / VPN policy routing — **RouterOS 6 & 7**.

### Features
- Roles: user / admin / superadmin
- Policy routing via mangle + routing mark/table
- VPN interfaces (WireGuard, OpenVPN, L2TP, ...)
- Live filtered logs with performance optimizations
- **Bilingual UI (Persian / English)**
- Auto backup under `DATA_DIR`
- Binary size typically 8–12 MB

### Documentation (bilingual)
- Index: [`docs/README.md`](docs/README.md)
- Persian: [`docs/fa/`](docs/fa/)
- English: [`docs/en/`](docs/en/)

### Build
```bash
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o internet-manager
```

### Language
Use **FA** / **EN** in the panel header (stored in cookie `lang`).

### Suggested GitHub repository name
`internet-manager-mikrotik-with-go`

### Author
Mostafa Nekooei ([@MosDevIR](https://github.com/MosDevIR)) — nekooei.developer@gmail.com
