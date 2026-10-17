package db

import (
		"database/sql"
		"log"
		"os"

		_ "modernc.org/sqlite"
)

// Открытие базы данных и накатывание миграции
func Open(dsn string) *sql.DB {
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
				log.Fatal("не удалось открыть базу данных:", err)
		}

		// Ограничение пула одним соединением из-за того, что SQLite плохо переносит много одновременных соединений
		db.SetMaxOpenConns(1)

		// Накатывание миграции
		migration, err := os.ReadFile("migrations/001_create_expenses.sql")
		if err != nil {
				log.Fatal("не удалось прочитать файл миграции:", err)
		}
		if _, err := db.Exec(string(migration)); err != nil {
				log.Fatal("не удалось выполнить миграцию:", err)
		}

		return db
}