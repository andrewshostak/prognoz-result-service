package client

type FixturesResponse struct {
	Response []Result `json:"response"`
}

type Result struct {
	Fixture Fixture `json:"fixture"`
	Teams   Teams   `json:"teams"`
	Goals   Goals   `json:"goals"`
	Score   Score   `json:"score"`
}

type Fixture struct {
	ID     uint   `json:"id"`
	Status Status `json:"status"`
}

type Teams struct {
	Home Team `json:"home"`
	Away Team `json:"away"`
}

type Team struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Goals struct {
	Home uint `json:"home"`
	Away uint `json:"away"`
}

type Score struct {
	Halftime  Goals `json:"halftime"`
	Fulltime  Goals `json:"fulltime"`
	Extratime Goals `json:"extratime"`
}

type Status struct {
	Short string `json:"short"`
}

type FixtureSearch struct {
	Season   uint
	Timezone string
	Date     *string
	TeamID   *uint
}
