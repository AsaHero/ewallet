package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const AnonsSendTaskName string = "anons:send"

type AnonsSendPayload struct {
	UserID string `json:"user_id"`
	AnonID string `json:"anon_id"`
}

func NewAnonsSendTask(userID string, anonID string) (*asynq.Task, error) {
	payload := AnonsSendPayload{
		UserID: userID,
		AnonID: anonID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(AnonsSendTaskName, data, asynq.Queue("medium")), nil
}
