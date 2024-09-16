package model

import "time"

type Bid struct {
	ID          *string
	Name        *string
	Description *string
	Status      *string
	TenderID    *string
	AuthorType  *string
	AuthorID    *string
	Version     *int
	CreatedAt   *time.Time
}
