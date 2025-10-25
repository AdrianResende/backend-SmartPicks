package handlers

import (
	"net/http"
	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/models"
)

// GetAllMatches @Summary Obter todas as partidas
// @Description Retorna uma lista de todas as partidas disponíveis
// @Tags Matches
func GetAllMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	rows, err := database.DB.Query(`
		SELECT id, team_a, team_b, match_date, location
		FROM matches
		ORDER BY match_date ASC
	`)
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar partidas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var matches []models.Match
	for rows.Next() {
		var m models.Match
		err := rows.Scan(&m.ID, &m.TeamA, &m.TeamB, &m.MatchDate, &m.Location)
		if err != nil {
			sendErrorResponse(w, "Erro ao ler partida: "+err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}
	sendSuccessResponse(w, map[string]interface{}{
		"matches": matches,
	})
}
