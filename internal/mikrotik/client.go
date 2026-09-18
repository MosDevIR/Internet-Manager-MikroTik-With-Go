package mikrotik

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-routeros/routeros/v3"
	"internet-manager-go/internal/models"
)

// انواع اینترفیس پشتیبانی‌شده (شامل VPN)
var supportedIfaceTypes = map[string]bool{
	"ether": true, "wlan": true, "pppoe-out": true, "lte": true, "bridge": true,
	"wg": true, "wireguard": true,
	"ovpn-out": true, "ovpn-in": true,
	"l2tp-out": true, "l2tp-in": true,
	"sstp-out": true, "sstp-in": true,
	"pptp-out": true, "pptp-in": true,
	"ipip": true, "gre": true, "eoip": true,
	"vti": true, "vlan": true,
}

type Client struct {
	host string
	user string
	pass string
	port string

	rosCache *models.ROSInfo
	rosMu    sync.Mutex

	// کش کوتاه‌مدت لاگ برای کاهش فشار روی API در حالت لایو
	logCache     []map[string]string
	logCacheKey  string
	logCacheAt   time.Time
	logMu        sync.Mutex
}

func New(host, user, pass, port string) *Client {
	return &Client{host: host, user: user, pass: pass, port: port}
}

func (c *Client) connect() (*routeros.Client, error) {
	addr := net.JoinHostPort(c.host, c.port)
	return routeros.DialTimeout(addr, c.user, c.pass, 8*time.Second)
}

// DetectROS نسخه RouterOS را تشخیص می‌دهد و کش می‌کند
func (c *Client) DetectROS() (*models.ROSInfo, error) {
	c.rosMu.Lock()
	defer c.rosMu.Unlock()
	if c.rosCache != nil {
		return c.rosCache, nil
	}

	cli, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	info := &models.ROSInfo{Major: 6, Version: "6.x", IsV7: false}
	reply, err := cli.Run("/system/resource/print")
	if err == nil && len(reply.Re) > 0 {
		ver := reply.Re[0].Map["version"]
		info.Version = ver
		parts := strings.Split(ver, ".")
		if len(parts) > 0 {
			if maj, e := strconv.Atoi(parts[0]); e == nil {
				info.Major = maj
				info.IsV7 = maj >= 7
			}
		}
	}
	c.rosCache = info
	log.Printf("Detected RouterOS %s (v7=%v)\n", info.Version, info.IsV7)
	return info, nil
}

// EnsureTable در v7 جدول را با fib می‌سازد (در v6 کاری نمی‌کند)
func (c *Client) EnsureTable(name string, ros *models.ROSInfo) error {
	if name == "" || name == "main" || !ros.IsV7 {
		return nil
	}
	cli, err := c.connect()
	if err != nil {
		return err
	}
	defer cli.Close()

	reply, err := cli.Run("/routing/table/print", "?name="+name)
	if err == nil && len(reply.Re) > 0 {
		return nil
	}
	_, err = cli.Run("/routing/table/add", "=name="+name, "=fib=")
	return err
}

// ManageMangle قوانین mangle کاربر را تنظیم می‌کند
func (c *Client) ManageMangle(ip, mark string, ros *models.ROSInfo) error {
	cli, err := c.connect()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, comment := range []string{"user:" + ip, "EXCEPTION: " + ip} {
		reply, e := cli.Run("/ip/firewall/mangle/print", "?comment="+comment)
		if e == nil {
			for _, re := range reply.Re {
				_, _ = cli.Run("/ip/firewall/mangle/remove", "=numbers="+re.Map[".id"])
			}
		}
	}

	if mark == "" || mark == "main" {
		return nil
	}

	if err := c.EnsureTable(mark, ros); err != nil {
		return fmt.Errorf("ساخت جدول روت: %w", err)
	}

	internal := internalNetwork(c.host)
	if internal != "" {
		_, _ = cli.Run("/ip/firewall/mangle/add",
			"=chain=prerouting",
			"=src-address="+ip,
			"=dst-address="+internal,
			"=action=accept",
			"=comment=EXCEPTION: "+ip,
		)
	}

	_, err = cli.Run("/ip/firewall/mangle/add",
		"=chain=prerouting",
		"=src-address="+ip,
		"=action=mark-routing",
		"=new-routing-mark="+mark,
		"=passthrough=yes",
		"=comment=user:"+ip,
	)
	return err
}

