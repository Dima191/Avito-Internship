package repository_bid_model

import "time"

type Bid struct {
	ID         string
	Name       string
	Status     string
	AuthorType string
	AuthorID   string
	Version    int
	CreatedAt  time.Time
}
