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
	userRepo        entities.UserRepository
	taskQueue       *asynq.Client
}

func NewRecordReminderCalculateUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	transactionRepo entities.TransactionRepository,
	userRepo entities.UserRepository,
	taskQueue *asynq.Client,
) *recordReminderCalculateUsecase {
	return &recordReminderCalculateUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		taskQueue:       taskQueue,
	}
}

func (r *recordReminderCalculateUsecase) RecordReminderCalculate(ctx context.Context, userID string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "RecordReminderCalculate",
		attribute.String("user_id", userID),
	)
	defer func() { end(err) }()

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

	user, err := r.userRepo.FindByID(ctx, input.userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	loc := time.UTC
	if user.Timezone != "" {
		if l, err := time.LoadLocation(user.Timezone); err == nil {
			loc = l
		}
	}

	now := time.Now().In(loc)
	from := now.Add(-60 * 24 * time.Hour)
	transactions, err := r.transactionRepo.GetAllBetween(ctx, input.userID, from.UTC(), time.Now().UTC())
	if err != nil {
		return err
	}

	// 1. Calculate Expenses Reminders
	if err := r.processExpenseReminders(ctx, transactions, user, now, loc); err != nil {
		r.logger.ErrorContext(ctx, "failed to process expense reminders", err)
	}

	// 2. Calculate Income Reminders
	if err := r.processIncomeReminders(ctx, transactions, user, now, loc); err != nil {
		r.logger.ErrorContext(ctx, "failed to process income reminders", err)
	}

	return nil
}

func (r *recordReminderCalculateUsecase) processExpenseReminders(
	ctx context.Context,
	allTransactions []*entities.Transaction,
	user *entities.User,
	now time.Time,
	loc *time.Location,
) error {
	// Filter: Last 14 days, Withdrawal, Weekday matches Now
	cutoff := now.Add(-14 * 24 * time.Hour)
	var filtered []*entities.Transaction

	for _, t := range allTransactions {
		tTime := t.CreatedAt.In(loc)
		if t.Type == entities.Withdrawal &&
			tTime.After(cutoff) &&
			tTime.Weekday() == now.Weekday() {
			filtered = append(filtered, t)
		}
	}

	return r.scheduleReminders(ctx, filtered, user, now, loc, false)
}

func (r *recordReminderCalculateUsecase) processIncomeReminders(
	ctx context.Context,
	allTransactions []*entities.Transaction,
	user *entities.User,
	now time.Time,
	loc *time.Location,
) error {
	// Filter: Last 60 days (already fetched), Deposit, Day matches Now
	// Note: We use the already fetched range which is 60 days.

	var filtered []*entities.Transaction
	for _, t := range allTransactions {
		tTime := t.CreatedAt.In(loc)
		if t.Type == entities.Deposit &&
			tTime.Day() == now.Day() {
			filtered = append(filtered, t)
		}
	}

	return r.scheduleReminders(ctx, filtered, user, now, loc, true)
}

func (r *recordReminderCalculateUsecase) scheduleReminders(
	ctx context.Context,
	transactions []*entities.Transaction,
	user *entities.User,
	now time.Time,
	loc *time.Location,
	isIncome bool,
) error {
	if len(transactions) == 0 {
		return nil
	}

	// Calculate peak 2-hour windows
	// Map window start hour (0, 2, ..., 22) -> count
	windowCounts := make(map[int]int)
	categoryCounts := make(map[string]int)

	for _, t := range transactions {
		h := t.CreatedAt.In(loc).Hour()
		windowStart := (h / 2) * 2 // 0, 2, 4...
		windowCounts[windowStart]++
		if t.Category.GetName(user.LanguageCode) != "" {
			categoryCounts[t.Category.GetName(user.LanguageCode)]++
		}
	}

	// Identify peaks with frequency >= 2 (Income) or >= 1 (Expense)
	var peakWindows []int
	for w, count := range windowCounts {
		if isIncome && count >= 2 {
			peakWindows = append(peakWindows, w)
		}

		if !isIncome && count >= 1 {
			peakWindows = append(peakWindows, w)
		}
	}

	if len(peakWindows) == 0 {
		return nil
	}
	sort.Ints(peakWindows)

	// Prepare text
	topCategories := r.getTopCategories(categoryCounts)
	var text string
	if isIncome {
		text = r.constructIncomeReminderText(topCategories)
	} else {
		text = r.constructExpenseReminderText(topCategories)
	}

	// Schedule tasks
	for _, w := range peakWindows {

		// Target time: Start of window + random offset (0-60 mins)
		randomOffset := utils.RandomInt(0, 60)

		nextRun := time.Date(now.Year(), now.Month(), now.Day(), w, 0, 0, 0, loc).Add(time.Duration(randomOffset) * time.Minute)

		// If time passed, skip for today. Do NOT schedule for tomorrow (as per user request).
		if nextRun.Before(now) {
			continue
		}

		delay := nextRun.Sub(now)

		task, err := tasks.NewRecordReminderSendTask(user.ID.String(), text)
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

func (r *recordReminderCalculateUsecase) getTopCategories(counts map[string]int) []string {
	type catCount struct {
		Name  string
		Count int
	}
	var sorted []catCount
	for n, c := range counts {
		sorted = append(sorted, catCount{n, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	var top []string
	limit := 3
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		top = append(top, sorted[i].Name)
	}
	return top
}

func (r *recordReminderCalculateUsecase) constructExpenseReminderText(topCategories []string) string {
	// Pick random template
	switch utils.RandomInt(1, 3) {
	case 1:
		return r.constructReminderText1(topCategories)
	case 2:
		return r.constructReminderText2(topCategories)
	default:
		return r.constructReminderText3(topCategories)
	}
}

func (r *recordReminderCalculateUsecase) constructIncomeReminderText(topCategories []string) string {
	text := "💰 Кажется, сегодня день получки?"
	text += "\n\nНе забудьте внести доход, чтобы мы могли посчитать, на что вы будете шиковать! 😎"
	return text
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
