package models

import "time"

// Comentario representa um comentário em um palpite
type Comentario struct {
	ID        int       `json:"id"`
	PalpiteID int       `json:"palpite_id"`
	UserID    int       `json:"user_id"`
	Texto     string    `json:"texto"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ComentarioStats representa um comentário com estatísticas de likes/dislikes
type ComentarioStats struct {
	ID            int       `json:"id"`
	PalpiteID     int       `json:"palpite_id"`
	UserID        int       `json:"user_id"`
	Texto         string    `json:"texto"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TotalLikes    int       `json:"total_likes"`
	TotalDislikes int       `json:"total_dislikes"`
	AutorNome     string    `json:"autor_nome"`
	AutorAvatar   *string   `json:"autor_avatar,omitempty"`
	UserReaction  *string   `json:"user_reaction,omitempty"` // 'like', 'dislike' ou null
}

// ComentarioRequest representa a requisição para criar/atualizar um comentário
type ComentarioRequest struct {
	PalpiteID int    `json:"palpite_id" binding:"required"`
	Texto     string `json:"texto" binding:"required,min=1,max=1000"`
}
