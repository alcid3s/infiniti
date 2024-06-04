package song_handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tcolgate/mp3"
	"gorm.io/gorm"
	"infiniti.com/pkg/audiopipeline"
	"infiniti.com/pkg/database"
	"infiniti.com/pkg/models"
)

var db *gorm.DB

func Init(database *gorm.DB) {
	db = database
}

func GetSpecifiedSong(c *gin.Context) {
	song, err := getSong(c)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
	} else {
		c.IndentedJSON(http.StatusOK, song)
	}
}

func HomeScreen(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Infiniti! \n\nAvailable endpoints: \n\nGET /songs \nGET /songs/:param \nGET /search/:param \n"+
		"GET /play/:param \nGET /remove/:param \nPOST /upload (example: curl -X POST http://127.0.0.1:9000/upload -F \"file=@/Users/guest/Music/Johannes Brahms - Hungarian Dance No.5.mp3\")"+
		"\n\nEnjoy!")
}

func PlaySong(c *gin.Context) {
	song, err := getSong(c)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		return
	}

	connPool := audiopipeline.NewConnectionPool()

	file, err := openFile(song)
	if err != nil {
		log.Println(err)
	}

	bytes, err := readSongContents(file)
	if err != nil {
		log.Println(err)
	}

	// create a go routine to stream the song
	go audiopipeline.Stream(connPool, bytes, float32(calculatePlayLength("../songs/"+song.Path)))

	audiopipeline.PlayAudiofile(connPool, song.FileType, c)
}

func GetSongs(c *gin.Context) {
	var songs = database.GetSongs(db)
	if songs != nil {
		c.IndentedJSON(http.StatusOK, songs)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Such empty"})
	}
}

func RemoveSong(c *gin.Context) {
	song, err := getSong(c)
	fmt.Println("song:", song, "err:", err)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
	} else {
		c.IndentedJSON(http.StatusOK, song)
		err = database.RemoveSong(db, song)
		if err != nil {
			log.Println(err)
		}
		database.RemoveSong(db, song)

	}
}

func UploadSong(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Println(err)
	}

	err = c.SaveUploadedFile(fileHeader, "../songs/"+fileHeader.Filename)
	if err != nil {
		log.Println(err)
	}

	file, err := os.Open(fileHeader.Filename)
	if err != nil {
		log.Println(err)
	}

	database.ParseFileToSongDatatype(db, file)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", fileHeader.Filename))
}

func SearchSong(c *gin.Context) {
	searchTerm := c.Param("param")
	songs, err := database.GetSongByTitle(db, searchTerm)
	if err != nil {
		log.Println(err)
	}

	c.IndentedJSON(http.StatusOK, songs)
}

/**
	Private functions
**/

func getSong(c *gin.Context) (models.Song, error) {
	param := c.Param("param")

	id, succes := strconv.Atoi(param)
	if succes == nil {
		song, err := database.GetSongById(db, id)
		if err == nil {
			return *song, nil
		}
		return models.Song{}, err
	} else {
		songs, err := database.GetSongByTitle(db, param)
		if err == nil && len(songs) > 0 {
			return songs[0], nil
		}
		err = fmt.Errorf("song not found")
		return models.Song{}, err
	}
}

// source: https://stackoverflow.com/questions/60281655/how-to-find-the-length-of-mp3-file-in-golang
func calculatePlayLength(path string) float64 {
	t := 0.0

	r, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0

	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return 0.0
		}

		t = t + f.Duration().Seconds()
	}
	return t
}

func openFile(song models.Song) (*os.File, error) {
	fname := "../songs/" + song.Path
	file, err := os.Open(fname)
	if err != nil {
		log.Println(err)
	}

	return file, err
}

func readSongContents(file *os.File) ([]byte, error) {
	ctn, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	return ctn, err
}
