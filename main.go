package main

import (
	"github.com/gin-gonic/gin"
	"photoUploader/controllers"
)

func main() {
	router := gin.Default()

	router.GET("/health", controllers.Health)

	router.GET("/photo", controllers.GetAllPhotos)
	router.GET("/photo/:photoName", controllers.GetPhotoByName)
	router.POST("/photo", controllers.UploadPhoto)
	router.DELETE("/photo/:photoName", controllers.DeletePhoto)

	// Run the server on port 8080 and log errors if any
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
