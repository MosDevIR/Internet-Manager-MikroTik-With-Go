# 1. Project Overview — Internet Manager for MikroTik

## Goal
A lightweight web panel to dynamically manage user internet paths on MikroTik routers.  
Users can switch between multiple egress paths (ADSL, mobile, VPN, …). Admins control all users; superadmins define system structure (tables and interfaces).

## Key features
- Compatible with **RouterOS 6 and 7** (automatic version detection)
- Static binary typically **8–12 MB**
- Three access levels: user / admin / superadmin
- Normal and VPN interfaces (WireGuard, OpenVPN, L2TP, SSTP, …)
- Settings storage + automatic backup on a mountable path
- **Bilingual UI (Persian / English)**
- Modular structure for easy extension

## Roles

| Role | Access |
|------|--------|
| **User** | Change only their own internet |
| **Admin** | Manage all users + default route (main) + view logs |
| **Superadmin** | Naming, create tables (v7), table↔interface mapping + logs |

## Architecture
```
internal/
  config/     Environment and settings
  models/     Data models + settings.json store/backup
  mikrotik/   Router API (mangle, route, table, interface, logs)
  handlers/   Web routes and role checks
  i18n/       FA / EN translations
```

## How it works with MikroTik
1. Connects via **binary API** (default port 8728).
2. Creates per-user `mangle` rules with `mark-routing`.
3. Traffic for that IP is directed to the selected routing table.
4. On RouterOS 7, tables must exist with the `fib` flag (created on demand).
5. On RouterOS 6, the `routing-mark` parameter is used.

## Operational notes
- Enable API on the router (`/ip service enable api`).
- API user needs rights for firewall mangle, routes, and interfaces.
- On-router containers need ROS 7.4+ and suitable hardware.
- Prefer external storage for `DATA_DIR` so settings survive container rebuilds.
