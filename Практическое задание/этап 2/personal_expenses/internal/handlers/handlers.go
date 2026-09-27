package handlers

import (
		"personal_expenses/internal/models"
		"html/template"
		"net/http"
		"strconv"
		"sync"
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
				http.Error(w, "Ошибка шаблона", 500)
				return
		}

		mu.Lock()
		defer mu.Unlock()

		tmpl.ExecuteTemplate(w, "layout", struct {
				Expenses []models.Expense
		}{Expenses: expenses})
}

// Показать форму и обработать отправку
func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
				tmpl, err := loadTemplate("add.html")
				if err != nil {
						http.Error(w, "Ошибка шаблона", 500)
						return
				}
				tmpl.ExecuteTemplate(w, "layout", nil)
				return
		}

		if r.Method == http.MethodPost {
				r.ParseForm()

				amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
				if err != nil {
						http.Error(w, "Некорректная сумма", 400)
						return
				}

				description := r.FormValue("description")
				date := r.FormValue("date")

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

		http.Error(w, "Метод не поддерживается", 405)
}