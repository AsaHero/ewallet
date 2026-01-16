// internal/usecase/notifications/templates.go
package notifications

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
)

// NOTE:
// - All strings are HTML-ready (for Telegram parse_mode=HTML).
// - Uses small randomized variants to avoid "banner blindness".
// - Tone gets slightly more direct if reminder time is overdue.
// - Includes debtor/creditor name in the message.

type DebtMsgParams struct {
	DebtType entities.DebtType
	Name     string
	Amount   float64
	Currency string
	RemindAt time.Time
	Loc      *time.Location
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func debtReminderText(lang entities.Language, p DebtMsgParams) string {
	// local time formatting
	t := p.RemindAt
	if p.Loc != nil {
		t = t.In(p.Loc)
	}
	when := t.Format(`2006-01-02 15:04`)

	// now in same loc (for "overdue" tone)
	now := time.Now()
	if p.Loc != nil {
		now = now.In(p.Loc)
	}

	name := normalizeName(lang, p.Name)

	// labels depend on debt type
	typeEmoji := `📥` // borrow (you owe, repay)
	actionEN := `repay`
	actionRU := `вернуть`
	actionUZ := `qaytarish`

	if p.DebtType == entities.Lend { // you lent, collect
		typeEmoji = `📤`
		actionEN = `collect`
		actionRU = `забрать`
		actionUZ = `undirish`
	}

	// Tone level: 0 soft (upcoming/on time), 1 direct (a bit overdue), 2 playful (overdue longer)
	tone := toneLevel(now, t)

	// one-liner about amount tier (tiny touch, not spammy)
	amountLine := amountHint(lang, p.Amount)

	// Build message templates per language/debt type/tone
	switch lang {
	case entities.EN:
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(pickOne(enLend[tone]),
				typeEmoji,
				name,
				actionEN,
				amountLine,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(pickOne(enBorrow[tone]),
			typeEmoji,
			name,
			actionEN,
			amountLine,
			p.Amount,
			p.Currency,
			when,
		)

	case entities.UZ:
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(pickOne(uzLend[tone]),
				typeEmoji,
				name,
				actionUZ,
				amountLine,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(pickOne(uzBorrow[tone]),
			typeEmoji,
			name,
			actionUZ,
			amountLine,
			p.Amount,
			p.Currency,
			when,
		)

	default: // RU
		if p.DebtType == entities.Lend {
			return fmt.Sprintf(pickOne(ruLend[tone]),
				typeEmoji,
				name,
				actionRU,
				amountLine,
				p.Amount,
				p.Currency,
				when,
			)
		}
		return fmt.Sprintf(pickOne(ruBorrow[tone]),
			typeEmoji,
			name,
			actionRU,
			amountLine,
			p.Amount,
			p.Currency,
			when,
		)
	}
}

func recordReminderText(lang entities.Language, variant RecordReminderVariant) string {
	switch lang {
	case entities.EN:
		if variant == RecordEvening {
			return pickOne(enRecordEvening)
		}
		return pickOne(enRecordDay)

	case entities.UZ:
		if variant == RecordEvening {
			return pickOne(uzRecordEvening)
		}
		return pickOne(uzRecordDay)

	default: // RU
		if variant == RecordEvening {
			return pickOne(ruRecordEvening)
		}
		return pickOne(ruRecordDay)
	}
}

/* ------------------------------ helpers ------------------------------ */

func pickOne(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[rand.Intn(len(list))]
}

func normalizeName(lang entities.Language, name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		switch lang {
		case entities.EN:
			return "that person"
		case entities.UZ:
			return "u odam"
		default:
			return "этот человек"
		}
	}
	// Keep it short-ish (Telegram UX)
	if len([]rune(n)) > 32 {
		r := []rune(n)
		return string(r[:32]) + "…"
	}
	return n
}

// toneLevel decides how "direct/playful" the reminder is.
// 0: soft (not overdue yet or just on time)
// 1: more direct (overdue > 2h)
// 2: playful/direct (overdue > 24h)
func toneLevel(now, remindAt time.Time) int {
	if now.Before(remindAt) || now.Equal(remindAt) {
		return 0
	}
	over := now.Sub(remindAt)
	if over > 24*time.Hour {
		return 2
	}
	if over > 2*time.Hour {
		return 1
	}
	return 0
}

func amountHint(lang entities.Language, amount float64) string {
	// These are intentionally simple and currency-agnostic.
	// Tune thresholds based on your product data later.
	switch {
	case amount <= 0:
		return ""
	case amount < 50000:
		switch lang {
		case entities.EN:
			return "Small one — quick win 😄"
		case entities.UZ:
			return "Kichkina summa — tez yopiladi 😄"
		default:
			return "Мелочь — можно закрыть быстро 😄"
		}
	case amount < 500000:
		switch lang {
		case entities.EN:
			return "Worth keeping tidy 🙂"
		case entities.UZ:
			return "Nazoratda bo‘lgani yaxshi 🙂"
		default:
			return "Лучше держать под контролем 🙂"
		}
	default:
		switch lang {
		case entities.EN:
			return "This one’s serious — better not drag it 👀"
		case entities.UZ:
			return "Bu jiddiyroq — cho‘zmaslik yaxshi 👀"
		default:
			return "Сумма серьёзная — лучше не тянуть 👀"
		}
	}
}

/* ------------------------------ templates ------------------------------ */

// Format params (for debt templates):
// 1) typeEmoji
// 2) name
// 3) action (repay/collect, вернуть/забрать, qaytarish/undirish)
// 4) amountHint line
// 5) amount
// 6) currency
// 7) when

var ruLend = [][]string{
	{ // tone 0 (soft)
		`⏰ <b>Долг напоминает о себе</b> %s

<b>%s</b> на связи 👀
Похоже, пора аккуратно <b>%s</b> этот долг. Без неловкостей 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Как поступим? 👇`,
		`⏰ <b>Мягкое напоминание</b> %s

Есть должок у <b>%s</b>.
Когда будет удобно — можно <b>%s</b> и закрыть вопрос ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбирай действие 👇`,
		`⏰ <b>Пинг без спешки</b> %s

<b>%s</b> всё ещё в списке.
Если сегодня будет момент — самое время <b>%s</b> 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Что делаем? 👇`,
	},
	{ // tone 1 (direct)
		`⏰ <b>Пора заняться долгом</b> %s

<b>%s</b> всё ещё ждёт.
Давай <b>%s</b> и снимем это с головы 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбери вариант 👇`,
		`⏰ <b>Напоминание</b> %s

Долг у <b>%s</b> пока открыт.
Самое время <b>%s</b> — будет легче ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Жми кнопку ниже 👇`,
	},
	{ // tone 2 (playful)
		`😅 <b>Кажется, мы уже друзья</b> %s

<b>%s</b> и этот долг всё ещё вместе.
Может, отправим его в архив и <b>%s</b>? 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Как решаем? 👇`,
		`👀 <b>Проверка на взрослость</b> %s

Долг у <b>%s</b> живёт тут дольше, чем хотелось бы.
Давай <b>%s</b> и закроем красиво ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбирай 👇`,
	},
}

var ruBorrow = [][]string{
	{ // tone 0
		`⏰ <b>Небольшой пинг</b> %s

<b>%s</b> всё ещё ждёт 🙂
Кажется, сейчас самое время <b>%s</b> долг и закрыть тему ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Что делаем? 👇`,
		`⏰ <b>Мягко напомню</b> %s

Долг перед <b>%s</b> пока открыт.
Когда удобно — можно <b>%s</b> и выдохнуть 😌

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбирай действие 👇`,
		`⏰ <b>Без давления</b> %s

<b>%s</b> и этот долг всё ещё знакомы.
Если будет момент — можно <b>%s</b> 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Ниже варианты 👇`,
	},
	{ // tone 1
		`⏰ <b>Пора закрыть хвостик</b> %s

Долг перед <b>%s</b> всё ещё висит.
Давай <b>%s</b> и освободим голову 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбери вариант 👇`,
		`⏰ <b>Напоминание</b> %s

<b>%s</b> ждёт — и долг тоже.
Самое время <b>%s</b> ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Жми кнопку ниже 👇`,
	},
	{ // tone 2
		`😄 <b>Этот долг уже прописался</b> %s

<b>%s</b> ждёт, а долг чувствует себя как дома 🙂
Может, пора <b>%s</b> и попрощаться? ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Как поступим? 👇`,
		`👀 <b>Проверка: закрываем?</b> %s

Долг перед <b>%s</b> уже долго тут.
Давай <b>%s</b> и закроем красиво 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Выбирай 👇`,
	},
}

var enLend = [][]string{
	{ // tone 0
		`⏰ <b>Debt check-in</b> %s

<b>%s</b> is still on the list 👀
When you have a moment, you can <b>%s</b> and keep things tidy 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

What’s the plan? 👇`,
		`⏰ <b>Quick nudge</b> %s

Hey 👋 <b>%s</b> still owes you.
No rush — just a good time to <b>%s</b> ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Pick an option 👇`,
	},
	{ // tone 1
		`⏰ <b>Time to handle it</b> %s

<b>%s</b> is still pending.
Let’s <b>%s</b> and move on 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Choose below 👇`,
	},
	{ // tone 2
		`😅 <b>This debt feels at home</b> %s

<b>%s</b> and this debt have been together for a while 🙂
Maybe it’s time to <b>%s</b> and archive it ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

What do we do? 👇`,
	},
}

var enBorrow = [][]string{
	{ // tone 0
		`⏰ <b>Friendly reminder</b> %s

Looks like <b>%s</b> is still waiting 😊
Maybe it’s time to <b>%s</b> and close this chapter ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Pick an option 👇`,
		`⏰ <b>Soft ping</b> %s

Your debt with <b>%s</b> is still open.
When it’s convenient, you can <b>%s</b> 😌

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Choose below 👇`,
	},
	{ // tone 1
		`⏰ <b>Let’s close the loop</b> %s

Debt with <b>%s</b> is still pending.
Time to <b>%s</b> and get it off your mind 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Select an action 👇`,
	},
	{ // tone 2
		`😄 <b>This debt got comfy</b> %s

<b>%s</b> is waiting, and the debt feels like it lives here 🙂
Maybe it’s time to <b>%s</b> and say goodbye ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

What’s next? 👇`,
	},
}

var uzLend = [][]string{
	{ // tone 0
		`⏰ <b>Qarz eslatmasi</b> %s

<b>%s</b> hali ro‘yxatda 👀
Vaqtingiz bo‘lsa, bu qarzni <b>%s</b> va masalani yopish mumkin 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Qanday qilamiz? 👇`,
		`⏰ <b>Kichkina ping</b> %s

<b>%s</b> sizdan qarzdor.
Shoshilmasdan — qulay paytda <b>%s</b> qilib qo‘ysangiz bo‘ladi ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Tanlang 👇`,
	},
	{ // tone 1
		`⏰ <b>Vaqti keldi</b> %s

<b>%s</b> bo‘yicha qarz hali ochiq.
Keling, <b>%s</b> qilib, yengil tortamiz 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Quyidan tanlang 👇`,
	},
	{ // tone 2
		`😅 <b>Bu qarz “o‘rnashib” qoldi</b> %s

<b>%s</b> va bu qarz ancha vaqtdan beri shu yerda 🙂
Balki <b>%s</b> qilib, chiroyli yoparmiz? ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Nima qilamiz? 👇`,
	},
}

var uzBorrow = [][]string{
	{ // tone 0
		`⏰ <b>Muloyim eslatma</b> %s

<b>%s</b> hali kutyapti 😊
Qulay payt bo‘lsa, bu qarzni <b>%s</b> qilib, masalani yopish mumkin ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Tanlang 👇`,
		`⏰ <b>Kichkina eslatma</b> %s

<b>%s</b> bilan qarz hali ochiq.
Shoshilmasdan — imkon bo‘lsa <b>%s</b> 😌

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Quyidan tanlang 👇`,
	},
	{ // tone 1
		`⏰ <b>Yopib qo‘yamizmi?</b> %s

<b>%s</b> bo‘yicha qarz hali turibdi.
Keling, <b>%s</b> qilib, boshni bo‘shatamiz 🙂

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Harakatni tanlang 👇`,
	},
	{ // tone 2
		`😄 <b>Bu qarz uy qilib oldi</b> %s

<b>%s</b> kutyapti, qarz esa bu yerda “o‘zini uyidagidek” his qilyapti 🙂
Balki <b>%s</b> qilib, xayrlashamiz? ✨

<i>%s</i>

💰 <b>%.2f %s</b>
🗓 %s

Nima qilamiz? 👇`,
	},
}

/* ----------------------- record reminder variants ---------------------- */

var ruRecordDay = []string{
	`📝 <b>Я быстро, обещаю</b> 👋

Если сегодня были траты — кинь сейчас.
Текст, голос или чек 🧾🎙️

Я сохраню и исчезну ✨`,
	`😌 <b>Чтобы потом не вспоминать</b>

Один текст / один чек — и свободен(а).
Текст, голос или фото чека 🧾🎙️

Я разберусь ✨`,
	`🧾 <b>Мини-пинг</b>

Были траты? Скинь как удобно:
• текст («кофе 25к»)
• голос 🎙️
• чек 🧾

Дальше я сам ✨`,
}

var ruRecordEvening = []string{
	`🌙 <b>Закрываем день?</b>

Последний штрих перед отдыхом 😌
Если были траты — пришли их сюда: текст, голос или чек 🧾🎙️

Я сохраню, ты — отдыхаешь 😴`,
	`🌙 <b>Вечерний чек-ин</b>

Если сегодня что-то потратил(а) — закинь сейчас.
Текст / голос / чек 🧾🎙️

Пара сообщений — и можно спать 😴`,
	`🌙 <b>Финалочка на сегодня</b>

Траты были? Скинь как есть.
Я всё аккуратно запишу ✨

Текст, голос или чек 🧾🎙️`,
}

var enRecordDay = []string{
	`📝 <b>Quick check-in</b> 👋

Spent anything today?
Send: text, voice, or a receipt 🧾🎙️

I’ll do the boring part ✨`,
	`😌 <b>Keep it clean</b>

One message and you’re done:
• text (“coffee 25k”)
• voice 🎙️
• receipt photo 🧾

I’ve got you ✨`,
	`🧾 <b>Small nudge</b>

If there were expenses — drop them here.
Text / voice / receipt 🧾🎙️

Done in seconds ✅`,
}

var enRecordEvening = []string{
	`🌙 <b>Evening check-in</b>

Before you rest — any expenses today?
Text, voice, or receipt 🧾🎙️

Two messages. Done. Sleep well 😴`,
	`🌙 <b>Wrap up the day</b>

If you spent anything — send it now.
I’ll save it neatly ✨

Text / voice / receipt 🧾🎙️`,
	`🌙 <b>Last step</b>

Expenses today?
Drop them here — I’ll handle the rest ✅

Text, voice, or receipt 🧾🎙️`,
}

var uzRecordDay = []string{
	`📝 <b>Kichkina eslatma</b> 👋

Bugun xarajat bo‘ldimi?
Matn, ovoz yoki chek yuboring 🧾🎙️

Qolganini men qilaman ✨`,
	`😌 <b>Statistika toza bo‘lsin</b>

Bitta xabar — tamom:
• matn (“qahva 25 ming”)
• ovoz 🎙️
• chek rasmi 🧾

Men tartiblab qo‘yaman ✨`,
	`🧾 <b>Mini ping</b>

Xarajat bo‘lsa — yuboring.
Matn / ovoz / chek 🧾🎙️

Bir necha soniya ✅`,
}

var uzRecordEvening = []string{
	`🌙 <b>Kechki tekshiruv</b>

Dam olishdan oldin — bugun xarajat bo‘ldimi?
Matn, ovoz yoki chek 🧾🎙️

Ikki xabar — va uxlash 😴`,
	`🌙 <b>Kunni yopamizmi?</b>

Bugun nimadir sarflangan bo‘lsa — hozir yuboring.
Men chiroyli qilib saqlayman ✨

Matn / ovoz / chek 🧾🎙️`,
	`🌙 <b>Oxirgi qadam</b>

Bugungi xarajatlar bormi?
Yuboring — qolganini men ✅

Matn, ovoz yoki chek 🧾🎙️`,
}
