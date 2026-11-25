package handlers

import (
		"log"
		"net/http"
		"strings"
		"time"
)

// Обёртка над ResponseWriter
type statusWriter struct {
		http.ResponseWriter
		status int
		wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
		if !w.wrote {
				w.status = code
				w.wrote = true
		}
		w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
		if !w.wrote {
				w.status = http.StatusOK
				w.wrote = true
		}
		return w.ResponseWriter.Write(b)
}

// Middleware логирования
func LogRequest(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
				next.ServeHTTP(sw, r)
				log.Printf("%-4s %-25s -> %d (%v)", r.Method, r.URL.Path, sw.status, time.Since(start))
		})
}

// Middleware восстановления после паники
func RecoverPanic(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
				defer func() {
						if rec := recover(); rec != nil {
								log.Printf("ПАНИКА на %s: %v", r.URL.Path, rec)
								if sw.wrote {
										return
								}
								if strings.HasPrefix(r.URL.Path, "/api/") {
										writeJSON(sw, http.StatusInternalServerError, map[string]string{
												"error": "внутренняя ошибка сервера",
										})
								} else {
										renderError(sw, r, http.StatusInternalServerError)
								}
						}
				}()
				next.ServeHTTP(sw, r)
		})
}

// Единый рендер страниц ошибок 404 и 500
func renderError(w http.ResponseWriter, r *http.Request, status int) {
		page := "500.html"
		if status == http.StatusNotFound {
				page = "404.html"
		}
		tmpl, err := loadTemplate(page)
		if err != nil {
				http.Error(w, "Ошибка сервера", status)
				return
		}
		data := map[string]any{}
		if user, ok := currentUser(r); ok {
				data["LoggedIn"] = true
				data["Email"] = user.Email
		}
		w.WriteHeader(status)
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
				log.Println("ошибка рендера страницы ошибки:", err)
		}
}

// Обработчик для всех маршрутов, которых нет в mux
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
		renderError(w, r, http.StatusNotFound)
}