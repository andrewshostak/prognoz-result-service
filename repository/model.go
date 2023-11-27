package repository

import (
	"time"

	"github.com/jackc/pgtype"
)

type Alias struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	TeamID uint   `gorm:"column:team_id"`
	Alias  string `gorm:"column:alias;unique"`

	FootballApiTeam *FootballApiTeam `gorm:"foreignKey:team_id"`
}

type Team struct {
	ID uint `gorm:"column:id;primaryKey"`

	Aliases []Alias
}

type FootballApiTeam struct {
	ID     uint `gorm:"column:id;primaryKey"`
	TeamID uint `gorm:"column:team_id"`
}

type Match struct {
	ID           uint         `gorm:"column:id;primaryKey"`
	HomeTeamID   uint         `gorm:"column:home_team_id"`
	AwayTeamID   uint         `gorm:"column:away_team_id"`
	StartsAt     time.Time    `gorm:"column:starts_at"`
	ResultStatus ResultStatus `gorm:"column:result_status;default:not_scheduled"`

	FootballApiFixtures []FootballApiFixture
	HomeTeam            *Team `gorm:"foreignKey:home_team_id"`
	AwayTeam            *Team `gorm:"foreignKey:away_team_id"`
}

type FootballApiFixture struct {
	ID      uint         `gorm:"column:id;primaryKey"`
	MatchID uint         `gorm:"column:match_id"`
	Data    pgtype.JSONB `gorm:"column:data"`

	Match *Match `gorm:"foreignKey:match_id"`
}

type Subscription struct {
	ID         uint       `gorm:"column:id;primaryKey"`
	Url        string     `gorm:"column:url;unique"`
	MatchID    uint       `gorm:"column:match_id"`
	Key        string     `gorm:"column:key;unique"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	NotifiedAt *time.Time `gorm:"column:notified_at"`
}

type ResultStatus string

const (
	NotScheduled    ResultStatus = "not_scheduled"
	Scheduled       ResultStatus = "scheduled"
	SchedulingError ResultStatus = "scheduling_error"
	Error           ResultStatus = "error"
	Successful      ResultStatus = "successful"
)

type Data struct {
	Fixture Fixture       `json:"fixture"`
	Teams   TeamsExternal `json:"teams"`
	Goals   Goals         `json:"goals"`
}

type Fixture struct {
	ID     uint   `json:"id"`
	Status Status `json:"status"`
	Date   string `json:"date"`
}

type TeamsExternal struct {
	Home TeamExternal `json:"home"`
	Away TeamExternal `json:"away"`
}

type TeamExternal struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Goals struct {
	Home uint `json:"home"`
	Away uint `json:"away"`
}

type Status struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}
