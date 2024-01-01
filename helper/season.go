package helper

import "time"

type SeasonHelper struct{}

func NewSeasonHelper() *SeasonHelper {
	return &SeasonHelper{}
}

// CurrentSeason returns current year if current time is after June 1, otherwise previous year
func (h *SeasonHelper) CurrentSeason() int {
	now := time.Now()

	seasonBound := time.Date(now.Year(), 6, 1, 0, 0, 0, 0, time.UTC)

	if now.After(seasonBound) {
		return now.Year()
	}

	return now.AddDate(-1, 0, 0).Year()
}
