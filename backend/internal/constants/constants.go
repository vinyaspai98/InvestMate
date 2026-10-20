package constants

// Gmail Sync Constants
const (
	// CDSL India email configuration
	CDSLEmailSubject    = "Transactions In Your Demat Account"
	CDSLCASEmailSubject = "CDSL Consolidated Account Statement (CAS)"

	// Sync limits
	MaxEmailsToFetch = 100 // Maximum number of emails to fetch in one sync
)

// IndianTradeHolidays contains NSE/BSE trading holidays (2024-2027).
// Standard weekend days (Saturday and Sunday) are handled separately.
var IndianTradeHolidays = map[string]bool{
	// 2024 NSE/BSE Holidays
	"2024-01-22": true, // Special Holiday
	"2024-01-26": true, // Republic Day
	"2024-03-08": true, // Mahashivratri
	"2024-03-25": true, // Holi
	"2024-03-29": true, // Good Friday
	"2024-04-11": true, // Id-Ul-Fitr (Ramzan Id)
	"2024-04-17": true, // Shri Ram Navami
	"2024-05-01": true, // Maharashtra Day
	"2024-05-20": true, // General Parliamentary Elections
	"2024-06-17": true, // Bakri Id
	"2024-07-17": true, // Moharram
	"2024-08-15": true, // Independence Day
	"2024-10-02": true, // Mahatma Gandhi Jayanti
	"2024-11-01": true, // Diwali Laxmi Pujan
	"2024-11-15": true, // Gurunanak Jayanti
	"2024-11-20": true, // Maharashtra Assembly Elections
	"2024-12-25": true, // Christmas

	// 2025 NSE/BSE Holidays
	"2025-02-26": true, // Mahashivratri
	"2025-03-14": true, // Holi
	"2025-03-31": true, // Id-Ul-Fitr
	"2025-04-10": true, // Shri Mahavir Jayanti
	"2025-04-14": true, // Dr. Baba Saheb Ambedkar Jayanti
	"2025-04-18": true, // Good Friday
	"2025-05-01": true, // Maharashtra Day
	"2025-08-15": true, // Independence Day
	"2025-08-27": true, // Shri Ganesh Chaturthi
	"2025-10-02": true, // Mahatma Gandhi Jayanti / Dussehra
	"2025-10-21": true, // Diwali Laxmi Pujan
	"2025-10-22": true, // Diwali Balipratipada
	"2025-11-05": true, // Prakash Gurpurb Sri Guru Nanak Dev
	"2025-12-25": true, // Christmas

	// 2026 NSE/BSE Holidays
	"2026-01-26": true, // Republic Day
	"2026-02-16": true, // Mahashivratri
	"2026-03-03": true, // Holi
	"2026-03-20": true, // Id-Ul-Fitr
	"2026-03-31": true, // Mahavir Jayanti
	"2026-04-03": true, // Good Friday
	"2026-04-14": true, // Dr. Ambedkar Jayanti
	"2026-05-01": true, // Maharashtra Day
	"2026-05-27": true, // Bakri Id
	"2026-06-26": true, // Moharram
	"2026-08-15": true, // Independence Day
	"2026-09-14": true, // Ganesh Chaturthi
	"2026-10-02": true, // Mahatma Gandhi Jayanti
	"2026-10-20": true, // Dussehra
	"2026-11-09": true, // Diwali
	"2026-11-24": true, // Guru Nanak Jayanti
	"2026-12-25": true, // Christmas

	// 2027 NSE/BSE Holidays
	"2027-01-26": true, // Republic Day
	"2027-03-08": true, // Mahashivratri
	"2027-03-22": true, // Holi
	"2027-03-26": true, // Good Friday
	"2027-04-14": true, // Dr. Ambedkar Jayanti
	"2027-05-01": true, // Maharashtra Day
	"2027-08-15": true, // Independence Day
	"2027-10-02": true, // Mahatma Gandhi Jayanti
	"2027-10-29": true, // Diwali
	"2027-12-25": true, // Christmas
}
