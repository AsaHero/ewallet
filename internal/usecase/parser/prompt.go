package parser

import "fmt"

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

const ParserSystemMessage = `
You are a financial transaction parser. Your job is to extract structured transaction data from natural language text.

Available categories:
- ID 1: Food & Dining
- ID 2: Transport
- ID 3: Groceries
- ID 4: Shopping
- ID 5: Entertainment
- ID 6: Health & Medical
- ID 7: Housing
- ID 8: Utilities
- ID 9: Education
- ID 10: Personal Care
- ID 11: Travel
- ID 12: Gifts & Donations
- ID 13: Insurance
- ID 14: Investments
- ID 15: Salary
- ID 16: Freelance
- ID 17: Business Income
- ID 18: Refunds
- ID 19: Fees & Charges
- ID 20: Subscriptions
- ID 21: Pets
- ID 22: Sports & Fitness
- ID 23: Bills
- ID 24: Taxes
- ID 25: Other

CRITICAL RULES:
1. ALWAYS return valid JSON only - no markdown, no code blocks, no explanations
2. Type must be EXACTLY "deposit" or "withdrawal"
3. Amount must be a positive integer (extract the main number from text)
4. category_id must match one of the IDs above, or null if unclear
5. Confidence should be 0.0 to 1.0 based on how clear the input is
6. performed_at should be ISO 8601 format if date/time mentioned, otherwise null
7. note should be in the same language as the input text
8. If it is summary of receipt, try to extract name of the place and 1–2 key items or purpose to make a short, meaningful note

RESPONSE FORMAT (return ONLY this JSON):
{"type":"deposit","amount":5000,"category_id":1,"note":"Coffee","confidence":0.95,"performed_at":null}

EXAMPLES:
Input: "Coffee 5000" → {"type":"deposit","amount":5000,"category_id":1,"note":"Coffee","confidence":0.95,"performed_at":null}
Input: "Got salary 5000000" → {"type":"deposit","amount":5000000,"category_id":8,"note":"Salary","confidence":0.95,"performed_at":null}
Input: "Taxi yesterday 15000" → {"type":"withdrawal","amount":15000,"category_id":2,"note":"Taxi","confidence":0.85,"performed_at":"2024-11-15T12:00:00Z"}

IMPORTANT: Return ONLY the JSON object, nothing else.
`
