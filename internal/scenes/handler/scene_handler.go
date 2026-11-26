package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/auth/middleware"
	"dungeonplannerserver/internal/scenes/model"
)

type SceneService interface {
	ListScenes(offset int) ([]model.Scene, int)
	ListPendingScenes(offset int) ([]model.Scene, int, error)
	GetSceneByID(id uuid.UUID) *model.Scene
	AddScene(scene model.Scene) error
	UpdateSceneModerationStatus(id uuid.UUID, status dto.ModerationStatus) (bool, error)
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
		ID:     r.ID,
		Name:   r.Name,
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

func (h *SceneHandler) ListPendingScenes(c echo.Context, offset int) error {
	scenes, total, err := h.service.ListPendingScenes(offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Failed to list pending scenes"})
	}
	responses := make([]dto.SceneResponse, 0, len(scenes))
	for _, scene := range scenes {
		responses = append(responses, sceneToResponse(scene))
	}
	return c.JSON(http.StatusOK, dto.SceneListResponse{
		Scenes:     responses,
		TotalCount: total,
		PageSize:   pageSize,
	})
}

func (h *SceneHandler) UpdateSceneModerationStatus(c echo.Context, id string, request dto.UpdateSceneModerationRequest) error {
	sceneID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid UUID format"})
	}
	if request.Status != dto.ModerationStatusApproved && request.Status != dto.ModerationStatusRejected {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid moderation status"})
	}
	updated, err := h.service.UpdateSceneModerationStatus(sceneID, request.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Failed to update scene moderation status"})
	}
	if !updated {
		return c.JSON(http.StatusConflict, struct{ Error string }{Error: "Scene not found or no longer pending review"})
	}
	return c.JSON(http.StatusOK, struct{ Status string }{Status: "Scene moderation status updated successfully"})
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
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, struct{ Error string }{Error: "Missing authentication claims"})
	}
	authorID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid user ID in token"})
	}
	sceneID, err := uuid.Parse(s.ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct{ Error string }{Error: "Invalid UUID format"})
	}
	if existing := h.service.GetSceneByID(sceneID); existing != nil {
		return c.JSON(http.StatusConflict, struct{ Error string }{Error: "Scene with this ID already exists"})
	}
	scene := requestToScene(s)
	scene.AuthorID = authorID
	if err := h.service.AddScene(scene); err != nil {
		return c.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Failed to add scene"})
	}
	return c.JSON(http.StatusOK, struct{ Status string }{Status: "Scene added successfully"})
}
