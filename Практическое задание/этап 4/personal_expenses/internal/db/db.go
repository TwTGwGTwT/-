package db

import (
		"database/sql"
		"log"
		"os"
		"sort"
		"strings"

		_ "modernc.org/sqlite"
)

// Открытие базы данных и накатывание миграций
func Open(dsn string) *sql.DB {
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
				log.Fatal("не удалось открыть базу данных:", err)
		}

		// Ограничение пула одним соединением из-за того, что SQLite плохо переносит много одновременных соединений
		db.SetMaxOpenConns(1)

		// Проверка внешних ключей
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
				log.Fatal("не удалось включить внешние ключи:", err)
		}

		// Накатывание всех SQL-файлов
		entries, err := os.ReadDir("migrations")
		if err != nil {
				log.Fatal("не удалось прочитать папку миграций:", err)
		}

		names := make([]string, 0, len(entries))
		for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".sql") {
						names = append(names, e.Name())
				}
		}
		sort.Strings(names)

		for _, name := range names {
			migration, err := os.ReadFile("migrations/" + name)
			if err != nil {
					log.Fatal("не удалось прочитать миграцию:", err)
			}
			if _, err := db.Exec(string(migration)); err != nil {
					log.Fatal("не удалось выполнить миграцию "+name+":", err)
			}
		}	

		return db
}