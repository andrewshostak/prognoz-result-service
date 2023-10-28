package repository

type Alias struct {
	Id     uint   `gorm:"column:id;primaryKey"`
	TeamId uint   `gorm:"column:team_id"`
	Alias  string `gorm:"column:alias;unique"`
}
