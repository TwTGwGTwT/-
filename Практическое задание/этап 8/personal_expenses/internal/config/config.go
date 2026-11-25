package config

import (
		"bufio"
		"os"
		"strings"
)

type Config struct {
		Port   string
		DBPath string
		Secret string
}

// Читает .env, переменные окружения, значения по умолчанию
func Load() Config {
		loadDotEnv(".env")

		return Config{
				Port:   getenv("PORT", "8080"),
				DBPath: getenv("DB_PATH", "expenses.db"),
				Secret: getenv("SECRET_KEY", ""),
		}
}

// Разбирает файл вида KEY=value и выставляет переменные окружения
func loadDotEnv(path string) {
		f, err := os.Open(path)
		if err != nil {
				return
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line == "" || strings.HasPrefix(line, "#") {
						continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
						continue
				}
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				if os.Getenv(key) == "" {
						os.Setenv(key, val)
				}
		}
}

func getenv(key, def string) string {
		if v := os.Getenv(key); v != "" {
				return v
		}
		return def
}