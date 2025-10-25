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
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Titulo    *string   `json:"titulo,omitempty"`
	ImgURL    string    `json:"img_url"`
	Avatar    *string   `json:"avatar,omitempty"`
	Link      *string   `json:"link,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Palpite) ToResponse() PalpiteResponse {
	return PalpiteResponse{
		ID:        p.ID,
		UserID:    p.UserID,
		Titulo:    p.Titulo,
		ImgURL:    p.ImgURL,
		Avatar:    p.Avatar,
		Link:      p.Link,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
