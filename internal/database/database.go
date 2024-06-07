package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dhowden/tag"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"infiniti.com/pkg/models"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorBlue  = "\033[34m"
)

func checkErrorAndExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Connect(username string, password string, database string) (*gorm.DB, error) {
	dsn := username + ":" + password + "@tcp(db:3306)/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	fmt.Println(colorRed + "DSN: " + dsn + colorReset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.Song{})
	if err != nil {
		log.Fatal(err)
	}
}

func Seed(db *gorm.DB, path string) {
	// getting all the songs in the songs directory
	fmt.Println(colorRed + "Initializing songs" + colorReset)
	err := os.Chdir(path)
	checkErrorAndExit(err)

	// reading files in current directory
	files, err := os.ReadDir(".")
	checkErrorAndExit(err)

	// iterating over files
	for _, file := range files {
		fileInfo, err := os.Stat(file.Name())
		checkErrorAndExit(err)

		// if file is not a directory
		if !fileInfo.IsDir() {
			file, err := os.Open(file.Name())
			if err != nil {
				log.Fatal(err)
			}
			ParseFileToSongDatatype(db, file)
		}
	}
}

func ParseFileToSongDatatype(db *gorm.DB, file *os.File) {
	// reading metadata from the file
	m, err := tag.ReadFrom(file)
	if err == nil {
		if (db.Where("title = ?", strings.Split(file.Name(), ".")[0]).First(&models.Song{}).Error != nil) {
			// save position of extension
			extensionsPosition := len(strings.Split(file.Name(), ".")) - 1

			// Make sure a . is allowed in the title as long as the extension isn't taken into account.
			title := strings.Join(strings.Split(file.Name(), ".")[:extensionsPosition], ".")

			// adding song to database
			AddSong(db, models.Song{
				Title:    title,
				FileType: strings.Split(file.Name(), ".")[extensionsPosition],
				Artist:   m.Artist(),
				Path:     file.Name(),
			})
		}
	}
}

func GetSongById(db *gorm.DB, id int) (*models.Song, error) {
	var song models.Song
	err := db.First(&song, id).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &song, nil
}

func GetSongByTitle(db *gorm.DB, searchTerm string) ([]models.Song, error) {
	searchTerm = strings.ToLower(strings.ReplaceAll(searchTerm, " ", ""))

	fmt.Println("Searching for: ", searchTerm)

	var songs []models.Song
	err := db.Where("LOWER(REPLACE(title, ' ', '')) LIKE ?", "%"+searchTerm+"%").Find(&songs).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return songs, nil
}

func GetSongs(db *gorm.DB) []models.Song {
	var songs []models.Song

	err := db.Find(&songs).Error
	if err != nil {
		log.Println(err)
	}

	return songs
}

func RemoveSong(db *gorm.DB, song models.Song) error {
	err := os.Remove(song.Path)
	if err == nil {
		return err
	}

	err = db.Delete(&models.Song{}, song.ID).Error
	if err != nil {
		return err
	}

	return nil
}

func AddSong(db *gorm.DB, song models.Song) error {
	err := db.Create(&song).Error
	if err != nil {
		return err
	}

	return nil
}
