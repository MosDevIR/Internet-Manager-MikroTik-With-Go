# ۵. متغیرهای محیطی و بکاپ تنظیمات

## متغیرهای محیطی

| متغیر | پیش‌فرض | توضیح |
|-------|---------|--------|
| `API_HOST` | 192.168.88.1 | آدرس روتر از دید برنامه |
| `API_USER` | admin | کاربر API |
| `API_PASS` | (خالی) | رمز API |
| `API_PORT` | 8728 | پورت API |
| `WEB_PORT` | 5000 | پورت پنل وب |
| `WEB_USER_PASSWORD` | 123 | رمز ورود کاربر |
| `WEB_ADMIN_PASSWORD` | 123456 | رمز ادمین |
| `WEB_SUPERADMIN_PASSWORD` | 123456789 | رمز سوپرادمین |
| `DATA_DIR` | /data | پوشه ذخیره settings و بکاپ |
| `SETTINGS_FILE` | `$DATA_DIR/settings.json` | مسیر فایل تنظیمات |

### ست کردن Env داخل کانتینر میکروتیک
```
/container/envs
add list=im-envs key=API_HOST value=172.17.0.1
add list=im-envs key=API_USER value=api
add list=im-envs key=API_PASS value=Secret
add list=im-envs key=DATA_DIR value=/data
add list=im-envs key=WEB_ADMIN_PASSWORD value=MyAdminPass
```
سپس در تعریف کانتینر: `envlist=im-envs`

## بکاپ خودکار
هر بار که تنظیمات ذخیره می‌شود:
1. فایل اصلی: `/data/settings.json`
2. یک کپی با timestamp: `/data/settings-backup-YYYYMMDD-HHMMSS.json`
3. حداکثر ۵ بکاپ آخر نگه داشته می‌شود؛ قدیمی‌ترها حذف می‌شوند.

### Mount روی میکروتیک
```
/container/mounts/add name=im-data src=disk1/im-data dst=/data
```
با این کار حتی بعد از حذف/بازسازی کانتینر، تنظیمات روی دیسک می‌ماند.

### بازیابی دستی
فایل بکاپ را روی `settings.json` کپی کنید و کانتینر را یک‌بار ریستارت کنید (یا صبر کنید تا دوباره Load شود در درخواست بعدی).
