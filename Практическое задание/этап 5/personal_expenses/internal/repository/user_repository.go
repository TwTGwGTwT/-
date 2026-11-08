package repository

import (
		"database/sql"
		"errors"
		"strings"
		"time"

		"personal_expenses/internal/models"
)

type UserRepository struct {
		db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
		return &UserRepository{db: db}
}

// Создание пользователя
func (r *UserRepository) Create(email, passwordHash string) (int, error) {
		res, err := r.db.Exec(
				"INSERT INTO users (email, password_hash) VALUES (?, ?)",
				email, passwordHash,
		)
		if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
						return 0, ErrEmailTaken
				}
				return 0, err
		}
		id, err := res.LastInsertId()
		if err != nil {
				return 0, err
		}
		return int(id), nil
}

// Поиск пользователя по email и хэш пароля
func (r *UserRepository) GetByEmail(email string) (models.User, string, error) {
		var u models.User
		var hash string
		err := r.db.QueryRow(
				"SELECT id, email, password_hash FROM users WHERE email = ?", email,
		).Scan(&u.ID, &u.Email, &hash)
		if errors.Is(err, sql.ErrNoRows) {
				return models.User{}, "", ErrNotFound
		}
		return u, hash, err
}

// Создание сессии на 7 дней
func (r *UserRepository) CreateSession(token string, userID int) error {
		expires := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
		_, err := r.db.Exec(
				"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
				token, userID, expires,
		)
		return err
}

// Поиск пользователя по токену сессии
func (r *UserRepository) GetUserBySession(token string) (models.User, error) {
		var u models.User
		err := r.db.QueryRow(
				"SELECT u.id, u.email FROM users u "+
						"JOIN sessions s ON s.user_id = u.id "+
						"WHERE s.token = ? AND s.expires_at > ?",
				token, time.Now().UTC().Format(time.RFC3339),
		).Scan(&u.ID, &u.Email)
		if errors.Is(err, sql.ErrNoRows) {
				return models.User{}, ErrNotFound
		}
		return u, err
}

// Удаление сессии
func (r *UserRepository) DeleteSession(token string) error {
		_, err := r.db.Exec("DELETE FROM sessions WHERE token = ?", token)
		return err
}