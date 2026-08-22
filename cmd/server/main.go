package main

import (
	"os"

	"DungeonPlannerServer/internal/auth"
	authHandlerPck "DungeonPlannerServer/internal/auth/handler" 
	authServicePck "DungeonPlannerServer/internal/auth/service"
	authRepoPck "DungeonPlannerServer/internal/auth/repository"
	sceneHandlerPck "DungeonPlannerServer/internal/scenes/handler"
	sceneRepoPck "DungeonPlannerServer/internal/scenes/repository"
	sceneServicePck "DungeonPlannerServer/internal/scenes/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Auth service
	tokenManager, err := auth.NewTokenManagerFromSecrets()
	if err != nil {
		e.Logger.Fatal("Failed to load auth secrets: ", err)
	}
	tokenRepo := authRepoPck.NewInMemoryTokenStore()
	userRepo, err := authRepoPck.NewUserRepository(e)
	if err != nil {
		e.Logger.Fatal("Failed to initialize user repository: ", err)
	}
	authService := authServicePck.NewAuthService(tokenRepo, tokenManager, userRepo)
	authHandler := authHandlerPck.NewAuthHandler(authService)
	authHandlerPck.SetupRoutes(e, authHandler, tokenManager)

	// Scenes service
	repository, err := sceneRepoPck.NewSceneRepository(e)
	if err != nil {
		e.Logger.Fatal("Failed to initialize repository: ", err)
	}
	sceneService := sceneServicePck.NewSceneService(repository)
	sceneHandler := sceneHandlerPck.NewSceneHandler(sceneService)

	sceneHandlerPck.SetupRoutes(e, sceneHandler, tokenManager)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}
