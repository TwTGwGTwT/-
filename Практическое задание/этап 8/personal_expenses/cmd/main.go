package main

import (
		"log"
		"net/http"

		"personal_expenses/internal/config"
		"personal_expenses/internal/db"
		"personal_expenses/internal/handlers"
		"personal_expenses/internal/repository"
)

func main() {
		cfg := config.Load()

		database := db.Open("cfg.DBPath")
		defer database.Close()

		repo := repository.NewExpenseRepository(database)
		userRepo := repository.NewUserRepository(database)

		handlers.Init(repo)
		handlers.InitAuth(userRepo)

		mux := http.NewServeMux()

		// Открытые маршруты
		mux.HandleFunc("GET /register", handlers.RegisterHandler)
		mux.HandleFunc("POST /register", handlers.RegisterHandler)
		mux.HandleFunc("GET /login", handlers.LoginHandler)
		mux.HandleFunc("POST /login", handlers.LoginHandler)
		mux.HandleFunc("POST /logout", handlers.LogoutHandler)

		// Защищённые маршруты
		mux.HandleFunc("GET /expenses", handlers.RequireAuth(handlers.ExpensesHandler))
		mux.HandleFunc("GET /expenses/add", handlers.RequireAuth(handlers.AddExpenseHandler))
		mux.HandleFunc("POST /expenses/add", handlers.RequireAuth(handlers.AddExpenseHandler))
		mux.HandleFunc("GET /expenses/{id}/edit", handlers.RequireAuth(handlers.EditExpenseHandler))
		mux.HandleFunc("POST /expenses/{id}/edit", handlers.RequireAuth(handlers.EditExpenseHandler))
		mux.HandleFunc("POST /expenses/{id}/delete", handlers.RequireAuth(handlers.DeleteExpenseHandler))
		mux.HandleFunc("GET /stats", handlers.RequireAuth(handlers.StatsPageHandler))
		mux.HandleFunc("GET /api/stats/by-category", handlers.RequireAuth(handlers.StatsByCategoryHandler))
		mux.HandleFunc("GET /api/stats/by-month", handlers.RequireAuth(handlers.StatsByMonthHandler))

		// Экспорт трат текущего пользователя в CSV
		mux.HandleFunc("GET /expenses/export", handlers.RequireAuth(handlers.ExportCSVHandler))

		// Раздача статических файлов
		fs := http.FileServer(http.Dir("web/static"))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

		// Главная - редирект на список трат
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		})

		mux.HandleFunc("/", handlers.NotFoundHandler)

		handler := handlers.LogRequest(handlers.RecoverPanic(mux))

		addr := ":" + cfg.Port
		log.Printf("Сервер запущен на http://localhost:%s", cfg.Port)
		log.Fatal(http.ListenAndServe(addr, handler))
}