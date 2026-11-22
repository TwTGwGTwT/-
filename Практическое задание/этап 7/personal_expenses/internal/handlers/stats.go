package handlers

import (
		"encoding/json"
		"net/http"
		"strconv"
)

// Универсальная отправка JSON
func writeJSON(w http.ResponseWriter, status int, data any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(data); err != nil {
				http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		}
}

// Страница статистики с графиками
func StatsPageHandler(w http.ResponseWriter, r *http.Request) {
		if _, ok := mustUserID(w, r); !ok {
				return
		}
		render(w, r, "stats.html", nil)
}

func StatsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
		userID, ok := mustUserID(w, r)
		if !ok {
				return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		stats, err := repo.GetByCategory(userID, from, to)
		if err != nil {
				http.Error(w, "Не удалось получить статистику", http.StatusInternalServerError)
				return
		}
		writeJSON(w, http.StatusOK, stats)
}

func StatsByMonthHandler(w http.ResponseWriter, r *http.Request) {
		userID, ok := mustUserID(w, r)
		if !ok {
				return
		}

		year := r.URL.Query().Get("year")
		if year != "" {
				if _, err := strconv.Atoi(year); err != nil || len(year) != 4 {
						http.Error(w, "Некорректный год", http.StatusBadRequest)
						return
				}
		}

		stats, err := repo.GetByMonth(userID, year)
		if err != nil {
				http.Error(w, "Не удалось получить статистику", http.StatusInternalServerError)
				return
		}
		writeJSON(w, http.StatusOK, stats)
}