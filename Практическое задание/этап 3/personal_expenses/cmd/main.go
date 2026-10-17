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
		handlers.Init(repo)

		mux := http.NewServeMux()

		// Маршруты
		mux.HandleFunc("GET /expenses", handlers.ExpensesHandler)
		mux.HandleFunc("GET /expenses/add", handlers.AddExpenseHandler)
		mux.HandleFunc("POST /expenses/add", handlers.AddExpenseHandler)
		mux.HandleFunc("GET /expenses/{id}/edit", handlers.EditExpenseHandler)
		mux.HandleFunc("POST /expenses/{id}/edit", handlers.EditExpenseHandler)
		mux.HandleFunc("POST /expenses/{id}/delete", handlers.DeleteExpenseHandler)

		// Раздача статических файлов
		fs := http.FileServer(http.Dir("web/static"))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

		// Главная - редирект на список трат
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		})

		log.Fatal(http.ListenAndServe(":8080", mux))
}