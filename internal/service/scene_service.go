package service

import (
	"github.com/google/uuid"

	"DungeonPlannerServer/internal/model"
)

type SceneRepo interface {
	ListApprovedScenes(offset int) ([]model.Scene, error)
	GetSceneByID(id uuid.UUID) (*model.Scene, error)
	AddScene(request model.Scene) error
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

func (s *SceneService) ListScenes(offset int) []model.Scene {
	scenes, err := s.repo.ListApprovedScenes(offset)
	if err != nil {
		return []model.Scene{}
	}
	if scenes == nil {
		return []model.Scene{}
	}
	return scenes
}

func (s *SceneService) GetSceneByID(id uuid.UUID) *model.Scene {
	scene, err := s.repo.GetSceneByID(id)
	if err != nil || scene == nil {
		return nil
	}
	return scene
}

func (s *SceneService) AddScene(request model.Scene) {
	_ = s.repo.AddScene(request)
}