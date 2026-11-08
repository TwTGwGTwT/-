package main

import (
		"log"
		"net/http"

		"personal_expenses/internal/db"
		"personal_expenses/internal/handlers"
		"personal_expenses/internal/repository"
)

func main() {
		database := db.Open("expenses.db")
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

		// Раздача статических файлов
		fs := http.FileServer(http.Dir("web/static"))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

		// Главная - редирект на список трат
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		})

		log.Fatal(http.ListenAndServe(":8080", mux))
}