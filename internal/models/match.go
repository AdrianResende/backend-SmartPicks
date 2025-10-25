package models

import "time"

type Match struct {
	ID        int       `json:"id"`
	TeamA     string    `json:"team_a"`
	TeamB     string    `json:"team_b"`
	MatchDate time.Time `json:"match_date"`
	Location  string    `json:"location"`
}
