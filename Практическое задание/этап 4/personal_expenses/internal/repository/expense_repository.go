package repository

import (
		"database/sql"
		"errors"
		"strings"

		"personal_expenses/internal/models"
)

// Возвращается, когда запись с указанным ID не существует
var ErrNotFound = errors.New("запись не найдена")

// Фильтры списка трат
type Filters struct {
		CategoryID int
		From	   string
		To		   string
}

// Инкапсуляция всего SQL-кода для работы с тратами
type ExpenseRepository struct {
		db *sql.DB
}

// Создание репозитория поверх открытого соединения с базой
func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
		return &ExpenseRepository{db: db}
}

// Добавление траты и её возвращение с присвоенным ID
func (r *ExpenseRepository) Create(e models.Expense) (models.Expense, error) {
		// Защита от SQL-инъекций
		res, err := r.db.Exec(
				"INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)",
				e.Amount, e.Description, e.Date, e.CategoryID,
		)
		if err != nil {
				return models.Expense{}, err
		}

		id, err := res.LastInsertId()
		if err != nil {
				return models.Expense{}, err
		}
		e.ID = int(id)
		return e, nil
}

// Возвращение трат с фильтрацией
func (r *ExpenseRepository) GetAll(f Filters) ([]models.Expense, error) {
		query := "SELECT e.id, e.amount, e.description, e.date, e.category_id, COALESCE(c.name, '') " +
				"FROM expenses e LEFT JOIN categories c ON c.id = e.category_id"

		var conds []string
		var args []any

		if f.CategoryID > 0 {
				conds = append(conds, "e.category_id = ?")
				args = append(args, f.CategoryID)
		}
		if f.From != "" {
				conds = append(conds, "e.date >= ?")
				args = append(args, f.From)
		}
		if f.To != "" {
				conds = append(conds, "e.date <= ?")
				args = append(args, f.To)
		}
		if len(conds) > 0 {
				query += " WHERE " + strings.Join(conds, " AND ")
		}
		query += " ORDER BY e.id DESC"

		rows, err := r.db.Query(query, args...)
		if err != nil {
				return nil, err
		}
		defer rows.Close()

		var expenses []models.Expense
		for rows.Next() {
				var e models.Expense
				if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName); err != nil {
						return nil, err
				}
				expenses = append(expenses, e)
		}
		return expenses, rows.Err()
}

// Возвращение одной траты по ID
func (r *ExpenseRepository) GetByID(id int) (models.Expense, error) {
		var e models.Expense
		err := r.db.QueryRow(
				"SELECT e.id, e.amount, e.description, e.date, e.category_id, COALESCE(c.name, '') "+
						"FROM expenses e LEFT JOIN categories c ON c.id = e.category_id "+
						"WHERE e.id = ?", id,
		).Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName)
		if errors.Is(err, sql.ErrNoRows) {
				return models.Expense{}, ErrNotFound
		}
		return e, err
}

// Все категории из базы
func (r *ExpenseRepository) GetAllCategories() ([]models.Category, error) {
		rows, err := r.db.Query("SELECT id, name FROM categories ORDER BY name")
		if err != nil {
				return nil, err
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
				var c models.Category
				if err := rows.Scan(&c.ID, &c.Name); err != nil {
						return nil, err
				}
				categories = append(categories, c)
		}
		return categories, rows.Err()
}

// Изменение существующей траты
func (r *ExpenseRepository) Update(e models.Expense) error {
		_, err := r.db.Exec(
				"UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?",
				e.Amount, e.Description, e.Date, e.CategoryID, e.ID,
		)
		return err
}

// Удаление траты по ID
func (r *ExpenseRepository) Delete(id int) error {
		_, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
		return err
}