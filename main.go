//YOOPY CHRISTIAN - Stockbit Golang Developer

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func readEnvVar(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {
	//gin setup
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	//route setup
	r.GET("/home", ApiLanding)
	r.GET("/movie", GetMovie)
	r.GET("/movie/:id", GetMovieByID)

	//Logging
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.SetOutput(file)
	log.Print("Logging Running")

	//http server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	//gracefully shutdown
	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func ApiLanding(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "hello"})
}

func GetMovie(c *gin.Context) {

	search_query := c.Query("searchword")
	page_query := c.Query("pagination")
	if len(search_query) == 0 || len(page_query) == 0 {
		log.Printf("Search Query or Page is not define")
		res := map[string]string{
			"status": "404",
		}
		c.JSON(http.StatusNotFound, res)
		return
	}

	api_url := readEnvVar("URL") + "?apikey=" + readEnvVar("OMDBKey") + "&s=" + search_query + "&page=" + page_query
	fmt.Println(api_url)

	response, err := http.Get(api_url)
	if err != nil || response.StatusCode != http.StatusOK {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	log.Printf("go to %s", api_url)

	reader := response.Body
	defer reader.Close()
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="test-stockbit"`,
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)

}

func GetMovieByID(c *gin.Context) {
	id_movie := c.Param("id")
	api_url := readEnvVar("URL") + "?apikey=" + readEnvVar("OMDBKEY") + "&i=" + id_movie

	response, err := http.Get(api_url)
	if err != nil || response.StatusCode != http.StatusOK {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	log.Printf("go to %s", api_url)

	reader := response.Body
	defer reader.Close()
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="test-stockbit"`,
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
