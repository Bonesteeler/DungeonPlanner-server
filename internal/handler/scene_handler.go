package handler

import (
  "net/http"

	"github.com/labstack/echo/v4"
	"github.com/google/uuid"

	"DungeonPlannerServer/internal/handler/dto"
)

type SceneService interface {
	ListScenes(offset int) []dto.SceneResponse
	GetSceneByID(id uuid.UUID) *dto.SceneResponse
	AddScene(scene dto.AddSceneRequest)
	GetSceneStats() int
}

type SceneHandler struct {
	service SceneService
}

func NewSceneHandler(service SceneService) *SceneHandler {
	return &SceneHandler{service: service}
}

func (h *SceneHandler) GetScenes(c echo.Context) error {
	scenes := h.service.ListScenes(0)
	return c.JSON(http.StatusOK, scenes)
}

func (h *SceneHandler) ListScenes(c echo.Context) error {
	return c.JSON(http.StatusOK, h.service.ListScenes(0))
}

func (h *SceneHandler) GetSceneByID(c echo.Context, id string) error {
	sceneID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid UUID format"})
	}
	return c.JSON(http.StatusOK, h.service.GetSceneByID(sceneID))
}

func (h *SceneHandler) AddScene(c echo.Context, s dto.AddSceneRequest) error {
	h.service.AddScene(s)
	return c.JSON(http.StatusOK, struct{ Status string }{Status: "Scene added successfully"})
}