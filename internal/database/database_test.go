package database

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	model "infiniti.com/model"
)

const PATH = "../../resources/test_songs"

func createMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening gorm database", err)
	}

	return gormDB, mock, nil
}

func closeDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestMigrate(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	closeDB(db)
}

func TestSeed(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	mock.ExpectExec("INSERT INTO `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Seed(db, PATH)

	closeDB(db)
}

func TestGetSongById(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	mock.ExpectExec("INSERT INTO `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Seed(db, PATH)

	mock.ExpectExec("SELECT * FROM `songs` WHERE `songs`.`id` = 1").WillReturnResult(sqlmock.NewResult(0, 1))
	GetSongById(db, 1)

	closeDB(db)
}

func TestGetSongByTitle(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	mock.ExpectExec("INSERT INTO `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Seed(db, PATH)

	searchTerm := "summer"

	mock.ExpectExec("SELECT * FROM `songs` WHERE `songs`.`title` LIKE ?" + "%" + searchTerm + "%").WillReturnResult(sqlmock.NewResult(0, 1))
	GetSongByTitle(db, searchTerm)

	closeDB(db)
}

func TestGetSongs(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	mock.ExpectExec("INSERT INTO `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Seed(db, PATH)

	mock.ExpectQuery("SELECT * FROM `songs`").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "artist", "album", "genre", "path"}))
	GetSongs(db)

	closeDB(db)
}

func TestRemoveSong(t *testing.T) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("Failed to open database: %v", err)
	}

	mock.ExpectExec("CREATE TABLE `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Migrate(db)

	mock.ExpectExec("INSERT INTO `songs`").WillReturnResult(sqlmock.NewResult(0, 1))
	Seed(db, PATH)

	mock.ExpectExec("SELECT * FROM `songs` WHERE `songs`.`id` = 1").WillReturnResult(sqlmock.NewResult(0, 1))
	RemoveSong(db, model.Song{ID: 1, Path: PATH + "/Recording.mp3"})

	closeDB(db)
}
