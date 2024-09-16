package repository_feedback_converter

import (
	"avito_intership/internal/model"
	repository_feedback_model "avito_intership/internal/repository/feedback/model"
)

func ToFeedbackFromRepository(feedback repository_feedback_model.Feedback) model.Feedback {
	return model.Feedback{
		ID:          feedback.ID,
		Description: feedback.Description,
		CreatedAt:   feedback.CreatedAt,
	}
}
