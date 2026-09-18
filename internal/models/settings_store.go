package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	path    string
	dataDir string
}

func NewStore(path string) *Store {
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)
	return &Store{path: path, dataDir: dir}
}

func (s *Store) Load() Settings {
	st := NewSettings()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return st
	}
	_ = json.Unmarshal(data, &st)
	if st.RoutingTables == nil {
		st.RoutingTables = make(map[string]string)
	}
	if st.Interfaces == nil {
		st.Interfaces = make(map[string]string)
	}
	if st.TableInterfaceMap == nil {
		st.TableInterfaceMap = make(map[string]string)
	}
	if st.UserLabels == nil {
		st.UserLabels = make(map[string]string)
	}
	if st.BlockedIPs == nil {
		st.BlockedIPs = []string{}
	}
	return st
}

func (s *Store) Save(st Settings) error {
	_ = os.MkdirAll(s.dataDir, 0755)
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}
	// بکاپ چرخشی ساده در همان پوشه داده
	backup := filepath.Join(s.dataDir, "settings-backup-"+time.Now().Format("20060102-150405")+".json")
	_ = os.WriteFile(backup, data, 0644)
	// نگه داشتن حداکثر ۵ بکاپ آخر
	s.pruneBackups(5)
	return nil
}

func (s *Store) pruneBackups(keep int) {
	matches, err := filepath.Glob(filepath.Join(s.dataDir, "settings-backup-*.json"))
	if err != nil || len(matches) <= keep {
		return
	}
	// ساده‌ترین روش: حذف قدیمی‌ترین‌ها (مرتب‌سازی نام بر اساس timestamp)
	for i := 0; i < len(matches)-keep; i++ {
		_ = os.Remove(matches[i])
	}
}

func AppendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func RemoveItem(slice []string, item string) []string {
	var res []string
	for _, s := range slice {
		if s != item {
			res = append(res, s)
		}
	}
	return res
}