func (c *Client) GetUserMark(ip string) string {
	cli, err := c.connect()
	if err != nil {
		return "main"
	}
	defer cli.Close()

	reply, err := cli.Run("/ip/firewall/mangle/print", "?comment=user:"+ip)
	if err != nil || len(reply.Re) == 0 {
		return "main"
	}
	if m := reply.Re[0].Map["new-routing-mark"]; m != "" {
		return m
	}
	return "main"
}

func (c *Client) GetDHCPLeases() []map[string]string {
	cli, err := c.connect()
	if err != nil {
		return nil
	}
	defer cli.Close()

	reply, err := cli.Run("/ip/dhcp-server/lease/print")
	if err != nil {
		return []map[string]string{}
	}
	res := make([]map[string]string, 0)
	for _, re := range reply.Re {
		if re.Map["address"] != "" {
			res = append(res, re.Map)
		}
	}
	return res
}

// GetRawTableNames همه جدول‌های موجود را از منابع مختلف جمع می‌کند
func (c *Client) GetRawTableNames(ros *models.ROSInfo) []string {
	cli, err := c.connect()
	if err != nil {
		return []string{"main"}
	}
	defer cli.Close()

	seen := map[string]bool{"main": true}
	res := []string{"main"}

	if ros.IsV7 {
		reply, e := cli.Run("/routing/table/print")
		if e == nil {
			for _, re := range reply.Re {
				if n := re.Map["name"]; n != "" && !seen[n] {
					seen[n] = true
					res = append(res, n)
				}
			}
		}
	}

	reply, e := cli.Run("/ip/firewall/mangle/print", "?action=mark-routing")
	if e == nil {
		for _, re := range reply.Re {
			if m := re.Map["new-routing-mark"]; m != "" && !seen[m] {
				seen[m] = true
				res = append(res, m)
			}
		}
	}

	reply, e = cli.Run("/ip/route/print")
	if e == nil {
		for _, re := range reply.Re {
			for _, key := range []string{"routing-table", "routing-mark"} {
				if t := re.Map[key]; t != "" && !seen[t] {
					seen[t] = true
					res = append(res, t)
				}
			}
		}
	}
	return res
}

func (c *Client) GetTables(s models.Settings, ros *models.ROSInfo) []models.TableItem {
	raw := c.GetRawTableNames(ros)
	res := make([]models.TableItem, 0, len(raw))
	for _, t := range raw {
		name := t
		if n, ok := s.RoutingTables[t]; ok && n != "" {
			name = n
		}
		res = append(res, models.TableItem{ID: t, Name: name})
	}
	return res
}

func (c *Client) GetRawInterfaces() []string {
	cli, err := c.connect()
	if err != nil {
		return nil
	}
	defer cli.Close()

	reply, err := cli.Run("/interface/print")
	if err != nil {
		return nil
	}
	var res []string
	for _, re := range reply.Re {
		t := re.Map["type"]
		name := re.Map["name"]
		if name != "" && (supportedIfaceTypes[t] || strings.HasPrefix(t, "wg") || strings.Contains(t, "vpn")) {
			res = append(res, name)
		}
	}
	return res
}

func (c *Client) GetInterfaces(s models.Settings) map[string]string {
	raw := c.GetRawInterfaces()
	res := make(map[string]string)
	for _, i := range raw {
		name := i
		if n, ok := s.Interfaces[i]; ok && n != "" {
			name = n
		}
		res[i] = name
	}
	return res
}

