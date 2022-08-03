package main

import (
	"fmt"
	"goqrs/database"
	"goqrs/envs"
	"goqrs/handlers"
	"goqrs/security"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/ksaucedo002/kcheck"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Eror loading .env file:", err)
		os.Exit(0)
	}
	kcheck.AddFunc("uuid", func(atom kcheck.Atom, _ string) error {
		if _, err := uuid.Parse(atom.Value); err != nil {
			return fmt.Errorf("el campo %s debe ser un identificador uuid", atom.Name)
		}
		return nil
	})
}
func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if err := security.LoadRSAKeys(); err != nil {
		log.Println("Err RSA Kes:", err)
		os.Exit(0)
	}
	log.Println("RSA KEYS OK!")
	if err := database.PrepareConnection(); err != nil {
		log.Println("Err database:", err)
		os.Exit(0)
	}
	log.Println("DATABASE OK!")
	e := echo.New()
	e.HideBanner = true
	e.Use(database.GormMiddleware)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"*"},
	}))
	handlers.StartRoutes(e)
	go func() {
		for sig := range c {
			if err := e.Close(); err != nil {
				log.Println("error trying to stop echo service:", err)
				os.Exit(0)
			}
			log.Println("stopping echo service:", sig)
			os.Exit(1)
		}
	}()
	err := e.Start(envs.FindEnv("GOQRS_ADDRESS", "localhost:8080"))
	if err != nil {
		log.Println("Error staring echo service:", err)
		os.Exit(0)
	}
}
