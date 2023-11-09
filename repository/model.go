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
}

type FootballApiTeam struct {
	ID     uint `gorm:"column:id;primaryKey"`
	TeamID uint `gorm:"column:team_id"`
}

type Match struct {
	ID         uint      `gorm:"column:id;primaryKey"`
	HomeTeamID uint      `gorm:"column:home_team_id"`
	AwayTeamID uint      `gorm:"column:away_team_id"`
	StartsAt   time.Time `gorm:"column:starts_at"`
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
