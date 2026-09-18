package models

// Settings ذخیره‌سازی محلی تنظیمات پنل
type Settings struct {
	RoutingTables     map[string]string `json:"routing_tables"`      // id → نام نمایشی
	Interfaces        map[string]string `json:"interfaces"`         // name → نام نمایشی
	TableInterfaceMap map[string]string `json:"table_interface_map"` // table → interface
	BlockedIPs        []string          `json:"blocked_ips"`
	UserLabels        map[string]string `json:"user_labels"` // ip → توضیح
}

func NewSettings() Settings {
	return Settings{
		RoutingTables:     make(map[string]string),
		Interfaces:        make(map[string]string),
		TableInterfaceMap: make(map[string]string),
		BlockedIPs:        []string{},
		UserLabels:        make(map[string]string),
	}
}

// Session اطلاعات نشست کاربر وب
type Session struct {
	Role string // user | admin | superadmin
}

// ROSInfo اطلاعات نسخه RouterOS
type ROSInfo struct {
	Major   int
	Version string
	IsV7    bool
}

// TableItem یک جدول روت با نام نمایشی
type TableItem struct {
	ID   string
	Name string
}

// UserStatus وضعیت یک کاربر در پنل ادمین
type UserStatus struct {
	IP      string
	MAC     string
	Host    string
	Mark    string
	Blocked bool
	Label   string
}
