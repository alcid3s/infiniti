package models

type Song struct {
	ID       uint   `gorm:"primaryKey"`
	Title    string `gorm:"unique"`
	FileType string
	Artist   string
	Path     string
}
