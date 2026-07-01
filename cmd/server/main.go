package main

import (
	"os"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/handler"
	"DungeonPlannerServer/internal/repository"
	"DungeonPlannerServer/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	repository, err := repository.NewSceneRepository(e)
	if err != nil {
		e.Logger.Fatal("Failed to initialize repository: ", err)
	}
	sceneService := service.NewSceneService(repository)
	sceneHandler := handler.NewSceneHandler(sceneService)

	handler.SetupRoutes(e, sceneHandler)

	if _, err := auth.NewTokenManagerFromSecrets(); err != nil {
		e.Logger.Fatal("Failed to load auth secrets: ", err)
	}

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}
