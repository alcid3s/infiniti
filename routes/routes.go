package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	song_handler "infiniti.com/controller"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.MaxMultipartMemory = 8 << 20

	router.GET("/", homeScreen)
	router.GET("/songs", getSongs)
	router.GET("/songs/:param", getSpecifiedSong)
	router.GET("/search/:param", searchSong)
	router.GET("/play/:param", playSong)
	router.POST("/upload", uploadSong)
	router.GET("/remove/:param", removeSong)

	return router
}

// @tags Home v1
// @summary Home screen
// @description Welcome message and available endpoints
// @produce  plain
// @success 200 {string} string
// @router / [get]
func homeScreen(c *gin.Context) {
	song_handler.HomeScreen(c)
}

// @Tags Get
// @Summary Get all songs
// @Description Get all songs in the database
// @Produce  json
// @Router /songs [get]
func getSongs(c *gin.Context) {
	song_handler.GetSongs(c)
}

// @Tags Get
// @Summary Get specified song
// @Description Get a song by its ID
// @Produce  json
// @Param param path int true "Song ID"
// @Router /songs/{param} [get]
func getSpecifiedSong(c *gin.Context) {
	song_handler.GetSpecifiedSong(c)
}

// @Tags Search
// @Summary Search for a song
// @Description Search for a song by its title
// @Produce  json
// @Param param path string true "Song title"
// @Router /search/{param} [get]
func searchSong(c *gin.Context) {
	song_handler.SearchSong(c)
}

// @Tags Play
// @Summary Play a song
// @Description Stream and play a song
// @Produce  json
// @Param param path int true "Song ID"
// @Router /play/{param} [get]
func playSong(c *gin.Context) {
	song_handler.PlaySong(c)
}

// @Tags Upload
// @Summary Upload a song
// @Description Upload a song to the database
// @Produce  json
// @Router /upload [post]
func uploadSong(c *gin.Context) {
	song_handler.UploadSong(c)
}

// @Tags Remove
// @Summary Remove a song
// @Description Remove a song from the database
// @Produce  json
// @Param param path int true "Song ID"
// @Router /remove/{param} [get]
func removeSong(c *gin.Context) {
	song_handler.RemoveSong(c)
}
