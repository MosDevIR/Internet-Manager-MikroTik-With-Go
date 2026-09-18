package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"internet-manager-go/internal/config"
	"internet-manager-go/internal/i18n"
	"internet-manager-go/internal/mikrotik"
	"internet-manager-go/internal/models"
)

type App struct {
	Cfg   *config.Config
	MT    *mikrotik.Client
	Store *models.Store
	Tmpl  *template.Template

	sessions map[string]*models.Session
	sessMu   sync.RWMutex
}

func New(cfg *config.Config, mt *mikrotik.Client, store *models.Store, tmpl *template.Template) *App {
	return &App{
		Cfg:      cfg,
		MT:       mt,
		Store:    store,
		Tmpl:     tmpl,
		sessions: make(map[string]*models.Session),
	}
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	if err := a.Tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Println("template error:", err)
		http.Error(w, "خطای داخلی سرور", 500)
	}
}

func (a *App) getCookie(r *http.Request) string {
	c, err := r.Cookie("sid")
	if err != nil {
		return ""
	}
	return c.Value
}

func (a *App) requireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sid := a.getCookie(r)
			a.sessMu.RLock()
			s, ok := a.sessions[sid]
			a.sessMu.RUnlock()
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			for _, role := range roles {
				if s.Role == role {
					next(w, r)
					return
				}
			}
			http.Error(w, i18n.T(i18n.FromRequest(r), "forbidden"), http.StatusForbidden)
		}
	}
}


func (a *App) baseData(r *http.Request) map[string]any {
	lang := i18n.FromRequest(r)
	return map[string]any{
		"Lang":     string(lang),
		"HTMLLang": i18n.HTMLLang(lang),
		"Dir":      i18n.Dir(lang),
		"T":        i18n.Dict(lang),
		"IsRTL":    i18n.IsRTL(lang),
	}
}

func merge(base map[string]any, extra map[string]any) map[string]any {
	for k, v := range extra {
		base[k] = v
	}
	return base
}

func (a *App) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	mux.HandleFunc("/login", a.login)
	mux.HandleFunc("/about", a.aboutPanel)
	mux.HandleFunc("/lang", a.setLang)
	mux.HandleFunc("/logout", a.logout)
	mux.HandleFunc("/user", a.requireRole("user", "admin", "superadmin")(a.userPanel))
	mux.HandleFunc("/admin", a.requireRole("admin", "superadmin")(a.adminPanel))
	mux.HandleFunc("/settings", a.requireRole("superadmin")(a.settingsPanel))
	mux.HandleFunc("/logs", a.requireRole("admin", "superadmin")(a.logsPanel))
	mux.HandleFunc("/api/logs", a.requireRole("admin", "superadmin")(a.logsAPI))
}

// ---------- Auth ----------

func (a *App) setLang(w http.ResponseWriter, r *http.Request) {
	lang := i18n.Parse(r.URL.Query().Get("lang"))
	i18n.SetCookie(w, lang)
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/login"
	}
	http.Redirect(w, r, ref, http.StatusFound)
}

func (a *App) aboutPanel(w http.ResponseWriter, r *http.Request) {
	a.render(w, "about.html", a.baseData(r))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pass := r.FormValue("password")
		role := ""
		switch pass {
		case a.Cfg.SuperAdminPassword:
			role = "superadmin"
		case a.Cfg.AdminPassword:
			role = "admin"
		case a.Cfg.UserPassword:
			role = "user"
		}
		if role != "" {
			sid := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000)
			a.sessMu.Lock()
			a.sessions[sid] = &models.Session{Role: role}
			a.sessMu.Unlock()
			http.SetCookie(w, &http.Cookie{
				Name: "sid", Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode,
			})
			switch role {
			case "superadmin":
				http.Redirect(w, r, "/settings", http.StatusFound)
			case "admin":
				http.Redirect(w, r, "/admin", http.StatusFound)
			default:
				http.Redirect(w, r, "/user", http.StatusFound)
			}
			return
		}
		{
			d := a.baseData(r)
			d["Error"] = i18n.T(i18n.FromRequest(r), "wrong_password")
			a.render(w, "login.html", d)
		}
		return
	}
	a.render(w, "login.html", a.baseData(r))
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	sid := a.getCookie(r)
	a.sessMu.Lock()
	delete(a.sessions, sid)
	a.sessMu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// ---------- User: فقط اینترنت خودش ----------

