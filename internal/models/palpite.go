package models

import "time"

type Palpite struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Titulo    *string   `json:"titulo,omitempty"`
	ImgURL    string    `json:"img_url"`
	Avatar    *string   `json:"avatar,omitempty"`
	Link      *string   `json:"link,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PalpiteResponse struct {
	ID               int               `json:"id"`
	UserID           int               `json:"user_id"`
	UserName         string            `json:"user_name"`
	Titulo           *string           `json:"titulo,omitempty"`
	ImgURL           string            `json:"img_url"`
	Avatar           *string           `json:"avatar,omitempty"`
	Link             *string           `json:"link,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	TotalLikes       int               `json:"total_likes"`
	TotalDislikes    int               `json:"total_dislikes"`
	TotalComentarios int               `json:"total_comentarios"`
	Comentarios      []ComentarioStats `json:"comentarios,omitempty"`
}
type PalpiteStats struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	Titulo           *string   `json:"titulo,omitempty"`
	ImgURL           string    `json:"img_url"`
	Link             *string   `json:"link,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	TotalLikes       int       `json:"total_likes"`
	TotalDislikes    int       `json:"total_dislikes"`
	TotalComentarios int       `json:"total_comentarios"`
	AutorNome        string    `json:"autor_nome"`
	AutorAvatar      *string   `json:"autor_avatar,omitempty"`
	UserReaction     *string   `json:"user_reaction,omitempty"`
}

func (p *Palpite) ToResponse() PalpiteResponse {
	return PalpiteResponse{
		ID:               p.ID,
		UserID:           p.UserID,
		Titulo:           p.Titulo,
		ImgURL:           p.ImgURL,
		Avatar:           p.Avatar,
		Link:             p.Link,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		TotalLikes:       0,
		TotalDislikes:    0,
		TotalComentarios: 0,
	}
}
