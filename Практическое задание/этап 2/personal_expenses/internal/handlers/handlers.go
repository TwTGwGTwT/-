package handlers

import (
		"html/template"
		"log"
		"net/http"
		"strconv"
		"strings"
		"sync"
		"time"

		"personal_expenses/internal/models"
)

// Хранилище трат в памяти
var (
		expenses []models.Expense
		nextID	 = 1
		mu		 sync.Mutex
)

// Функция загрузки шаблонов
func loadTemplate(page string) (*template.Template, error) {
		return template.ParseFiles(
				"web/templates/layout.html",
				"web/templates/"+page,
		)
}

// Список трат
func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
		tmpl, err := loadTemplate("expenses.html")
		if err != nil {
				http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
				return
		}

		mu.Lock()
		defer mu.Unlock()

		if err := tmpl.ExecuteTemplate(w, "layout", struct {
				Expenses []models.Expense
		}{Expenses: expenses}); err != nil {
				log.Println("ошибка рендера списка:", err)
		}
}

// Показать форму и обработать отправку
func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
				tmpl, err := loadTemplate("add.html")
				if err != nil {
						http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
						return
				}
				if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
						log.Println("ошибка рендера формы:", err)
				}
				return
		}

		if r.Method == http.MethodPost {
				// Обработка ошибки чтения формы
				if err := r.ParseForm(); err != nil {
						http.Error(w, "Не удалось прочитать форму", http.StatusBadRequest)
						return
				}

				// Валидация суммы
				amount, err := strconv.ParseFloat(strings.TrimSpace(r.FormValue("amount"), 64)
				if err != nil || amount <= 0 {
						http.Error(w, "Некорректная сумма: должно быть число больше нуля", http.StatusBadRequest)
						return
				}

				// Валидация описания
				description := strings.TrimSpace(r.FormValue("description"))
				if description == "" || len(description) > 200 {
						http.Error(w, "Описание обязательно и не длиннее 200 символов", http.StatusBadRequest)
						return
				}

				// Валидация даты
				date := r.FormValue("date")
				if _, err := time.Parse("2006-01-02", date); err != nil {
						http.Error(w, "Некорректная дата", http.StatusBadRequest)
						return
				}

				mu.Lock()
				expenses = append(expenses, models.Expense{
						ID:			 nextID,
						Amount:		 amount,
						Description: description,
						Date:		 date,
				})
				nextID++
				mu.Unlock()

				// Редирект на список
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
				return
		}

		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}