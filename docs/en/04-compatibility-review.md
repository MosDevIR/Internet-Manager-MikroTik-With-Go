# 4. Professional compatibility review

## Summary
The project targets practical use on RouterOS 6 and 7. The main path (mangle + policy routing) works on both. ROS 7 specifics (`fib` tables and `routing-table`) are handled explicitly.

## Compatibility matrix

| Feature | ROS 6 | ROS 7 | Notes |
|---------|-------|-------|-------|
| Binary API (8728) | ✅ | ✅ | go-routeros/v3 |
| Version detection | ✅ | ✅ | `/system/resource/print` |
| mangle mark-routing | ✅ | ✅ | Shared |
| Custom route parameter | `routing-mark` | `routing-table` | Auto-selected |
| List tables | mangle+route | `/routing/table`+mangle+route | Merged |
| Create table + fib | Not required | ✅ | Only when IsV7 |
| DHCP leases | ✅ | ✅ | |
| VPN interfaces | ✅ (if listed) | ✅ | Expanded type list |
| Container on router | Usually no | ✅ (7.4+) | Hardware limits |

## Data flow
1. User/admin requests an internet change.
2. Previous mangle rules for that IP are removed.
3. On v7, missing tables are created with `fib`.
4. `mark-routing` rule is added; optional internal-network exception.
5. Table routes are applied from superadmin mapping (table↔interface) and detected gateways.

## Operational tips
- **Gateway:** read from dhcp-client (bound) then default routes. Dynamic VPN gateways may need manual routes on the router.
- **Internal exception:** accepts traffic to the API host’s /24 to avoid locking yourself out.
- **Table names:** prefer simple names without spaces (e.g. `to-vpn1`).
- **API rights:** do not use a read-only API user.

## Known limits
1. UI table creation is meaningful on v7 only.
2. “Block user” is stored in panel settings; real firewall block can be added later.
3. Single-router architecture in this version.
4. No built-in HTTPS — use a reverse proxy if needed.

## Production recommendations
- Mount `DATA_DIR` on durable storage (USB/SSD).
- Set env via `/container/envs` without rebuilding the image.
- Smoke-test on the same ROS version before production.
- Keep container logging enabled: `logging=yes`.