func (a *App) userPanel(w http.ResponseWriter, r *http.Request) {
	clientIP := clientIP(r)
	settings := a.Store.Load()

	ros, err := a.MT.DetectROS()
	if err != nil {
		a.render(w, "user.html", merge(a.baseData(r), map[string]any{"Error": i18n.T(i18n.FromRequest(r), "conn_failed") + ": " + err.Error(), "IP": clientIP}))
		return
	}

	if r.Method == http.MethodPost {
		newMark := r.FormValue("routing_mark")
		if err := a.MT.ManageMangle(clientIP, newMark, ros); err != nil {
			a.render(w, "user.html", merge(a.baseData(r), map[string]any{"Error": err.Error(), "IP": clientIP}))
			return
		}
		http.Redirect(w, r, "/user?ok=1", http.StatusFound)
		return
	}

	tables := a.MT.GetTables(settings, ros)
	current := a.MT.GetUserMark(clientIP)
	currentName := current
	for _, t := range tables {
		if t.ID == current {
			currentName = t.Name
			break
		}
	}

	data := merge(a.baseData(r), map[string]any{
		"IP": clientIP, "Tables": tables, "Current": current,
		"CurrentName": currentName, "ROSVersion": ros.Version,
	})
	if r.URL.Query().Get("ok") == "1" {
		data["Success"] = i18n.T(i18n.FromRequest(r), "success_changed")
	}
	a.render(w, "user.html", data)
}

// ---------- Admin: مدیریت همه کاربران ----------

func (a *App) adminPanel(w http.ResponseWriter, r *http.Request) {
	settings := a.Store.Load()
	ros, err := a.MT.DetectROS()
	if err != nil {
		a.render(w, "admin.html", merge(a.baseData(r), map[string]any{
			"Error": i18n.T(i18n.FromRequest(r), "conn_failed") + ": " + err.Error(),
			"Users": []models.UserStatus{}, "Tables": []models.TableItem{}, "Leases": []map[string]string{},
			"Interfaces": map[string]string{}, "Gateways": map[string]string{},
			"CustomCount": 0, "BlockedCount": 0,
		}))
		return
	}

	if r.Method == http.MethodPost {
		ip := strings.TrimSpace(r.FormValue("client_ip"))
		msg := ""
		ok := true

		switch {
		case r.FormValue("change_internet") != "":
			if e := a.MT.ManageMangle(ip, r.FormValue("new_internet"), ros); e != nil {
				ok, msg = false, e.Error()
			} else {
				msg = "اینترنت کاربر تغییر کرد"
			}
		case r.FormValue("remove_internet") != "":
			if e := a.MT.ManageMangle(ip, "", ros); e != nil {
				ok, msg = false, e.Error()
			} else {
				msg = "بازگشت به پیش‌فرض انجام شد"
			}
		case r.FormValue("block_ip") != "":
			settings.BlockedIPs = models.AppendUnique(settings.BlockedIPs, ip)
			_ = a.Store.Save(settings)
			msg = "کاربر مسدود شد"
		case r.FormValue("unblock_ip") != "":
			settings.BlockedIPs = models.RemoveItem(settings.BlockedIPs, ip)
			_ = a.Store.Save(settings)
			msg = "کاربر فعال شد"
		case r.FormValue("save_label") != "":
			settings.UserLabels[ip] = strings.TrimSpace(r.FormValue("new_label"))
			_ = a.Store.Save(settings)
			msg = "توضیح ذخیره شد"
		case r.FormValue("change_default") != "":
			iface := r.FormValue("default_iface")
			if e := a.MT.SetDefaultRoute(iface, ros); e != nil {
				ok, msg = false, e.Error()
			} else {
				msg = "روت پیش‌فرض به‌روز شد"
			}
		}

		if ok {
			http.Redirect(w, r, "/admin?ok="+msg, http.StatusFound)
		} else {
			a.render(w, "admin.html", merge(a.baseData(r), map[string]any{
				"Error": msg, "ROSVersion": ros.Version,
				"Users": []models.UserStatus{}, "Tables": []models.TableItem{}, "Leases": []map[string]string{},
				"Interfaces": map[string]string{}, "Gateways": map[string]string{},
				"CustomCount": 0, "BlockedCount": 0,
			}))
		}
		return
	}

	users := a.MT.GetUserStatusList(settings)
	tables := a.MT.GetTables(settings, ros)
	leases := a.MT.GetDHCPLeases()

	custom, blocked := 0, 0
	for _, u := range users {
		if u.Mark != "main" {
			custom++
		}
		if u.Blocked {
			blocked++
		}
	}

	ifaces := a.MT.GetInterfaces(settings)
	gw := a.MT.GetInterfaceGateways()

	data := merge(a.baseData(r), map[string]any{
		"Users": users, "Tables": tables, "Leases": leases,
		"Interfaces": ifaces, "Gateways": gw,
		"CustomCount": custom, "BlockedCount": blocked,
		"ROSVersion": ros.Version,
	})
	if m := r.URL.Query().Get("ok"); m != "" {
		data["Success"] = m
	}
	a.render(w, "admin.html", data)
}

