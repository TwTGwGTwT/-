package main

import (
		"fmt"
		"net/http"
		"log"
)

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
			http.Error(w, "Страница не найдена", http.StatusNotFound)
			return
	}
	fmt.Fprintln(w, "Добро пожаловать в «Учёт личных трат»!")

}

// Страница about
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/about" {
			http.Error(w, "Страница не найдена", http.StatusNotFound)
			return
	}
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

		// Обработка всех остальных путей
		mux.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Страница не найдена", http.StatusNotFound)
		})

		// Обработка ошибок запуска сервера
		log.Println("Сервер запущен на http://localhost:8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
				log.Fatalf("Ошибка запуска сервера: %v", err)
		}
}