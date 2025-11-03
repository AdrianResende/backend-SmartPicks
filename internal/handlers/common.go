package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"smartpicks-backend/internal/database"
)

// sendErrorResponse envia uma resposta de erro padronizada
func sendErrorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// sendSuccessResponse envia uma resposta de sucesso padronizada
func sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// sendJSONResponse envia uma resposta JSON com status code customizado
func sendJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func userExists(field, value string) bool {
	var count int
	allowedFields := map[string]bool{
		"email":    true,
		"username": true,
		"id":       true,
	}
	if !allowedFields[field] {
		return false
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s = $1", field)
	err := database.DB.QueryRow(query, value).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
