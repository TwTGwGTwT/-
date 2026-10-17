package handlers

import (
		"errors"
		"html/template"
		"log"
		"net/http"
		"strconv"
		"strings"
		"time"

		"personal_expenses/internal/models"
		"personal_expenses/internal/repository"
)

var repo *repository.ExpenseRepository


// Передача в репозиторий
func Init(r * repository.ExpenseRepository) {
		repo = r
}

// Функция загрузки шаблонов
func loadTemplate(page string) (*template.Template, error) {
		return template.ParseFiles(
				"web/templates/layout.html",
				"web/templates/"+page,
		)
}

// Разбор и валидация формы
type expenseForm struct {
		Amount		float64
		Description	string
		Date 		string
}

func parseExpenseForm(r * http.Request) (expenseForm, error) {
		if err := r.ParseForm(); err != nil {
			return expenseForm{}, errors.New("не удалось прочитать форму")
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(r.FormValue("amount")), 64)
		if err != nil || amount <= 0 {
				return expenseForm{}, errors.New("некорректная сумма: должно быть число больше нуля")
		}

		description := strings.TrimSpace(r.FormValue("description"))
		if description == "" || len(description) > 200 {
				return expenseForm{}, errors.New("описание обязательно и не длиннее 200 символов")
		}

		date := r.FormValue("date")
		if _, err := time.Parse("2006-01-02", date); err != nil {
				return expenseForm{}, errors.New("некорректная дата")
		}

		return expenseForm{Amount: amount, Description: description, Date: date}, nil
}

func render(w http.ResponseWriter, page string, data any) {
		tmpl, err := loadTemplate(page)
		if err != nil {
				http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
				return
		}
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
				log.Println("ошибка рендера:", err)
		}
}

// Список трат
func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
		expenses, err := repo.GetAll()
		if err != nil {
				http.Error(w, "Не удалось получить список трат", http.StatusInternalServerError)
				return
		}
		render(w, "expenses.html", struct {
				Expenses []models.Expense
		}{Expenses: expenses})
}

// Показать форму добавления и обработать отправку
func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {			
		case http.MethodGet:
				render(w, "add.html", nil)

		case http.MethodPost:
				form, err := parseExpenseForm(r)
				if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
				}

				if _, err := repo.Create(models.Expense{
						Amount: 	 form.Amount,
						Description: form.Description,
						Date:		 form.Date,
				}); err != nil {
						log.Println("ошибка сохранения траты:", err)
						http.Error(w, "Не удалось сохранить трату", http.StatusInternalServerError)
						return
				}

				http.Redirect(w, r, "/expenses", http.StatusSeeOther)

		default:
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
}

// Показать форму редактирования и обработать отправку
func EditExpenseHandler(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
				http.Error(w, "Некорректный ID траты", http.StatusBadRequest)
				return
		}

		switch r.Method {
		case http.MethodGet:
				expense, err := repo.GetByID(id)
				if errors.Is(err, repository.ErrNotFound) {
						http.Error(w, "Трата не найдена", http.StatusNotFound)
						return
				}
				if err != nil {
						http.Error(w, "Не удалось получить трату", http.StatusInternalServerError)
						return
				}
				render(w, "edit.html", expense)

		case http.MethodPost:
				form, err := parseExpenseForm(r)
				if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
				}

				expense := models.Expense{
						ID:			 id,
						Amount:		 form.Amount,
						Description: form.Description,
						Date:		 form.Date,
				}
				if err := repo.Update(expense); err != nil {
						http.Error(w, "Не удалось обновить трату", http.StatusInternalServerError)
						return
				}

				http.Redirect(w, r, "/expenses", http.StatusSeeOther)

		default:
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
}

// Удаление траты
func DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
				return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
				http.Error(w, "Некорректный ID траты", http.StatusBadRequest)
				return
		}

		if err := repo.Delete(id); err != nil {
				http.Error(w, "Не удалось удалить трату", http.StatusInternalServerError)
				return
		}

		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}