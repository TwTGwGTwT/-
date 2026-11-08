package handlers

import (
		"context"
		"crypto/rand"
		"encoding/hex"
		"errors"
		"log"
		"net/http"
		"strings"
		"time"

		"personal_expenses/internal/models"
		"personal_expenses/internal/repository"

		"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "session_token"

var authRepo *repository.UserRepository

func InitAuth(r *repository.UserRepository) {
		authRepo = r
}

// Ключ для передачи пользователя через контекст запроса
type contextKey string

const userContextKey contextKey = "user"

func currentUser(r *http.Request) (models.User, bool) {
		u, ok := r.Context().Value(userContextKey).(models.User)
		return u, ok
}

// Пускает дальше только с валидной сессией
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
				c, err := r.Cookie(sessionCookieName)
				if err != nil || c.Value == "" {
						http.Redirect(w, r, "/login", http.StatusSeeOther)
						return
				}
				user, err := authRepo.GetUserBySession(c.Value)
				if err != nil {
						http.Redirect(w, r, "/login", http.StatusSeeOther)
						return
				}
				ctx := context.WithValue(r.Context(), userContextKey, user)
				next(w, r.WithContext(ctx))
		}
}

// Случайный токен сессии
func newToken() (string, error) {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
				return "", err
		}
		return hex.EncodeToString(b), nil
}

func setSessionCookie(w http.ResponseWriter, token string) {
		http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
		})
}

func clearSessionCookie(w http.ResponseWriter) {
		http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
		})
}

// Регистрация
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
				render(w, r, "register.html", nil)
		case http.MethodPost:
				if err := r.ParseForm(); err != nil {
						http.Error(w, "не удалось прочитать форму", http.StatusBadRequest)
						return
				}
				email := strings.TrimSpace(r.FormValue("email"))
				password := r.FormValue("password")

				if !strings.Contains(email, "@") || len(email) > 200 {
						render(w, r, "register.html", map[string]any{
								"Error": "введите корректный email",
						})
						return
				}
				if len(password) < 6 {
						render(w, r, "register.html", map[string]any{
								"Error": "пароль должен быть не короче 6 символов",
						})
						return
				}

				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
						http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
						return
				}

				userID, err := authRepo.Create(email, string(hash))
				if errors.Is(err, repository.ErrEmailTaken) {
						render(w, r, "register.html", map[string]any{
								"Error": "этот email уже зарегистрирован",
						})
						return
				}
				if err != nil {
						log.Println("ошибка регистрации:", err)
						http.Error(w, "Не удалось зарегистрироваться", http.StatusInternalServerError)
						return
				}

				//Создание сессии и пользователь входит автоматически
				token, err := newToken()
				if err == nil {
						if err := authRepo.CreateSession(token, userID); err == nil {
								setSessionCookie(w, token)
						}
				}
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		default:
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
}

// Вход
func LoginHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
				render(w, r, "login.html", nil)
		case http.MethodPost:
				if err := r.ParseForm(); err != nil {
						http.Error(w, "не удалось прочитать форму", http.StatusBadRequest)
						return
				}
				email := strings.TrimSpace(r.FormValue("email"))
				password := r.FormValue("password")

				user, hash, err := authRepo.GetByEmail(email)
				// Та же ошибка, чтобы не подсказывать, какие email существуют
				if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
						render(w, r, "login.html", map[string]any{
								"Error": "неверный email или пароль",
						})
						return
				}

				token, err := newToken()
				if err != nil {
						http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
						return
				}
				if err := authRepo.CreateSession(token, user.ID); err != nil {
						http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
						return
				}
				setSessionCookie(w, token)
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		default:
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
}

// Выход
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
				return
		}
		if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
				if err := authRepo.DeleteSession(c.Value); err != nil {
						log.Println("ошибка удаления сессии:", err)
				}
		}
		clearSessionCookie(w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
}