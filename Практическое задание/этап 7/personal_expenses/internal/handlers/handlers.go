package handlers

import (
		"errors"
		"html/template"
		"log"
		"net/http"
		"strconv"
		"strings"
		"time"
		"fmt"

		"personal_expenses/internal/models"
		"personal_expenses/internal/repository"
)

var repo *repository.ExpenseRepository


// Передача в репозиторий
func Init(r *repository.ExpenseRepository) {
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

func render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
		tmpl, err := loadTemplate(page)
		if err != nil {
				http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
				return
		}
		if data == nil {
				data = map[string]any{}
		}
		if user, ok := currentUser(r); ok {
				data["LoggedIn"] = true
				data["Email"] = user.Email
		}
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
				log.Println("ошибка рендера:", err)
		}
}

// Обязательное получение ID пользователя
func mustUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
		user, ok := currentUser(r)
		if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return 0, false
		}
		return user.ID, true
}

// Список трат с фильтрами
func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
		userID, ok := mustUserID(w, r)
		if !ok {
				return
		}
		filter := parseFilters(r)

		expenses, err := repo.GetAll(filter, userID)
		if err != nil {
				http.Error(w, "Не удалось получить список трат", http.StatusInternalServerError)
				return
		}
		categories, err := repo.GetAllCategories()
		if err != nil {
				http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
				return
		}

		render(w, r, "expenses.html", map[string]any{
				"Expenses": 	expenses,
				"Categories": 	categories,
				"Filter":		filter,
		})
}

// Показать форму добавления и обработать отправку
func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
				if _, ok := mustUserID(w, r); !ok {
						return
				}
				categories, err := repo.GetAllCategories()
				if err != nil {
						http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
						return
				}
				render(w, r, "add.html", map[string]any{"Categories": categories})

		case http.MethodPost:
				userID, ok := mustUserID(w, r)
				if !ok {
						return
				}
				categories, err := repo.GetAllCategories()
				if err != nil {
						http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
						return
				}

				form, parseErr := parseExpenseForm(r)
				if parseErr != nil {
						render(w, r, "add.html", map[string]any{
								"Categories": categories,
								"Error":      parseErr.Error(),
								"Form": map[string]string{
										"amount":      r.FormValue("amount"),
										"description": r.FormValue("description"),
										"date":        r.FormValue("date"),
										"category_id": r.FormValue("category_id"),
								},
						})
						return
				}

				if _, err := repo.Create(models.Expense{
						Amount:      form.Amount,
						Description: form.Description,
						Date:        form.Date,
						CategoryID:  form.CategoryID,
				}, userID); err != nil {
						log.Println("ошибка создания траты:", err)
						http.Error(w, "Не удалось сохранить трату", http.StatusInternalServerError)
						return
				}
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)

		default:
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
}

// Обработать отправку
func EditExpenseHandler(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
				renderError(w, r, http.StatusBadRequest)
				return
		}

		switch r.Method {
		case http.MethodGet:
				userID, ok := mustUserID(w, r)
				if !ok {
						return
				}
				e, err := repo.GetByID(id, userID)
				if errors.Is(err, repository.ErrNotFound) {
						renderError(w, r, http.StatusNotFound)
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
				render(w, r, "edit.html", map[string]any{
						"ID":        e.ID,
						"Expense":   e,
						"Categories": categories,
						"Form": map[string]string{
								"amount":      fmt.Sprintf("%.2f", e.Amount),
								"description": e.Description,
								"date":        e.Date,
								"category_id": strconv.Itoa(e.CategoryID),
						},
				})

		case http.MethodPost:
				userID, ok := mustUserID(w, r)
				if !ok {
						return
				}
				if _, err := repo.GetByID(id, userID); errors.Is(err, repository.ErrNotFound) {
						renderError(w, r, http.StatusNotFound)
						return
				} else if err != nil {
						http.Error(w, "Не удалось получить трату", http.StatusInternalServerError)
						return
				}
				categories, err := repo.GetAllCategories()
				if err != nil {
						http.Error(w, "Не удалось получить список категорий", http.StatusInternalServerError)
						return
				}

				form, parseErr := parseExpenseForm(r)
				if parseErr != nil {
						render(w, r, "edit.html", map[string]any{
								"ID":         id,
								"Categories": categories,
								"Error":      parseErr.Error(),
								"Form": map[string]string{
										"amount":      r.FormValue("amount"),
										"description": r.FormValue("description"),
										"date":        r.FormValue("date"),
										"category_id": r.FormValue("category_id"),
								},
						})
						return
				}

				if err := repo.Update(models.Expense{
						ID:          id,
						Amount:      form.Amount,
						Description: form.Description,
						Date:        form.Date,
						CategoryID:  form.CategoryID,
				}, userID); err != nil {
						log.Println("ошибка обновления траты:", err)
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
		userID, ok := mustUserID(w, r)
		if !ok {
				return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
				renderError(w, r, http.StatusBadRequest)
				return
		}
		if err := repo.Delete(id, userID); err != nil {
				http.Error(w, "Не удалось удалить трату", http.StatusInternalServerError)
				return
		}
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}