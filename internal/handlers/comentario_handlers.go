package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

// GetComentariosByPalpite retorna todos os comentários de um palpite com estatísticas
func GetComentariosByPalpite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	palpiteID := vars["id"]
	userID := GetUserIDFromRequest(r)

	query := `
		SELECT 
			c.id,
			c.palpite_id,
			c.user_id,
			c.texto,
			c.created_at,
			c.updated_at,
			COALESCE(likes.count, 0) AS total_likes,
			COALESCE(dislikes.count, 0) AS total_dislikes,
			u.nome AS autor_nome,
			u.avatar AS autor_avatar,
			ur.tipo AS user_reaction
		FROM comentarios c
		LEFT JOIN users u ON c.user_id = u.id
		LEFT JOIN (
			SELECT comentario_id, COUNT(*) as count 
			FROM comentarios_reactions 
			WHERE tipo = 'like' 
			GROUP BY comentario_id
		) likes ON c.id = likes.comentario_id
		LEFT JOIN (
			SELECT comentario_id, COUNT(*) as count 
			FROM comentarios_reactions 
			WHERE tipo = 'dislike' 
			GROUP BY comentario_id
		) dislikes ON c.id = dislikes.comentario_id
		LEFT JOIN comentarios_reactions ur ON c.id = ur.comentario_id AND ur.user_id = $2
		WHERE c.palpite_id = $1
		ORDER BY c.created_at ASC
	`

	rows, err := database.DB.Query(query, palpiteID, userID)
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar comentários", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comentarios []models.ComentarioStats
	for rows.Next() {
		var comentario models.ComentarioStats
		err := rows.Scan(
			&comentario.ID,
			&comentario.PalpiteID,
			&comentario.UserID,
			&comentario.Texto,
			&comentario.CreatedAt,
			&comentario.UpdatedAt,
			&comentario.TotalLikes,
			&comentario.TotalDislikes,
			&comentario.AutorNome,
			&comentario.AutorAvatar,
			&comentario.UserReaction,
		)
		if err != nil {
			sendErrorResponse(w, "Erro ao processar comentários", http.StatusInternalServerError)
			return
		}
		comentarios = append(comentarios, comentario)
	}

	sendJSONResponse(w, comentarios, http.StatusOK)
}

// CreateComentario cria um novo comentário em um palpite
func CreateComentario(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		sendErrorResponse(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	var req models.ComentarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Texto == "" || len(req.Texto) > 1000 {
		sendErrorResponse(w, "Texto do comentário deve ter entre 1 e 1000 caracteres", http.StatusBadRequest)
		return
	}

	// Verificar se o palpite existe
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM palpites WHERE id = $1)", req.PalpiteID).Scan(&exists)
	if err != nil || !exists {
		sendErrorResponse(w, "Palpite não encontrado", http.StatusNotFound)
		return
	}

	// Inserir comentário
	var comentario models.Comentario
	err = database.DB.QueryRow(`
		INSERT INTO comentarios (palpite_id, user_id, texto)
		VALUES ($1, $2, $3)
		RETURNING id, palpite_id, user_id, texto, created_at, updated_at
	`, req.PalpiteID, userID, req.Texto).Scan(
		&comentario.ID,
		&comentario.PalpiteID,
		&comentario.UserID,
		&comentario.Texto,
		&comentario.CreatedAt,
		&comentario.UpdatedAt,
	)

	if err != nil {
		sendErrorResponse(w, "Erro ao criar comentário", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, comentario, http.StatusCreated)
}

// UpdateComentario atualiza um comentário existente
func UpdateComentario(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		sendErrorResponse(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	comentarioID := vars["id"]

	var req struct {
		Texto string `json:"texto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Texto == "" || len(req.Texto) > 1000 {
		sendErrorResponse(w, "Texto do comentário deve ter entre 1 e 1000 caracteres", http.StatusBadRequest)
		return
	}

	// Verificar se o comentário pertence ao usuário
	var comentarioUserID int
	err := database.DB.QueryRow("SELECT user_id FROM comentarios WHERE id = $1", comentarioID).Scan(&comentarioUserID)
	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Comentário não encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar comentário", http.StatusInternalServerError)
		return
	}

	if comentarioUserID != userID {
		sendErrorResponse(w, "Você não tem permissão para editar este comentário", http.StatusForbidden)
		return
	}

	// Atualizar comentário
	var comentario models.Comentario
	err = database.DB.QueryRow(`
		UPDATE comentarios 
		SET texto = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, palpite_id, user_id, texto, created_at, updated_at
	`, req.Texto, comentarioID).Scan(
		&comentario.ID,
		&comentario.PalpiteID,
		&comentario.UserID,
		&comentario.Texto,
		&comentario.CreatedAt,
		&comentario.UpdatedAt,
	)

	if err != nil {
		sendErrorResponse(w, "Erro ao atualizar comentário", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, comentario, http.StatusOK)
}

// DeleteComentario deleta um comentário
func DeleteComentario(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		sendErrorResponse(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	comentarioID := vars["id"]

	// Verificar se o comentário pertence ao usuário ou se é admin
	var comentarioUserID int
	var userPerfil string
	err := database.DB.QueryRow(`
		SELECT c.user_id, u.perfil
		FROM comentarios c
		JOIN users u ON u.id = $1
		WHERE c.id = $2
	`, userID, comentarioID).Scan(&comentarioUserID, &userPerfil)

	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Comentário não encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar comentário", http.StatusInternalServerError)
		return
	}

	if comentarioUserID != userID && userPerfil != "admin" {
		sendErrorResponse(w, "Você não tem permissão para deletar este comentário", http.StatusForbidden)
		return
	}

	// Deletar comentário
	_, err = database.DB.Exec("DELETE FROM comentarios WHERE id = $1", comentarioID)
	if err != nil {
		sendErrorResponse(w, "Erro ao deletar comentário", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"message": "Comentário deletado com sucesso"}, http.StatusOK)
}

// GetUserIDFromRequest retorna o ID do usuário da requisição (pelo token JWT)
func GetUserIDFromRequest(r *http.Request) int {
	// Esta função deve extrair o userID do token JWT
	// Por enquanto, retorna 0 (implementar autenticação JWT posteriormente)
	userIDStr := r.Header.Get("X-User-ID") // Temporário
	if userIDStr == "" {
		return 0
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0
	}
	return userID
}
