package routes

import (
	"net/http"

	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/handlers"

	"github.com/gorilla/mux"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := map[string]bool{
			"http://localhost:9000":                    true,
			"https://smartpicks-88709.web.app":         true,
			"https://smartpicks-88709.firebaseapp.com": true,
		}
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, Accept, X-User-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RegisterRoutes(r *mux.Router) {
	database.Connect()

	r.Use(enableCORS)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", handlers.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/register", handlers.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/users", handlers.GetAllUsers).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/permissions", handlers.CheckUserPermissions).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/profile", handlers.GetUsersByProfile).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/avatar", handlers.UpdateAvatar).Methods("POST", "PUT", "OPTIONS")
	api.HandleFunc("/users/avatar", handlers.DeleteAvatar).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/users/{id}", handlers.GetUserByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/matches", handlers.GetAllMatches).Methods("GET", "OPTIONS")

	api.HandleFunc("/palpites/stats", handlers.GetAllPalpitesWithStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/palpites/{id}/stats", handlers.GetPalpiteStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/palpites/{id}/react", handlers.TogglePalpiteReaction).Methods("POST", "OPTIONS")
	api.HandleFunc("/palpites/{id}/comentarios", handlers.GetComentariosByPalpite).Methods("GET", "OPTIONS")
	api.HandleFunc("/palpites/{id}", handlers.GetPalpiteByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/palpites", handlers.GetPalpites).Methods("GET", "OPTIONS")
	api.HandleFunc("/palpites", handlers.PostPalpite).Methods("POST", "OPTIONS")

	api.HandleFunc("/comentarios/{id}/react", handlers.ToggleComentarioReaction).Methods("POST", "OPTIONS")
	api.HandleFunc("/comentarios/{id}", handlers.UpdateComentario).Methods("PUT", "OPTIONS")
	api.HandleFunc("/comentarios/{id}", handlers.DeleteComentario).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/comentarios", handlers.CreateComentario).Methods("POST", "OPTIONS")

	api.HandleFunc("/upload", handlers.UploadImageHandler).Methods("POST", "OPTIONS")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "API rodando", "version": "1.0.0"}`))
	}).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "healthy"}`))
	}).Methods("GET")
}
