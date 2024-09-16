package repository_tender_model

import "time"

type Tender struct {
	ID          string
	Name        string
	Description string
	Status      string
	ServiceType string
	Version     int
	CreatedAt   time.Time
}
