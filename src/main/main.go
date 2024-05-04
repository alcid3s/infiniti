package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"infiniti/audiopipeline"

	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"
	"github.com/tcolgate/mp3"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

type song struct {
	Id       int         `json:"id"`
	Title    string      `json:"title"`
	FileType string      `json:"fileType"`
	Artist   string      `json:"artist"`
	Length   float64     `json:"length"`
	Image    tag.Picture `json:"image"`
}

var songs []song

const PORT = "9000"

func checkErrorAndExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkErrorAndPass(err error) {
	if err != nil {
		log.Print(err)
	}
}

func init() {
	fmt.Println(colorRed + "Initializing songs" + colorReset)
	err := os.Chdir("../songs")
	checkErrorAndExit(err)

	files, err := os.ReadDir(".")
	checkErrorAndExit(err)

	i := 0
	for _, file := range files {
		fileInfo, err := os.Stat(file.Name())
		checkErrorAndExit(err)

		if !fileInfo.IsDir() {
			file, err := os.Open(file.Name())
			checkErrorAndExit(err)

			m, err := tag.ReadFrom(file)
			if err == nil {
				songs = append(songs, song{
					Id:       i,
					Title:    strings.Split(file.Name(), ".")[0],
					FileType: strings.Split(file.Name(), ".")[1],
					Artist:   m.Artist(),
					Length:   calculateTrackLength(file),
					Image:    *m.Picture(),
				})
			}
		}
		i++
	}
}

// source: https://stackoverflow.com/questions/60281655/how-to-find-the-length-of-mp3-file-in-golang
func calculateTrackLength(file *os.File) float64 {
	d := mp3.NewDecoder(file)
	var f mp3.Frame
	skipped := 0
	t := 0.0
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

	if t == 0.0 {
		log.Print(colorRed + "Couldn't calculate track length" + colorReset)
	}
	return t
}

func retrieveFile(name string, id int) (song, error) {
	makeComparable := func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", ""))
	}

	for _, song := range songs {
		if name != "" && makeComparable(song.Title) == makeComparable(name) {
			return song, nil
		}
		if id != -1 && song.Id == id {
			return song, nil
		}
	}

	return song{}, errors.New("song not found")
}

func getSongs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, songs)
}

func getSong(c *gin.Context) (song, error) {
	param := c.Param("param")

	var song song
	var err error
	if id, err := strconv.Atoi(param); err == nil {
		song, err = retrieveFile("", id)
		checkErrorAndPass(err)
	} else {
		song, err = retrieveFile(param, -1)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
			checkErrorAndPass(err)
		}
	}

	return song, err
}

func readSongContents(song song) ([]byte, error) {
	fname := "../songs/" + song.Title + "." + song.FileType
	file, err := os.Open(fname)
	checkErrorAndPass(err)

	ctn, err := io.ReadAll(file)
	checkErrorAndPass(err)

	return ctn, err
}

func playSong(c *gin.Context) {
	song, err := getSong(c)
	checkErrorAndExit(err)

	connPool := audiopipeline.NewConnectionPool()

	bytes, err := readSongContents(song)
	checkErrorAndExit(err)

	// create a go routine to stream the song
	go audiopipeline.Stream(connPool, bytes, float32(song.Length))

	audiopipeline.PlayAudiofile(connPool, song.FileType, c, song.Image.Data)
}

func main() {
	router := gin.Default()
	router.GET("/songs", getSongs)

	// get song via title or id
	router.GET("/songs/:param", func(c *gin.Context) {
		song, err := getSong(c)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		} else {
			c.IndentedJSON(http.StatusOK, song)
		}
	})

	// play song via title or id
	router.GET("/play/:param", playSong)

	router.Run("0.0.0.0:" + PORT)
}
