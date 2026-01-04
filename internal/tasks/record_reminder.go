package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const RecordReminderScheduleTaskName string = "record_reminder:schedule"
const RecordReminderSendTaskName string = "record_reminder:send"

func NewRecordReminderScheduleTask() (*asynq.Task, error) {
	return asynq.NewTask(RecordReminderScheduleTaskName, nil, asynq.Queue("high")), nil
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

	return asynq.NewTask(RecordReminderSendTaskName, data, asynq.Queue("medium")), nil
}
