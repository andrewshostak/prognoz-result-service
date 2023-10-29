package repository

import "time"

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
