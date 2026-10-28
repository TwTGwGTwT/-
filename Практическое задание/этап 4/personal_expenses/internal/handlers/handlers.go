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
		CategoryID	int
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

		categoryID, err := strconv.Atoi(r.FormValue("category_id"))
		if err != nil || categoryID <= 0 {
				return expenseForm{}, errors.New("выберите категорию")
		}

		return expenseForm{
			Amount:		 amount,
			Description: description,
			Date: 		 date,
			CategoryID:	 categoryID,
		}, nil
}

// Чтение фильтров из query-параметров URL
func parseFilters(r *http.Request) repository.Filters {
		var f repository.Filters
		if id, err := strconv.Atoi(r.URL.Query().Get("category")); err == nil {
				f.CategoryID = id
		}
		f.From = r.URL.Query().Get("from")
		f.To = r.URL.Query().Get("to")
		return f
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

// Данные для страницы списка трат
type expensesPage struct {
		Expenses 	[]models.Expense
		Categories	[]models.Category
		Filter 		repository.Filters
}

// Список трат с фильтрами
func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
		filter := parseFilters(r)

		expenses, err := repo.GetAll(filter)
		if err != nil {
				http.Error(w, "Не удалось получить список трат", http.StatusInternalServerError)
				return
		}

		categories, err := repo.GetAllCategories()
		if err != nil {
				http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
				return
		}

		render(w, "expenses.html", expensesPage{
				Expenses: 	expenses,
				Categories: categories,
				Filter:		filter,
			})
}

// Показать форму добавления и обработать отправку
func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {			
		case http.MethodGet:
				categories, err := repo.GetAllCategories()
				if err != nil {
						http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
						return
				}
				render(w, "add.html", struct{ Categories []models.Category }{Categories: categories})

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
						CategoryID:  form.CategoryID,
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
func editPage(expense models.Expense, categories []models.Category) struct {
		Expense    models.Expense
		Categories []models.Category
} {
		return struct {
				Expense    models.Expense
				Categories []models.Category
		}{Expense: expense, Categories: categories}
}

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
				categories, err := repo.GetAllCategories()
				if err != nil {
						http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
						return
				}
				render(w, "edit.html", editPage(expense, categories))

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
						CategoryID:	 form.CategoryID,
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