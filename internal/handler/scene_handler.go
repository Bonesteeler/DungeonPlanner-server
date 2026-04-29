package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/handler/dto"
	"DungeonPlannerServer/internal/model"
)

type SceneService interface {
	ListScenes(offset int) []model.Scene
	GetSceneByID(id uuid.UUID) *model.Scene
	AddScene(scene model.Scene)
	GetSceneStats() int
}

type SceneHandler struct {
	service SceneService
}

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
	scenes := h.service.ListScenes(0)
	responses := make([]dto.SceneResponse, 0, len(scenes))
	for _, s := range scenes {
		responses = append(responses, sceneToResponse(s))
	}
	return c.JSON(http.StatusOK, responses)
}

func (h *SceneHandler) ListScenes(c echo.Context) error {
	scenes := h.service.ListScenes(0)
	responses := make([]dto.SceneResponse, 0, len(scenes))
	for _, s := range scenes {
		responses = append(responses, sceneToResponse(s))
	}
	return c.JSON(http.StatusOK, responses)
}

func (h *SceneHandler) GetSceneByID(c echo.Context, id string) error {
	sceneID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid UUID format"})
	}
	scene := h.service.GetSceneByID(sceneID)
	if scene == nil {
		return c.JSON(http.StatusOK, nil)
	}
	r := sceneToResponse(*scene)
	return c.JSON(http.StatusOK, &r)
}

func (h *SceneHandler) AddScene(c echo.Context, s dto.AddSceneRequest) error {
	h.service.AddScene(requestToScene(s))
	return c.JSON(http.StatusOK, struct{ Status string }{Status: "Scene added successfully"})
}