// internal/usecase/notifications/templates.go
package notifications

import (
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
)

type DebtMsgParams struct {
	DebtType entities.DebtType
	Amount   float64
	Currency string
	RemindAt time.Time
	Loc      *time.Location
}

func debtReminderText(lang entities.Language, p DebtMsgParams) string {
	// local time formatting
	t := p.RemindAt
	if p.Loc != nil {
		t = t.In(p.Loc)
	}
	when := t.Format(`2006-01-02 15:04`)

	// labels depend on debt type
	typeEmoji := `📥` // borrow
	actionEN := `repay`
	actionRU := `вернуть`
	actionUZ := `qaytarish`

	if p.DebtType == entities.Lend {
		typeEmoji = `📤`
		actionEN = `collect`
		actionRU = `забрать`
		actionUZ = `undirish`
	}

	switch lang {
	case entities.RU:
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(
				`⏰ <b>Напоминание о долге</b> %s

Мягко напомню: ты планировал(а) <b>%s</b> этот долг 💛

💰 <b>%.2f %s</b>
🗓 %s

Выбери действие ниже 👇`,
				typeEmoji,
				actionRU,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(
			`⏰ <b>Напоминание о долге</b> %s

Тёплый пинг без спешки: кажется, пора <b>%s</b> этот долг 💛

💰 <b>%.2f %s</b>
🗓 %s

Выбери действие ниже 👇`,
			typeEmoji,
			actionRU,
			p.Amount,
			p.Currency,
			when,
		)

	case entities.UZ:
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(
				`⏰ <b>Qarz eslatmasi</b> %s

Yengilgina eslatma: bu qarzni <b>%s</b> rejangiz bor edi 💛

💰 <b>%.2f %s</b>
🗓 %s

Quyidan tanlang 👇`,
				typeEmoji,
				actionUZ,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(
			`⏰ <b>Qarz eslatmasi</b> %s

Shoshmasdan, iliq eslatma: bu qarzni <b>%s</b> vaqti kelganga o'xshaydi 💛

💰 <b>%.2f %s</b>
🗓 %s

Quyidan tanlang 👇`,
			typeEmoji,
			actionUZ,
			p.Amount,
			p.Currency,
			when,
		)

	default: // EN
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(
				`⏰ <b>Debt reminder</b> %s

A gentle nudge: you planned to <b>%s</b> this debt 💛

💰 <b>%.2f %s</b>
🗓 %s

Choose an action below 👇`,
				typeEmoji,
				actionEN,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(
			`⏰ <b>Debt reminder</b> %s

Just a friendly ping: it might be time to <b>%s</b> this one 💛

💰 <b>%.2f %s</b>
🗓 %s

Choose an action below 👇`,
			typeEmoji,
			actionEN,
			p.Amount,
			p.Currency,
			when,
		)
	}
}

func recordReminderText(lang entities.Language, variant RecordReminderVariant) string {
	switch lang {
	case entities.RU:
		if variant == RecordEvening {
			return `🌙 <b>Вечерняя проверка</b>

Если сегодня были траты — самое время спокойно закрыть день.
Текст, голос или чек 🧾🎙️

Пара сообщений — и готово ✅`
		}
		return `📝 <b>Небольшой толчок</b>

Хочешь, чтобы статистика сегодня была "чистая"?
Просто отправь мне:
• текст (например: «кофе 25к»)
• голос 🎙️
• или фото чека 🧾

Остальное я соберу сам ✨`

	case entities.UZ:
		if variant == RecordEvening {
			return `🌙 <b>Kechki tekshiruv</b>

Bugun xarajat bo'lgan bo'lsa — hozir yuborib qo'ying.
Matn, ovoz yoki chek rasmi 🧾🎙️

Ikki xabar — tamom ✅`
		}
		return `📝 <b>Kichik eslatma</b>

Bugun statistikangiz "toza" bo'lsinmi?
Shunchaki yuboring:
• matn (masalan: "qahva 25 ming")
• ovoz 🎙️
• yoki chek rasmi 🧾

Qolganini men qilaman ✨`

	default: // EN
		if variant == RecordEvening {
			return `🌙 <b>Evening check-in</b>

If you had any expenses today — now's a perfect time to drop them here.
Text, voice, or receipt photo 🧾🎙️

Two messages and you're done ✅`
		}
		return `📝 <b>Small nudge</b>

Want your stats to stay clean today?
Send me:
• a quick text ("coffee 25k")
• a voice note 🎙️
• or a receipt photo 🧾

I'll do the rest ✨`
	}
}
