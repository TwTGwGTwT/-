package handlers

import (
		"encoding/csv"
		"fmt"
		"log"
		"net/http"
		"strconv"
		"time"
)

// Экспорт трат текущего пользователя в CSV-файл
func ExportCSVHandler(w http.ResponseWriter, r *http.Request) {
		userID, ok := mustUserID(w, r)
		if !ok {
				return
		}

		// Учитывание фильтров, что на странице списка
		filter := parseFilters(r)
		expenses, err := repo.GetAll(filter, userID)
		if err != nil {
				log.Println("ошибка выборки трат для экспорта:", err)
				http.Error(w, "Не удалось получить траты для экспорта", http.StatusInternalServerError)
				return
		}

		filename := "expenses_" + time.Now().Format("2006-01-02") + ".csv"
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

		w.Write([]byte{0xEF, 0xBB, 0xBF})

		cw := csv.NewWriter(w)
		cw.Comma = ';'

		cw.Write([]string{"ID", "Сумма", "Описание", "Категория", "Дата"})
		for _, e := range expenses {
				cw.Write([]string{
						strconv.Itoa(e.ID),
						fmt.Sprintf("%.2f", e.Amount),
						e.Description,
						e.CategoryName,
						e.Date,
				})
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
				log.Println("ошибка записи CSV:", err)
		}
}