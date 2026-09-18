# ۶. راهنمای توسعه

## اضافه کردن قابلیت جدید
1. منطق API → `internal/mikrotik/client.go`
2. مدل داده → `internal/models/`
3. مسیر وب و نقش → `internal/handlers/handlers.go`
4. UI → `templates/`

`main.go` را تا حد ممکن دست‌نخورده نگه دارید.

## تست محلی بدون روتر واقعی
می‌توانید با CHR (Cloud Hosted Router) یا یک روتر آزمایشی API را باز کنید و با env به آن وصل شوید.

## بیلد سریع بعد از تغییر
```bash
go build -trimpath -ldflags="-s -w" -o internet-manager .
```

## پیشنهادهای توسعه بعدی
- بلاک واقعی با address-list / filter
- پشتیبانی چند روتر
- HTTPS داخلی یا TLS termination
- لاگ عملیات ادمین
- احراز هویت بهتر از رمز ثابت (مثلاً فایل کاربران)
