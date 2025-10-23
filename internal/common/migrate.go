package common

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB, dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, filepath.Join(dir, file.Name()))
		}
	}

	sort.Strings(migrations)

	for _, file := range migrations {
		sql, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return err
		}
	}

	return nil
}
