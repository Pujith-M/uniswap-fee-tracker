package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "Service is healthy")
	})

	return r
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := setupRouter()

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
