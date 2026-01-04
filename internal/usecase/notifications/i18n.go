// internal/usecase/notifications/i18n.go
package notifications

import (
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
)

// resolveLang returns user's language (EN fallback).
func resolveLang(u *entities.User) entities.Language {
	if u == nil {
		return entities.EN
	}
	switch u.LanguageCode {
	case entities.RU, entities.UZ, entities.EN:
		return u.LanguageCode
	default:
		return entities.EN
	}
}

// resolveLocation returns user's timezone Location (UTC fallback).
func resolveLocation(u *entities.User) *time.Location {
	loc := time.UTC
	if u == nil {
		return loc
	}
	if u.Timezone != "" {
		if l, err := time.LoadLocation(u.Timezone); err == nil {
			loc = l
		}
	}
	return loc
}

type RecordReminderVariant string

const (
	RecordNoon    RecordReminderVariant = `noon`
	RecordEvening RecordReminderVariant = `evening`
)

type DebtActionLabels struct {
	Paid        string
	RemindLater string
	Cancel      string
}

func debtActionLabels(lang entities.Language) DebtActionLabels {
	switch lang {
	case entities.RU:
		return DebtActionLabels{
			Paid:        `✅ Оплачено`,
			RemindLater: `⏳ Напомнить позже`,
			Cancel:      `❌ Отмена`,
		}
	case entities.UZ:
		return DebtActionLabels{
			Paid:        `✅ To‘landi`,
			RemindLater: `⏳ Keyinroq`,
			Cancel:      `❌ Bekor qilish`,
		}
	default:
		return DebtActionLabels{
			Paid:        `✅ Paid`,
			RemindLater: `⏳ Remind later`,
			Cancel:      `❌ Cancel`,
		}
	}
}

// Telegram Bot API inline keyboard in raw JSON-like form.
// Most bot clients can send this as ReplyMarkup (serialized).
func buildDebtInlineKeyboard(lang entities.Language, debtID string) map[string]any {
	labels := debtActionLabels(lang)

	// callback_data format: debt:<action>:<debtID>
	return map[string]any{
		`inline_keyboard`: [][]map[string]any{
			{
				{`text`: labels.Paid, `callback_data`: fmt.Sprintf(`debt:paid:%s`, debtID)},
				{`text`: labels.RemindLater, `callback_data`: fmt.Sprintf(`debt:remind_later:%s`, debtID)},
			},
			{
				{`text`: labels.Cancel, `callback_data`: fmt.Sprintf(`debt:cancel:%s`, debtID)},
			},
		},
	}
}
