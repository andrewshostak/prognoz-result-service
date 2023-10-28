package repository

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
