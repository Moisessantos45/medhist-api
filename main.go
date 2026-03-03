package main

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/pkg/middleware"
	"api_citas/internal/routes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("Hello World")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	err = db.Connect()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	err = db.InitializeDatabase()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	err = config.InitRedis(context.Background())
	if err != nil {
		fmt.Println("Error initializing Redis:", err)
		return
	}

	HOST_URL_DEV := os.Getenv("HOST_URL_DEV")
	HOST_URL_PROD := os.Getenv("HOST_URL_PROD")
	HOST_URL_PROD_WWW := os.Getenv("HOST_URL_PROD_WWW")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.RedirectTrailingSlash = false

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{HOST_URL_DEV, HOST_URL_PROD, HOST_URL_PROD_WWW},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.RateLimiterMiddleware(rate.Every(time.Second), 30))

	api := r.Group("/api/v1")
	{
		routes.AppointmentsRoutes(api)
		routes.PatientsRoutes(api)
		routes.VaccinationsRoutes(api)
		routes.VeterinariansRoutes(api)
		routes.MedicalRecordsRoutes(api)
	}

	auth := r.Group("/api/v1/auth")
	{
		routes.AuthRoutes(auth)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the API",
		})
	})

	var wg sync.WaitGroup
	wg.Go(func() {
		middleware.StartCleanup()
	})

	log.Println("Server starting on :4100...")
	srv := &http.Server{
		Addr:    ":4101",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	wg.Wait()
	log.Println("Server exiting")
}
