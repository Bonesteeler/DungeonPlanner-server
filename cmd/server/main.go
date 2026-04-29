package main

import (
	"os"

	"DungeonPlannerServer/internal/handler"
	"DungeonPlannerServer/internal/service"
	"DungeonPlannerServer/internal/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	repository, err := repository.NewSceneRepository()
	if err != nil {
		e.Logger.Fatal("Failed to initialize repository: ", err)
	}
	sceneService := service.NewSceneService(repository)
	sceneHandler := handler.NewSceneHandler(sceneService)

	handler.SetupRoutes(e, sceneHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}
