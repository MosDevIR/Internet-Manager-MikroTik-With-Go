# 5. Environment variables & settings backup

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_HOST` | 192.168.88.1 | Router address as seen by the app |
| `API_USER` | admin | API username |
| `API_PASS` | (empty) | API password |
| `API_PORT` | 8728 | API port |
| `WEB_PORT` | 5000 | Web panel port |
| `WEB_USER_PASSWORD` | 123 | User login password |
| `WEB_ADMIN_PASSWORD` | 123456 | Admin password |
| `WEB_SUPERADMIN_PASSWORD` | 123456789 | Superadmin password |
| `DATA_DIR` | /data | Directory for settings and backups |
| `SETTINGS_FILE` | `$DATA_DIR/settings.json` | Settings file path |

### Setting env on MikroTik container
```
/container/envs
add list=im-envs key=API_HOST value=172.17.0.1
add list=im-envs key=API_USER value=api
add list=im-envs key=API_PASS value=Secret
add list=im-envs key=DATA_DIR value=/data
add list=im-envs key=WEB_ADMIN_PASSWORD value=MyAdminPass
```
Then set `envlist=im-envs` on the container.

## Automatic backup
On every settings save:
1. Main file: `/data/settings.json`
2. Timestamped copy: `/data/settings-backup-YYYYMMDD-HHMMSS.json`
3. Keep at most 5 backups; older files are removed.

### Mount on MikroTik
```
/container/mounts/add name=im-data src=disk1/im-data dst=/data
```
Settings survive container remove/recreate.

### Manual restore
Copy a backup file over `settings.json` and restart the container (or wait for the next load).

## UI language
Independent of these env vars. Users switch **FA / EN** in the panel (cookie `lang`).
