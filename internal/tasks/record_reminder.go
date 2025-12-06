package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const RecordReminderCalculateTaskName string = "record_reminder:calculate"
const RecordReminderSendTaskName string = "record_reminder:send"

type RecordReminderCalculatePayload struct {
	UserID string `json:"user_id"`
}

func NewRecordReminderCalculateTask(userID string) (*asynq.Task, error) {
	payload := RecordReminderCalculatePayload{
		UserID: userID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(RecordReminderCalculateTaskName, data), nil
}

type RecordReminderSendPayload struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func NewRecordReminderSendTask(userID string, text string) (*asynq.Task, error) {
	payload := RecordReminderSendPayload{
		UserID: userID,
		Text:   text,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(RecordReminderSendTaskName, data), nil
}
