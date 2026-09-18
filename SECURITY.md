# Security

- Change default web passwords (`WEB_*_PASSWORD`) before production.
- Restrict MikroTik API access to trusted networks only.
- Prefer running the panel on a dedicated host; if on-router container is used, mount `DATA_DIR` on external storage.
- Log viewer is read-only (`/log/print`).
- This project does not implement HTTPS; terminate TLS on a reverse proxy if needed.
