# ۲. نصب روی MikroTik (Container)

## پیش‌نیاز
- RouterOS **۷.۴ یا بالاتر**
- پکیج `container` نصب و device-mode فعال
- فضای کافی (ترجیحاً USB/SSD؛ برای تست می‌توان tmpfs استفاده کرد)
- معماری: arm / arm64 / x86

## ۱) فعال‌سازی Container
```
/system/device-mode/update container=yes
```
بعد از درخواست، دکمه ریست فیزیکی را بزنید و صبر کنید تا ریبوت شود.

## ۲) آماده‌سازی شبکه veth
```
/interface/veth/add name=veth-im address=172.17.0.2/24 gateway=172.17.0.1
/interface/bridge/add name=containers
/ip/address/add address=172.17.0.1/24 interface=containers
/interface/bridge/port/add bridge=containers interface=veth-im
/ip/firewall/nat/add chain=srcnat action=masquerade src-address=172.17.0.0/24
```

## ۳) مسیر داده و بکاپ روی دیسک
```
/disk/print
# فرض: usb1 یا disk1
/file/add name=disk1/im-data type=directory
```

## ۴) ساخت کانتینر از ایمیج محلی یا registry
اگر ایمیج را از قبل به صورت tar روی روتر کپی کرده‌اید:
```
/container/add file=disk1/internet-manager.tar \
  interface=veth-im \
  root-dir=disk1/im-root \
  mounts=im-data \
  envlist=im-envs \
  start-on-boot=yes \
  logging=yes
```

تعریف mount:
```
/container/mounts/add name=im-data src=disk1/im-data dst=/data
```

## ۵) تنظیم متغیرهای محیطی (Env) روی خود میکروتیک
```
/container/envs
add list=im-envs key=API_HOST value=192.168.88.1
add list=im-envs key=API_USER value=apiuser
add list=im-envs key=API_PASS value=StrongPass
add list=im-envs key=API_PORT value=8728
add list=im-envs key=WEB_PORT value=5000
add list=im-envs key=DATA_DIR value=/data
add list=im-envs key=WEB_USER_PASSWORD value=123
add list=im-envs key=WEB_ADMIN_PASSWORD value=123456
add list=im-envs key=WEB_SUPERADMIN_PASSWORD value=123456789
```

`API_HOST` را روی IP خود روتر از دید کانتینر تنظیم کنید (معمولاً gateway شبکه containers، مثلاً `172.17.0.1` یا IP LAN).

## ۶) پورت‌فوروارد برای دسترسی به پنل
```
/ip/firewall/nat/add chain=dstnat protocol=tcp dst-port=5000 \
  action=dst-nat to-addresses=172.17.0.2 to-ports=5000
```

## ۷) استارت
```
/container/start [find name~"internet"]
/container/print
```

پنل روی `http://IP-روتر:5000` در دسترس است.

## بکاپ تنظیمات
فایل‌ها روی mount ذخیره می‌شوند:
- `/data/settings.json` — تنظیمات فعلی
- `/data/settings-backup-YYYYMMDD-HHMMSS.json` — بکاپ‌های خودکار (حداکثر ۵ عدد)

با کپی پوشه `disk1/im-data` می‌توانید تنظیمات را نگه دارید.

## نکته امنیتی
- رمزهای پیش‌فرض را عوض کنید.
- سرویس API را فقط از شبکه داخلی باز بگذارید.
- در محیط واقعی ترجیحاً از HTTPS معکوس (یا دسترسی فقط از LAN) استفاده کنید.
