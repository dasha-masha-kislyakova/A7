package common

import (
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB, dir string) error {
	list := []string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			list = append(list, path)
		}
		return nil
	})
	sort.Strings(list)
	for _, p := range list {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return err
		}
	}
	return nil
}
