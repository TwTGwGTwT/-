package main

import (
		"personal_expenses/internal/handlers"
		"net/http"
)

func main() {
		mux := http.NewServeMux()

		// Маршруты
		mux.HandleFunc("/expenses", handlers.ExpensesHandler)
		mux.HandleFunc("/expenses/add", handlers.AddExpenseHandler)

		// Раздача статических файлов
		fs := http.FileServer(http.Dir("web/static"))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))

		// Главная - редирект на список трат
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		})

		http.ListenAndServe(":8080", mux)
}