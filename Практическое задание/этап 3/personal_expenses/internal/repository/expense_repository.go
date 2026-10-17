package repository

import (
		"database/sql"
		"errors"

		"personal_expenses/internal/models"
)

// Возвращается, когда запись с указанным ID не существует
var ErrNotFound = errors.New("запись не найдена")

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
				"INSERT INTO expenses (amount, description, date) VALUES (?, ?, ?)",
				e.Amount, e.Description, e.Date,
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

// Возвращение всех трат
func (r *ExpenseRepository) GetAll() ([]models.Expense, error) {
		rows, err := r.db.Query("SELECT id, amount, description, date FROM expenses ORDER BY id DESC")
		if err != nil {
				return nil, err
		}
		defer rows.Close()

		var expenses []models.Expense
		for rows.Next() {
				var e models.Expense
				if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date); err != nil {
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
				"SELECT id, amount, description, date FROM expenses WHERE id = ?", id,
		).Scan(&e.ID, &e.Amount, &e.Description, &e.Date)

		if errors.Is(err, sql.ErrNoRows) {
				return models.Expense{}, ErrNotFound
		}
		return e, err
}

// Изменение существующей траты
func (r *ExpenseRepository) Update(e models.Expense) error {
		_, err := r.db.Exec(
				"UPDATE expenses SET amount = ?, description = ?, date = ? WHERE id = ?",
				e.Amount, e.Description, e.Date, e.ID,
		)
		return err
}

// Удаление траты по ID
func (r *ExpenseRepository) Delete(id int) error {
		_, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
		return err
}