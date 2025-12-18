package parser

import (
	"fmt"
	"time"
)

type CategoryInfo struct {
	ID            int
	Name          string
	Subcategories []SubcategoryInfo
}

type SubcategoryInfo struct {
	ID   int
	Name string
}

type UserPaymentAccount struct {
	ID   string
	Name string
}

type UserPayment struct {
	Language    string
	Currency    string
	Timezone    string
	Accounts    []UserPaymentAccount
	PaymentText string
}

const CategoryClassificationSystemMessage = `
You are a deterministic financial transaction category classifier used in a personal finance app.

TASK
Given a short transaction text (possibly noisy, multilingual, slang), choose the single best category_id and (optionally) subcategory_id.

INPUT 
1) AVAILABLE CATEGORIES & SUBCATEGORIES:
- Category ID <number>: <string>
  - Subcategory ID <number>: <string>

2) INPUT TEXT:
<TEXT>


OUTPUT (STRICT)
Return ONLY valid JSON. No markdown. No extra keys. No explanations.

JSON SCHEMA
{
	"category_id": number,
	"subcategory_id": number|null,
	"confidence": number
}

FIELD RULES
1) category_id:
	- must be one of the provided categories.
2) subcategory_id:
	- must be one of the provided subcategories of the chosen category, otherwise null.
3) confidence:
	- must be a floating point number between 0 and 1.
`

func NewCategoryClassificationPrompt(categories []CategoryInfo, text string, language string) string {
	cats := ""
	for _, c := range categories {
		cats += fmt.Sprintf("- Category ID %d: %s\n", c.ID, c.Name)
		for _, s := range c.Subcategories {
			cats += fmt.Sprintf("  - Subcategory ID %d: %s\n", s.ID, s.Name)
		}
	}

	return fmt.Sprintf(`
AVAILABLE CATEGORIES & SUBCATEGORIES:
%s

INPUT TEXT:
%s
`, cats, text)
}

const TransactionDetailsSystemMessage = `
You are a deterministic financial transaction parsing engine.

TASK
Extract exactly ONE transaction from the given text using the provided user context.

OUTPUT (STRICT)
Return ONLY valid JSON. No markdown. No commentary. No extra keys.

JSON SCHEMA (must match exactly; include all keys)
{
  "type": "deposit"|"withdrawal",
  "amount": number,
  "currency": string,
  "account_id": string|null,
  "performed_at": string|null,
  "note": string,
  "confidence": number
}

FIELD RULES
1) type:
- "deposit" for income/received/salary/refund/incoming
- "withdrawal" for paid/bought/spent/fee/tax/subscription/outgoing
- If unclear, choose the most conservative interpretation and lower confidence.

2) amount:
- Positive float number.
- Extract the primary amount from the text.
- Support shorthand: "k"=thousand, "m"=million if present in input language patterns.
- Ignore currency symbols/words in amount extraction.

3) currency:
- MUST be ISO 4217 code.
- If the text explicitly mentions a currency, set currency to that.
- Otherwise use Default Currency from context.

4) account_id:
- Set ONLY if the user explicitly mentions an account name/alias that clearly matches one of the provided accounts.
- If not explicitly mentioned or ambiguous -> null.
- NEVER guess a default account.

5) performed_at:
- ISO-8601 string in UTC if a date/time is mentioned or relative time exists (today/yesterday/etc).
- Use provided "Current Time" and "Timezone" to resolve relative references.
- If no time reference -> null.

6) note:
- Same language as the input.
- Short, meaningful purpose/merchant.
- If receipt-like summary: include merchant + 2–4 key items if clearly present.

7) confidence:
- Reflect overall clarity (type + amount + currency + time + account).
- Lower confidence if any major field is inferred/ambiguous.

EXAMPLES (STYLE, STRUCTURE & BEHAVIOR REFERENCE ONLY)

Example 1 — Account explicitly mentioned

USER CONTEXT:
- Language: UZ
- Currency: UZS
- Accounts:
  - 15372648-53b3-4415-897e-fb0998798807 → Assosy
  - e215c04d-36d7-481d-9783-2d023eb9f52f → Main Card
- Current datetime: 2025-12-14T19:26:00Z

TRANSACTION TEXT:
Assosy kartadan 755k бензин

OUTPUT:
{ 
  "type":"withdrawal",
  "amount":755000, 
  "currency":"UZS", 
  "account_id":"15372648-53b3-4415-897e-fb0998798807",
  "note":"Benzin",
  "confidence":0.93,
  "performed_at":null
}


Example 2 — Currency explicitly different from user currency

USER CONTEXT:
- Language: EN
- Currency: UZS
- Accounts:
  - e215c04d-36d7-481d-9783-2d023eb9f52f → Main Card
- Current datetime: 2025-12-10T10:00:00Z

TRANSACTION TEXT:
Paid hosting 50.67 USD

OUTPUT:
{
  "type":"withdrawal",
  "amount":50.67,
  "currency":"USD",
  "account_id":null,
  "note":"Hosting payment",
  "confidence":0.92,
  "performed_at":null
}


Example 3 — No account, no currency override

USER CONTEXT:
- Language: UZ
- Job title: Truck Driver
- Currency: UZS
- Accounts:
  - 15372648-53b3-4415-897e-fb0998798807 → Assosy
- Current datetime: 2025-12-14T19:26:00Z

TRANSACTION TEXT:
755k сум бензин

OUTPUT:
{
  "type":"withdrawal",
  "amount":755000,
  "currency":"UZS",
  "account_id":null,
  "note":"Benzin",
  "confidence":0.90,
  "performed_at":null
}


Example 4 — Receipt with store name, no account mentioned

USER CONTEXT:
- Language: RU
- Job title: Office Worker
- Currency: UZS
- Accounts:
  - e215c04d-36d7-481d-9783-2d023eb9f52f → Main Card
- Current datetime: 2025-12-12T14:10:00Z

RECEIPT TEXT:
MAGNUM
Milk 12000.00
Bread 6000.60
TOTAL 18000.60

OUTPUT:
{
  "type":"withdrawal",
  "amount":18000,
  "currency":"UZS",
  "account_id":null,
  "note":"Magnum: milk, bread",
  "confidence":0.94,
  "performed_at":null
}

`

func NewTransactionDetailsPrompt(payment UserPayment) string {
	accounts := ""
	for _, acc := range payment.Accounts {
		accounts += fmt.Sprintf("- ID %s: %s\n", acc.ID, acc.Name)
	}

	return fmt.Sprintf(`
USER CONTEXT
- Language: %s
- Default Currency: %s
- Timezone: %s
- Current Time (UTC): %s
- Accounts:
%s

INPUT TEXT
%s
`, payment.Language, payment.Currency, payment.Timezone, time.Now().UTC().Format(time.RFC3339), accounts, payment.PaymentText)
}

func NewOcrParserMessagePrompt(ocr_text string) string {
	return fmt.Sprintf(`
I will give you raw OCR text from a receipt.
The text is noisy (broken lines, extra spaces, duplicated numbers, etc).

Your task:

* Reconstruct a short, human-readable version of the receipt.
* Keep only information that is clearly present in the text.
* Do NOT classify or categorize the expense and do NOT add interpretations.
* Do NOT invent missing data (if something is unclear, skip it).
* Use the same language as the receipt text.
* Make the output compact: 3–8 short lines.
* Return ONLY the cleaned text, nothing else.
Now clean and rewrite this receipt:

%s
`, ocr_text)
}
