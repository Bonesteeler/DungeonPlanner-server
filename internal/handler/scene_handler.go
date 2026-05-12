package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/handler/dto"
	"DungeonPlannerServer/internal/model"
)

type SceneService interface {
	ListScenes(offset int) ([]model.Scene, int)
	GetSceneByID(id uuid.UUID) *model.Scene
	AddScene(scene model.Scene) error
	GetSceneStats() int
}

type SceneHandler struct {
	service SceneService
}

const pageSize = 20

func NewSceneHandler(service SceneService) *SceneHandler {
	return &SceneHandler{service: service}
}

func sceneToResponse(s model.Scene) dto.SceneResponse {
	layers := make([]dto.LayerResponse, 0, len(s.Layers))
	for _, l := range s.Layers {
		tiles := make([]dto.TileResponse, 0, len(l.Tiles))
		for _, t := range l.Tiles {
			tiles = append(tiles, dto.TileResponse{
				TileID:   t.TileID,
				Rotation: t.Rotation,
				XPos:     t.XPos,
				YPos:     t.YPos,
			})
		}
		layers = append(layers, dto.LayerResponse{Height: l.Height, Tiles: tiles})
	}
	return dto.SceneResponse{
		ID:            s.ID,
		Name:          s.Name,
		Author:        s.Author,
		UniqueTileIDs: s.UniqueTileIDs,
		Layers:        layers,
	}
}

func requestToScene(r dto.AddSceneRequest) model.Scene {
	layers := make([]model.Layer, 0, len(r.Layers))
	for _, l := range r.Layers {
		tiles := make([]model.Tile, 0, len(l.Tiles))
		for _, t := range l.Tiles {
			tiles = append(tiles, model.Tile{
				TileID:   t.TileID,
				Rotation: t.Rotation,
				XPos:     t.XPos,
				YPos:     t.YPos,
			})
		}
		layers = append(layers, model.Layer{Height: l.Height, Tiles: tiles})
	}
	return model.Scene{
		Name:   r.Name,
		Author: r.Author,
		Layers: layers,
	}
}

func (h *SceneHandler) GetScenes(c echo.Context) error {
	stats := h.service.GetSceneStats()
	return c.JSON(http.StatusOK, stats)
}

func (h *SceneHandler) ListScenes(c echo.Context, offset int) error {
	scenes, total := h.service.ListScenes(offset)
	responses := make([]dto.SceneResponse, 0, len(scenes))
	for _, s := range scenes {
		responses = append(responses, sceneToResponse(s))
	}
	return c.JSON(http.StatusOK, dto.SceneListResponse{
		Scenes:     responses,
		TotalCount: total,
		PageSize:   pageSize,
	})
}

func (h *SceneHandler) GetSceneByID(c echo.Context, id string) error {
	sceneID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid UUID format"})
	}
	scene := h.service.GetSceneByID(sceneID)
	if scene == nil {
		return c.JSON(http.StatusNotFound, nil)
	}
	r := sceneToResponse(*scene)
	return c.JSON(http.StatusOK, &r)
}

func (h *SceneHandler) AddScene(c echo.Context, s dto.AddSceneRequest) error {
	err := h.service.AddScene(requestToScene(s))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Failed to add scene"})
	}
	return c.JSON(http.StatusOK, struct{ Status string }{Status: "Scene added successfully"})
}