func (c *Client) GetInterfaceGateways() map[string]string {
	cli, err := c.connect()
	if err != nil {
		return nil
	}
	defer cli.Close()

	res := make(map[string]string)
	reply, _ := cli.Run("/ip/dhcp-client/print")
	for _, re := range reply.Re {
		if re.Map["status"] == "bound" {
			if iface, gw := re.Map["interface"], re.Map["gateway"]; iface != "" && gw != "" {
				res[iface] = gw
			}
		}
	}
	reply, _ = cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0")
	for _, re := range reply.Re {
		iface, gw := re.Map["interface"], re.Map["gateway"]
		if iface != "" && gw != "" {
			if _, exists := res[iface]; !exists {
				res[iface] = gw
			}
		}
	}
	return res
}

func (c *Client) addRoute(cli *routeros.Client, dst, gateway, table string, ros *models.ROSInfo) error {
	args := []string{
		"=dst-address=" + dst,
		"=gateway=" + gateway,
		"=check-gateway=ping",
		"=comment=panel-route",
	}
	if table != "" && table != "main" {
		if ros.IsV7 {
			args = append(args, "=routing-table="+table)
		} else {
			args = append(args, "=routing-mark="+table)
		}
	}
	full := append([]string{"/ip/route/add"}, args...)
	_, err := cli.RunArgs(full)
	return err
}

func (c *Client) removeDefaultRoutes(cli *routeros.Client, table string, ros *models.ROSInfo) {
	var reply *routeros.Reply
	var err error
	if table == "" || table == "main" {
		reply, err = cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0")
	} else if ros.IsV7 {
		reply, err = cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0", "?routing-table="+table)
	} else {
		reply, err = cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0", "?routing-mark="+table)
	}
	if err != nil {
		return
	}
	for _, re := range reply.Re {
		_, _ = cli.Run("/ip/route/remove", "=numbers="+re.Map[".id"])
	}
}

func (c *Client) SetDefaultRoute(iface string, ros *models.ROSInfo) error {
	gws := c.GetInterfaceGateways()
	gw, ok := gws[iface]
	if !ok || gw == "" {
		return fmt.Errorf("گیت‌وی برای اینترفیس %s یافت نشد", iface)
	}
	cli, err := c.connect()
	if err != nil {
		return err
	}
	defer cli.Close()
	c.removeDefaultRoutes(cli, "main", ros)
	return c.addRoute(cli, "0.0.0.0/0", gw, "main", ros)
}

// GetCurrentDefaultIface اینترفیس و گیت‌وی روت پیش‌فرض جدول main را برمی‌گرداند
func (c *Client) GetCurrentDefaultIface() (iface, gateway string) {
	cli, err := c.connect()
	if err != nil {
		return "", ""
	}
	defer cli.Close()

	reply, err := cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0")
	if err != nil {
		return "", ""
	}

	for _, re := range reply.Re {
		table := re.Map["routing-table"]
		if table != "" && table != "main" {
			continue
		}
		if re.Map["active"] == "false" {
			continue
		}

		gw := re.Map["gateway"]
		ifc := re.Map["interface"]

		if ifc != "" {
			return ifc, gw
		}

		if gw != "" {
			dhcp, _ := cli.Run("/ip/dhcp-client/print")
			for _, d := range dhcp.Re {
				if d.Map["gateway"] == gw && d.Map["status"] == "bound" {
					return d.Map["interface"], gw
				}
			}
			routes, _ := cli.Run("/ip/route/print", "?gateway="+gw)
			for _, r := range routes.Re {
				if r.Map["interface"] != "" {
					return r.Map["interface"], gw
				}
			}
			return "", gw
		}
	}
	return "", ""
}

