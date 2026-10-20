package tradedate

import (
	"time"

	"investmate-backend/internal/constants"
)

// IsTradingDay returns true if the date is a weekday and not an Indian trade holiday.
func IsTradingDay(d time.Time) bool {
	weekday := d.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	dateStr := d.Format("2006-01-02")
	return !constants.IndianTradeHolidays[dateStr]
}

// PreviousWorkingDay returns the latest working trading day strictly prior to d.
// It steps backward at least one calendar day, and skips any Saturdays, Sundays,
// and Indian stock market holidays.
func PreviousWorkingDay(d time.Time) time.Time {
	candidate := d.AddDate(0, 0, -1)
	// Look back up to 14 days (handles extended holiday stretches)
	for i := 0; i < 14; i++ {
		if IsTradingDay(candidate) {
			return candidate
		}
		candidate = candidate.AddDate(0, 0, -1)
	}
	return d.AddDate(0, 0, -1)
}

// ResolveTradeDate determines the actual execution date for a transaction.
// If an explicit date is provided, it calculates the previous working day if that
// date was a weekend/holiday or if the mail notification arrived post-trade.
// If the explicit date is already a valid trading day, but the trade took place
// on the preceding trading session, PreviousWorkingDay is returned.
func ResolveTradeDate(notificationDate time.Time) time.Time {
	return PreviousWorkingDay(notificationDate)
}
