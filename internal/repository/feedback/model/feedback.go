package repository_feedback_model

import "time"

type Feedback struct {
	ID          int
	Description string
	CreatedAt   time.Time
}