// ---------- SuperAdmin: تعاریف، جدول‌ها، نگاشت ----------

func (a *App) settingsPanel(w http.ResponseWriter, r *http.Request) {
	settings := a.Store.Load()
	ros, err := a.MT.DetectROS()
	if err != nil {
		a.render(w, "settings.html", merge(a.baseData(r), map[string]any{"Error": i18n.T(i18n.FromRequest(r), "conn_failed") + ": " + err.Error()}))
		return
	}

	if r.Method == http.MethodPost {
		msg := "تنظیمات ذخیره شد"

		// نام اینترفیس‌ها
		for _, iface := range a.MT.GetRawInterfaces() {
			name := strings.TrimSpace(r.FormValue("iface_" + iface))
			if name != "" {
				settings.Interfaces[iface] = name
			} else {
				delete(settings.Interfaces, iface)
			}
		}

		// نام جدول‌ها
		for _, t := range a.MT.GetRawTableNames(ros) {
			name := strings.TrimSpace(r.FormValue("table_" + t))
			if name != "" {
				settings.RoutingTables[t] = name
			} else {
				delete(settings.RoutingTables, t)
			}
		}

		// ساخت جدول جدید (v7)
		if r.FormValue("create_table") != "" {
			name := strings.TrimSpace(r.FormValue("new_table_name"))
			if name != "" && name != "main" {
				if e := a.MT.EnsureTable(name, ros); e != nil {
					a.render(w, "settings.html", merge(a.baseData(r), map[string]any{"Error": e.Error(), "ROSVersion": ros.Version}))
					return
				}
				msg = "جدول ساخته شد: " + name
			}
		}

		// نگاشت جدول ↔ اینترفیس
		if r.FormValue("update_table_interfaces") != "" {
			newMap := make(map[string]string)
			for _, table := range a.MT.GetRawTableNames(ros) {
				if table == "main" {
					continue
				}
				if val := r.FormValue("interface_for_" + table); val != "" {
					newMap[table] = val
				}
			}
			settings.TableInterfaceMap = newMap
			a.MT.ApplyTableRoutes(settings, ros)
			msg = "ارتباط جدول‌ها ذخیره شد"
		}

		_ = a.Store.Save(settings)
		http.Redirect(w, r, "/settings?ok="+msg, http.StatusFound)
		return
	}

	tables := a.MT.GetTables(settings, ros)
	ifaces := a.MT.GetInterfaces(settings)
	gateways := a.MT.GetInterfaceGateways()

	data := merge(a.baseData(r), map[string]any{
		"Interfaces": ifaces, "Tables": tables, "Settings": settings,
		"Gateways": gateways, "ROSVersion": ros.Version, "IsV7": ros.IsV7,
	})
	if m := r.URL.Query().Get("ok"); m != "" {
		data["Success"] = m
	}
	a.render(w, "settings.html", data)
}

// ---------- Logs: فقط مشاهده برای ادمین و سوپرادمین ----------

func parseLogQuery(r *http.Request) (limit int, topic, search string) {
	limit = 100
	if v := r.URL.Query().Get("limit"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	topic = strings.TrimSpace(r.URL.Query().Get("topic"))
	search = strings.TrimSpace(r.URL.Query().Get("q"))
	return
}

func (a *App) logsPanel(w http.ResponseWriter, r *http.Request) {
	limit, topic, search := parseLogQuery(r)
	ros, _ := a.MT.DetectROS()
	ver := ""
	if ros != nil {
		ver = ros.Version
	}
	logs := a.MT.GetLogs(limit, topic, search)
	live := r.URL.Query().Get("live") == "1"
	a.render(w, "logs.html", merge(a.baseData(r), map[string]any{
		"Logs":       logs,
		"Limit":      limit,
		"Topic":      topic,
		"Search":     search,
		"Live":       live,
		"ROSVersion": ver,
		"Count":      len(logs),
	}))
}

func (a *App) logsAPI(w http.ResponseWriter, r *http.Request) {
	limit, topic, search := parseLogQuery(r)
	logs := a.MT.GetLogs(limit, topic, search)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count": len(logs),
		"logs":  logs,
	})
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return x
	}
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
