package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/models"
	"smartpicks-backend/internal/services"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func GetPalpites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query(`
		SELECT 
			p.id, 
			p.user_id, 
			p.titulo, 
			p.img_url, 
			p.link, 
			p.created_at, 
			p.updated_at,
			u.avatar,
			COALESCE(likes.count, 0) AS total_likes,
			COALESCE(dislikes.count, 0) AS total_dislikes,
			COALESCE(comments.count, 0) AS total_comentarios
		FROM palpites p
		JOIN users u ON u.id = p.user_id
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
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar palpites: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var palpites []models.PalpiteResponse
	for rows.Next() {
		var p models.Palpite
		var avatar *string
		var totalLikes, totalDislikes, totalComentarios int

		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Titulo, &p.ImgURL, &p.Link, &p.CreatedAt, &p.UpdatedAt,
			&avatar,
			&totalLikes, &totalDislikes, &totalComentarios,
		); err != nil {
			sendErrorResponse(w, "Erro ao ler palpite: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if !strings.HasPrefix(p.ImgURL, "http") {
			bucketName := os.Getenv("AWS_BUCKET_NAME")
			region := os.Getenv("AWS_REGION")
			trimmedKey := strings.TrimPrefix(p.ImgURL, "/")
			p.ImgURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, trimmedKey)
		}

		p.Avatar = avatar

		response := p.ToResponse()
		response.TotalLikes = totalLikes
		response.TotalDislikes = totalDislikes
		response.TotalComentarios = totalComentarios

		comentarios, err := getComentariosByPalpiteID(p.ID)
		if err != nil {
			sendErrorResponse(w, "Erro ao buscar comentários: "+err.Error(), http.StatusInternalServerError)
			return
		}
		response.Comentarios = comentarios

		palpites = append(palpites, response)
	}

	sendSuccessResponse(w, map[string]interface{}{
		"palpites": palpites,
	})
}

// GetPalpiteByID @Summary Buscar um palpite específico por ID
// @Description Retorna um palpite específico com todas as estatísticas e comentários
// @Tags Palpites
// @Produce json
// @Param id path int true "ID do palpite"
// @Success 200 {object} models.PalpiteResponse "Palpite encontrado"
// @Failure 404 {object} map[string]string "Palpite não encontrado"
// @Failure 500 {object} map[string]string "Erro interno do servidor"
// @Router /palpites/{id} [get]
func GetPalpiteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		sendErrorResponse(w, "ID do palpite é obrigatório", http.StatusBadRequest)
		return
	}

	palpiteID, err := strconv.Atoi(id)
	if err != nil {
		sendErrorResponse(w, "ID do palpite inválido", http.StatusBadRequest)
		return
	}

	row := database.DB.QueryRow(`
		SELECT 
			p.id,
			p.user_id,
			p.titulo,
			p.img_url,
			p.link,
			p.created_at,
			p.updated_at,
			u.avatar,
			COALESCE(likes.count, 0) AS total_likes,
			COALESCE(dislikes.count, 0) AS total_dislikes,
			COALESCE(comentarios.count, 0) AS total_comentarios
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
		) comentarios ON p.id = comentarios.palpite_id
		WHERE p.id = $1
	`, palpiteID)

	var palpite models.Palpite
	var avatar *string
	var totalLikes, totalDislikes, totalComentarios int

	err = row.Scan(
		&palpite.ID,
		&palpite.UserID,
		&palpite.Titulo,
		&palpite.ImgURL,
		&palpite.Link,
		&palpite.CreatedAt,
		&palpite.UpdatedAt,
		&avatar,
		&totalLikes,
		&totalDislikes,
		&totalComentarios,
	)

	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Palpite não encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		sendErrorResponse(w, "Erro ao buscar palpite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Buscar comentários do palpite
	comentarios, err := getComentariosByPalpiteID(palpite.ID)
	if err != nil {
		sendErrorResponse(w, "Erro ao buscar comentários: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Construir URL completa da imagem se for caminho S3
	if strings.HasPrefix(palpite.ImgURL, "palpites/") {
		bucketName := os.Getenv("AWS_BUCKET_NAME")
		region := os.Getenv("AWS_REGION")
		palpite.ImgURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, palpite.ImgURL)
	}

	// Construir resposta
	response := models.PalpiteResponse{
		ID:               palpite.ID,
		UserID:           palpite.UserID,
		Titulo:           palpite.Titulo,
		ImgURL:           palpite.ImgURL,
		Link:             palpite.Link,
		CreatedAt:        palpite.CreatedAt,
		UpdatedAt:        palpite.UpdatedAt,
		Avatar:           avatar,
		TotalLikes:       totalLikes,
		TotalDislikes:    totalDislikes,
		TotalComentarios: totalComentarios,
		Comentarios:      comentarios,
	}

	sendSuccessResponse(w, map[string]interface{}{
		"palpite": response,
	})
}

// Helper function para buscar comentários de um palpite
func getComentariosByPalpiteID(palpiteID int) ([]models.ComentarioStats, error) {
	rows, err := database.DB.Query(`
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
			p.link AS palpite_link
		FROM comentarios c
		JOIN users u ON c.user_id = u.id
		JOIN palpites p ON c.palpite_id = p.id
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
		WHERE c.palpite_id = $1
		ORDER BY c.created_at ASC
	`, palpiteID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comentarios []models.ComentarioStats
	for rows.Next() {
		var c models.ComentarioStats
		if err := rows.Scan(
			&c.ID,
			&c.PalpiteID,
			&c.UserID,
			&c.Texto,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.TotalLikes,
			&c.TotalDislikes,
			&c.AutorNome,
			&c.AutorAvatar,
			&c.PalpiteLink,
		); err != nil {
			return nil, err
		}
		comentarios = append(comentarios, c)
	}

	return comentarios, nil
}

// PostPalpite @Summary Criar um novo palpite
// @Description Cria um novo palpite com upload para o S3
// @Tags Palpites
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData int true "ID do usuário"
// @Param titulo formData string false "Título do palpite"
// @Param link formData string false "Link do palpite"
// @Param image formData file true "Imagem do palpite"
// @Success 201 {object} map[string]interface{} "Palpite criado com sucesso"
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 500 {object} map[string]string "Erro interno do servidor"
// @Router /palpites [post]
func PostPalpite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		sendErrorResponse(w, fmt.Sprintf("Content-Type inválido: %s, esperado multipart/form-data", contentType), http.StatusBadRequest)
		return
	}

	// Removemos o limitador de 20MB
	if err := r.ParseMultipartForm(0); err != nil {
		sendErrorResponse(w, "Erro ao processar multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	titulo := r.FormValue("titulo")
	link := r.FormValue("link")

	if userID == "" {
		sendErrorResponse(w, "user_id é obrigatório", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		sendErrorResponse(w, "Erro ao receber arquivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	s3Service, err := services.NewS3Service()
	if err != nil {
		sendErrorResponse(w, "Erro ao configurar S3: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ext := ".jpg"
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("palpite_%d%s", timestamp, ext)

	imageURL, err := s3Service.UploadFile(file, fileName, "image/jpeg")
	if err != nil {
		sendErrorResponse(w, "Erro ao enviar para S3: "+err.Error(), http.StatusInternalServerError)
		return
	}

	palpite := models.Palpite{
		UserID:    stringToInt(userID),
		Titulo:    &titulo,
		ImgURL:    imageURL,
		Link:      &link,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = database.DB.QueryRow(`
		INSERT INTO palpites (user_id, titulo, img_url, link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		palpite.UserID, palpite.Titulo, palpite.ImgURL, palpite.Link,
		palpite.CreatedAt, palpite.UpdatedAt,
	).Scan(&palpite.ID)

	if err != nil {
		sendErrorResponse(w, "Erro ao salvar palpite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, map[string]interface{}{
		"palpite": palpite.ToResponse(),
		"message": "Palpite criado com sucesso",
	})
}

func stringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
