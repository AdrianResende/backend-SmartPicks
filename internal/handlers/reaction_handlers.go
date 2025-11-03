package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/models"

	"github.com/gorilla/mux"
)

// TogglePalpiteReaction adiciona, remove ou altera a reação de um usuário em um palpite
func TogglePalpiteReaction(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		sendErrorResponse(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	palpiteID := vars["id"]

	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Tipo != "like" && req.Tipo != "dislike" {
		sendErrorResponse(w, "Tipo de reação inválido. Use 'like' ou 'dislike'", http.StatusBadRequest)
		return
	}

	// Verificar se o palpite existe
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM palpites WHERE id = $1)", palpiteID).Scan(&exists)
	if err != nil || !exists {
		sendErrorResponse(w, "Palpite não encontrado", http.StatusNotFound)
		return
	}

	// Chamar a função do PostgreSQL para toggle de reação
	var response models.ReactionResponse
	err = database.DB.QueryRow(`
		SELECT * FROM toggle_palpite_reaction($1, $2, $3)
	`, palpiteID, userID, req.Tipo).Scan(
		&response.Action,
		&response.TotalLikes,
		&response.TotalDislikes,
	)

	if err != nil {
		sendErrorResponse(w, "Erro ao processar reação", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, response, http.StatusOK)
}

// ToggleComentarioReaction adiciona, remove ou altera a reação de um usuário em um comentário
func ToggleComentarioReaction(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		sendErrorResponse(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	comentarioID := vars["id"]

	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Tipo != "like" && req.Tipo != "dislike" {
		sendErrorResponse(w, "Tipo de reação inválido. Use 'like' ou 'dislike'", http.StatusBadRequest)
		return
	}

	// Verificar se o comentário existe
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM comentarios WHERE id = $1)", comentarioID).Scan(&exists)
	if err != nil || !exists {
		sendErrorResponse(w, "Comentário não encontrado", http.StatusNotFound)
		return
	}

	// Chamar a função do PostgreSQL para toggle de reação
	var response models.ReactionResponse
	err = database.DB.QueryRow(`
		SELECT * FROM toggle_comentario_reaction($1, $2, $3)
	`, comentarioID, userID, req.Tipo).Scan(
		&response.Action,
		&response.TotalLikes,
		&response.TotalDislikes,
	)

	if err != nil {
		sendErrorResponse(w, "Erro ao processar reação", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, response, http.StatusOK)
}

// GetPalpiteStats retorna as estatísticas de um palpite específico
func GetPalpiteStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	palpiteID := vars["id"]
	userID := GetUserIDFromRequest(r)

	query := `
		SELECT 
			p.id,
			p.user_id,
			p.titulo,
			p.img_url,
			p.link,
			p.created_at,
			p.updated_at,
			COALESCE(likes.count, 0) AS total_likes,
			COALESCE(dislikes.count, 0) AS total_dislikes,
			COALESCE(comments.count, 0) AS total_comentarios,
			u.nome AS autor_nome,
			u.avatar AS autor_avatar,
			ur.tipo AS user_reaction
		FROM palpites p
		LEFT JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM palpites_reactions 
			WHERE tipo = 'like' 
			GROUP BY palpite_id
		) likes ON p.id = likes.palpite_id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM palpites_reactions 
			WHERE tipo = 'dislike' 
			GROUP BY palpite_id
		) dislikes ON p.id = dislikes.palpite_id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM comentarios 
			GROUP BY palpite_id
		) comments ON p.id = comments.palpite_id
		LEFT JOIN palpites_reactions ur ON p.id = ur.palpite_id AND ur.user_id = $2
		WHERE p.id = $1
	`

	var palpite models.PalpiteStats
	err := database.DB.QueryRow(query, palpiteID, userID).Scan(
		&palpite.ID,
		&palpite.UserID,
		&palpite.Titulo,
		&palpite.ImgURL,
		&palpite.Link,
		&palpite.CreatedAt,
		&palpite.UpdatedAt,
		&palpite.TotalLikes,
		&palpite.TotalDislikes,
		&palpite.TotalComentarios,
		&palpite.AutorNome,
		&palpite.AutorAvatar,
		&palpite.UserReaction,
	)

	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Palpite não encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar palpite", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, palpite, http.StatusOK)
}

// GetAllPalpitesWithStats retorna todos os palpites com estatísticas
func GetAllPalpitesWithStats(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)

	query := `
		SELECT 
			p.id,
			p.user_id,
			p.titulo,
			p.img_url,
			p.link,
			p.created_at,
			p.updated_at,
			COALESCE(likes.count, 0) AS total_likes,
			COALESCE(dislikes.count, 0) AS total_dislikes,
			COALESCE(comments.count, 0) AS total_comentarios,
			u.nome AS autor_nome,
			u.avatar AS autor_avatar,
			ur.tipo AS user_reaction
		FROM palpites p
		LEFT JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM palpites_reactions 
			WHERE tipo = 'like' 
			GROUP BY palpite_id
		) likes ON p.id = likes.palpite_id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM palpites_reactions 
			WHERE tipo = 'dislike' 
			GROUP BY palpite_id
		) dislikes ON p.id = dislikes.palpite_id
		LEFT JOIN (
			SELECT palpite_id, COUNT(*) as count 
			FROM comentarios 
			GROUP BY palpite_id
		) comments ON p.id = comments.palpite_id
		LEFT JOIN palpites_reactions ur ON p.id = ur.palpite_id AND ur.user_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar palpites", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var palpites []models.PalpiteStats
	for rows.Next() {
		var palpite models.PalpiteStats
		err := rows.Scan(
			&palpite.ID,
			&palpite.UserID,
			&palpite.Titulo,
			&palpite.ImgURL,
			&palpite.Link,
			&palpite.CreatedAt,
			&palpite.UpdatedAt,
			&palpite.TotalLikes,
			&palpite.TotalDislikes,
			&palpite.TotalComentarios,
			&palpite.AutorNome,
			&palpite.AutorAvatar,
			&palpite.UserReaction,
		)
		if err != nil {
			sendErrorResponse(w, "Erro ao processar palpites", http.StatusInternalServerError)
			return
		}
		palpites = append(palpites, palpite)
	}

	sendJSONResponse(w, palpites, http.StatusOK)
}
