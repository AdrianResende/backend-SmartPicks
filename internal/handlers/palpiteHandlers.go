package handlers

import (
	"fmt"
	"net/http"
	"os"
	"smartpicks-backend/internal/database"
	"smartpicks-backend/internal/models"
	"smartpicks-backend/internal/services"
	"strconv"
	"strings"
	"time"
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
			u.avatar
		FROM palpites p
		JOIN users u ON u.id = p.user_id
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

		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Titulo, &p.ImgURL, &p.Link, &p.CreatedAt, &p.UpdatedAt,
			&avatar,
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

		palpites = append(palpites, p.ToResponse())
	}

	sendSuccessResponse(w, map[string]interface{}{
		"palpites": palpites,
	})
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
