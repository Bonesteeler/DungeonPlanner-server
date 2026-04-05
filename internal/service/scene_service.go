package service

import (
    "github.com/google/uuid"

    "DungeonPlannerServer/internal/handler/dto"
)

type SceneRepo interface {
    ListApprovedScenes(offset int) ([]dto.SceneResponse, error)
    GetSceneByID(id uuid.UUID) (*dto.SceneResponse, error)
    AddScene(request dto.AddSceneRequest) error
}

type SceneService struct {
    repo SceneRepo
}

func NewSceneService(repo SceneRepo) *SceneService {
    return &SceneService{repo: repo}
}

func (s *SceneService) GetSceneStats() int {
    return 0
}

func (s *SceneService) ListScenes(offset int) []dto.SceneResponse {
    scenes, err := s.repo.ListApprovedScenes(offset)
    if err != nil {
        return []dto.SceneResponse{}
    }
    if scenes == nil {
        return []dto.SceneResponse{}
    }
    return scenes
}

func (s *SceneService) GetSceneByID(id uuid.UUID) *dto.SceneResponse {
    scene, err := s.repo.GetSceneByID(id)
    if err != nil || scene == nil {
        return nil
    }
    return scene
}

func (s *SceneService) AddScene(request dto.AddSceneRequest) {
    _ = s.repo.AddScene(request)
}