package routes

import (
	"github.com/gin-gonic/gin"
	song_handler "infiniti.com/pkg/handles"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
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

func homeScreen(c *gin.Context) {
	song_handler.HomeScreen(c)
}

func getSongs(c *gin.Context) {
	song_handler.GetSongs(c)
}

func getSpecifiedSong(c *gin.Context) {
	song_handler.GetSpecifiedSong(c)
}

func searchSong(c *gin.Context) {
	song_handler.SearchSong(c)
}

func playSong(c *gin.Context) {
	song_handler.PlaySong(c)
}

func uploadSong(c *gin.Context) {
	song_handler.UploadSong(c)
}

func removeSong(c *gin.Context) {
	song_handler.RemoveSong(c)
}
