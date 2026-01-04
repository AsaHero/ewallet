package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const DebtReminderCheckTaskName string = "debt_reminder:check"
const DebtReminderSendTaskName string = "debt_reminder:send"

type DebtReminderSendPayload struct {
	DebtID string `json:"debt_id"`
}

func NewDebtReminderCheckTask() (*asynq.Task, error) {
	// No payload needed for periodic check
	return asynq.NewTask(DebtReminderCheckTaskName, nil, asynq.Queue("high")), nil
}

func NewDebtReminderSendTask(debtID string) (*asynq.Task, error) {
	payload := DebtReminderSendPayload{
		DebtID: debtID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(DebtReminderSendTaskName, data, asynq.Queue("medium")), nil
}
