package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/handler/dto"
)

func SetupRoutes(e *echo.Echo, sceneHandler *SceneHandler) {
    scenes := e.Group("v1//scenes")

    scenes.GET("/", sceneHandler.GetScenes)
    scenes.GET("/list/:start", sceneHandler.ListScenes)
    scenes.GET("/:id", func(c echo.Context) error {
        id := c.Param("id")
        return sceneHandler.GetSceneByID(c, id)
    })
    scenes.POST("/add", func(c echo.Context) error {
        var s dto.AddSceneRequest
        if err := c.Bind(&s); err != nil {
            return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid request"})
        }
        return sceneHandler.AddScene(c, s)
    })
}