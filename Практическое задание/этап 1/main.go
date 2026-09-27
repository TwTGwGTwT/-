package main

import (
		"fmt"
		"net/http"
)

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Добро пожаловать в «Учёт личных трат»!")
}

// Страница about
func aboutHandler(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Проект: «Учёт личных трат». Веб-разработка на Go.")
}

// Маршрут /ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
				w.Header().Set("Allow", "GET")
				http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
				return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "pong")
}

func main() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", homeHandler)
		mux.HandleFunc("/about", aboutHandler)
		mux.HandleFunc("/ping", pingHandler)

		fmt.Println("Сервер запущен на http://localhost:8080")
		http.ListenAndServe(":8080", mux)
}