# 2. Install on MikroTik (Container)

## Requirements
- RouterOS **7.4 or newer**
- `container` package installed and device-mode enabled
- Enough storage (USB/SSD preferred)
- Architecture: arm / arm64 / x86

## 1) Enable containers
```
/system/device-mode/update container=yes
```
Then press the physical reset button when prompted and wait for reboot.

## 2) Prepare veth network
```
/interface/veth/add name=veth-im address=172.17.0.2/24 gateway=172.17.0.1
/interface/bridge/add name=containers
/ip/address/add address=172.17.0.1/24 interface=containers
/interface/bridge/port/add bridge=containers interface=veth-im
/ip/firewall/nat/add chain=srcnat action=masquerade src-address=172.17.0.0/24
```

## 3) Data path for settings & backups
```
/disk/print
/file/add name=disk1/im-data type=directory
```

## 4) Create container from local image
```
/container/mounts/add name=im-data src=disk1/im-data dst=/data

/container/add file=disk1/internet-manager.tar \
  interface=veth-im \
  root-dir=disk1/im-root \
  mounts=im-data \
  envlist=im-envs \
  start-on-boot=yes \
  logging=yes
```

## 5) Environment variables on the router
```
/container/envs
add list=im-envs key=API_HOST value=172.17.0.1
add list=im-envs key=API_USER value=apiuser
add list=im-envs key=API_PASS value=StrongPass
add list=im-envs key=API_PORT value=8728
add list=im-envs key=WEB_PORT value=5000
add list=im-envs key=DATA_DIR value=/data
add list=im-envs key=WEB_USER_PASSWORD value=123
add list=im-envs key=WEB_ADMIN_PASSWORD value=123456
add list=im-envs key=WEB_SUPERADMIN_PASSWORD value=123456789
```

Set `API_HOST` to the address the container uses to reach the router (often the containers bridge gateway).

## 6) Port forward to the panel
```
/ip/firewall/nat/add chain=dstnat protocol=tcp dst-port=5000 \
  action=dst-nat to-addresses=172.17.0.2 to-ports=5000
```

## 7) Start
```
/container/start [find name~"internet"]
/container/print
```

Open `http://ROUTER-IP:5000`.

## Settings backup
On the mount:
- `/data/settings.json` — current settings
- `/data/settings-backup-YYYYMMDD-HHMMSS.json` — rotating backups (max 5)

Copy `disk1/im-data` to preserve configuration.

## Security
- Change default web passwords.
- Restrict API to trusted networks.
- Prefer TLS via reverse proxy in production.