// GetDetectedTableInterfaces از روت‌های واقعی روتر نگاشت جدول → اینترفیس را تشخیص می‌دهد
func (c *Client) GetDetectedTableInterfaces(ros *models.ROSInfo) map[string]string {
	cli, err := c.connect()
	if err != nil {
		return nil
	}
	defer cli.Close()

	// gateway → interface (از dhcp-client و روت‌ها)
	gwToIface := make(map[string]string)
	gws := c.GetInterfaceGateways()
	for iface, gw := range gws {
		if gw != "" {
			gwToIface[gw] = iface
		}
	}

	res := make(map[string]string)
	reply, err := cli.Run("/ip/route/print", "?dst-address=0.0.0.0/0")
	if err != nil {
		return res
	}

	for _, re := range reply.Re {
		table := re.Map["routing-table"]
		if table == "" {
			table = re.Map["routing-mark"]
		}
		if table == "" || table == "main" {
			continue
		}

		ifc := re.Map["interface"]
		gw := re.Map["gateway"]

		if ifc != "" {
			res[table] = ifc
			continue
		}
		if gw != "" {
			if iface, ok := gwToIface[gw]; ok {
				res[table] = iface
			}
		}
	}
	return res
}

func (c *Client) ApplyTableRoutes(s models.Settings, ros *models.ROSInfo) {
	gws := c.GetInterfaceGateways()
	cli, err := c.connect()
	if err != nil {
		return
	}
	defer cli.Close()

	for table, iface := range s.TableInterfaceMap {
		gw, ok := gws[iface]
		if !ok || gw == "" {
			continue
		}
		_ = c.EnsureTable(table, ros)
		c.removeDefaultRoutes(cli, table, ros)
		_ = c.addRoute(cli, "0.0.0.0/0", gw, table, ros)
	}
}

func (c *Client) GetUserStatusList(s models.Settings) []models.UserStatus {
	leases := c.GetDHCPLeases()
	blocked := make(map[string]bool)
	for _, ip := range s.BlockedIPs {
		blocked[ip] = true
	}
	res := make([]models.UserStatus, 0)
	for _, lease := range leases {
		ip := lease["address"]
		if ip == "" {
			continue
		}
		res = append(res, models.UserStatus{
			IP:      ip,
			MAC:     lease["mac-address"],
			Host:    lease["host-name"],
			Mark:    c.GetUserMark(ip),
			Blocked: blocked[ip],
			Label:   s.UserLabels[ip],
		})
	}
	return res
}

func internalNetwork(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return ""
	}
	return ip.Mask(net.CIDRMask(24, 32)).String() + "/24"
}

// GetLogs آخرین لاگ‌های سیستم را برمی‌گرداند (فقط خواندن)
// topicFilter و search اختیاری هستند.
func (c *Client) GetLogs(limit int, topicFilter, search string) []map[string]string {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	topicFilter = strings.ToLower(strings.TrimSpace(topicFilter))
	search = strings.ToLower(strings.TrimSpace(search))
	cacheKey := topicFilter + "|" + search + "|" + strconv.Itoa(limit)

	// کش ۲ ثانیه‌ای: درخواست‌های لایو تکراری API را نمی‌زنند
	c.logMu.Lock()
	if c.logCache != nil && c.logCacheKey == cacheKey && time.Since(c.logCacheAt) < 2*time.Second {
		out := make([]map[string]string, len(c.logCache))
		copy(out, c.logCache)
		c.logMu.Unlock()
		return out
	}
	c.logMu.Unlock()

	cli, err := c.connect()
	if err != nil {
		return nil
	}
	defer cli.Close()

	reply, err := cli.Run("/log/print")
	if err != nil {
		return nil
	}

	var filtered []map[string]string
	for _, re := range reply.Re {
		topics := re.Map["topics"]
		message := re.Map["message"]
		if topicFilter != "" && !strings.Contains(strings.ToLower(topics), topicFilter) {
			continue
		}
		if search != "" {
			hay := strings.ToLower(topics + " " + message)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		filtered = append(filtered, map[string]string{
			"time":    re.Map["time"],
			"topics":  topics,
			"message": message,
		})
	}

	start := 0
	if len(filtered) > limit {
		start = len(filtered) - limit
	}
	res := filtered[start:]
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}

	c.logMu.Lock()
	c.logCache = res
	c.logCacheKey = cacheKey
	c.logCacheAt = time.Now()
	c.logMu.Unlock()
	return res
}
