package models

import "time"

// PalpiteReaction representa uma reação (like/dislike) em um palpite
type PalpiteReaction struct {
	ID        int       `json:"id"`
	PalpiteID int       `json:"palpite_id"`
	UserID    int       `json:"user_id"`
	Tipo      string    `json:"tipo"` // 'like' ou 'dislike'
	CreatedAt time.Time `json:"created_at"`
}

// ComentarioReaction representa uma reação (like/dislike) em um comentário
type ComentarioReaction struct {
	ID           int       `json:"id"`
	ComentarioID int       `json:"comentario_id"`
	UserID       int       `json:"user_id"`
	Tipo         string    `json:"tipo"` // 'like' ou 'dislike'
	CreatedAt    time.Time `json:"created_at"`
}

// ReactionRequest representa a requisição para dar like/dislike
type ReactionRequest struct {
	Tipo string `json:"tipo" binding:"required,oneof=like dislike"`
}

// ReactionResponse representa a resposta após toggle de reação
type ReactionResponse struct {
	Action        string `json:"action"` // 'added', 'removed', 'changed'
	TotalLikes    int64  `json:"total_likes"`
	TotalDislikes int64  `json:"total_dislikes"`
}
