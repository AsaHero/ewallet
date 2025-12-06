package notifications

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/utils"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type recordReminderCalculateUsecase struct {
	contextTimeout  time.Duration
	logger          *logger.Logger
	transactionRepo entities.TransactionRepository
	taskQueue       *asynq.Client
}

func NewRecordReminderCalculateUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	transactionRepo entities.TransactionRepository,
	taskQueue *asynq.Client,
) *recordReminderCalculateUsecase {
	return &recordReminderCalculateUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		transactionRepo: transactionRepo,
		taskQueue:       taskQueue,
	}
}

func (r *recordReminderCalculateUsecase) RecordReminderCalculate(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "RecordReminderCalculate",
		attribute.String("user_id", userID),
	)
	defer func() { end(nil) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error

		input.userID, err = uuid.Parse(userID)
		if err != nil {
			return inerr.NewErrValidation("user_id", err.Error())
		}
	}

	transactions, _, err := r.transactionRepo.GetByUserID(ctx, 0, 0, input.userID)
	if err != nil {
		return err
	}

	// Calculate peak moments
	hourCounts := make(map[int]int)
	categoryCounts := make(map[string]int)

	for _, t := range transactions {
		// Use the hour from the transaction's CreatedAt time
		// Assuming CreatedAt is in the same timezone as we want to schedule, or we rely on relative consistency
		hourCounts[t.CreatedAt.Hour()]++
		categoryCounts[t.Category.Name]++
	}

	type hourCount struct {
		Hour  int
		Count int
	}
	var sortedHours []hourCount
	for h, c := range hourCounts {
		sortedHours = append(sortedHours, hourCount{h, c})
	}
	sort.Slice(sortedHours, func(i, j int) bool {
		return sortedHours[i].Count > sortedHours[j].Count
	})

	var peakHours []int
	if len(sortedHours) == 0 {
		peakHours = []int{12} // Default to 12:00 PM if no transactions
	} else {
		limit := 3
		if len(sortedHours) < limit {
			limit = len(sortedHours)
		}
		for i := 0; i < limit; i++ {
			peakHours = append(peakHours, sortedHours[i].Hour)
		}
	}

	// Construct texts based on categories of transactions
	type categoryCount struct {
		Name  string
		Count int
	}
	var sortedCategories []categoryCount
	for n, c := range categoryCounts {
		sortedCategories = append(sortedCategories, categoryCount{n, c})
	}
	sort.Slice(sortedCategories, func(i, j int) bool {
		return sortedCategories[i].Count > sortedCategories[j].Count
	})

	var topCategories []string
	catLimit := 3
	if len(sortedCategories) < catLimit {
		catLimit = len(sortedCategories)
	}
	for i := 0; i < catLimit; i++ {
		topCategories = append(topCategories, sortedCategories[i].Name)
	}

	var text string
	switch utils.RandomInt(1, 3) {
	case 1:
		text = r.constructReminderText1(topCategories)
	case 2:
		text = r.constructReminderText2(topCategories)
	case 3:
		text = r.constructReminderText3(topCategories)
	}

	// Construct tasks
	now := time.Now()
	for _, h := range peakHours {
		// Calculate next occurrence of h
		// We use the current location
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), h, 0, 0, 0, now.Location())

		// If the time has already passed today, schedule for tomorrow
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		delay := nextRun.Sub(now)

		task, err := tasks.NewRecordReminderSendTask(input.userID.String(), text)
		if err != nil {
			r.logger.ErrorContext(ctx, "failed to create task", err)
			continue
		}

		if _, err := r.taskQueue.Enqueue(task, asynq.Queue("medium"), asynq.ProcessIn(delay)); err != nil {
			r.logger.ErrorContext(ctx, "failed to enqueue task", err)
		}
	}

	return nil
}

func (r *recordReminderCalculateUsecase) constructReminderText1(topCategories []string) string {
	text := "👀 А ну признавайтесь, уже потратились?"
	if len(topCategories) > 0 {
		text += fmt.Sprintf("\n\nПодозреваю: снова <i>%s</i> ☕😄", strings.Join(topCategories, ", "))
	}
	text += "\n\nЗапишите, пока кофе не сделал дырку в бюджете 😅"
	return text
}

func (r *recordReminderCalculateUsecase) constructReminderText2(topCategories []string) string {
	text := "💸 Ваши деньги снова решили прогуляться!"
	if len(topCategories) > 0 {
		text += fmt.Sprintf("\n\nИ как всегда — в направлении: <i>%s</i> 🤷‍♂️", strings.Join(topCategories, ", "))
	}
	text += "\n\nДавайте поймаем их в списочек расходов 🕵️‍♂️"
	return text
}

func (r *recordReminderCalculateUsecase) constructReminderText3(topCategories []string) string {
	text := "📝 Пора записать расходы, пока они не сбежали!"
	if len(topCategories) > 0 {
		text += fmt.Sprintf("\n\nИ да… снова <i>%s</i> 😏", strings.Join(topCategories, ", "))
	}
	text += "\n\nКошелёк ведь не резиновый… пока 😅"
	return text
}
