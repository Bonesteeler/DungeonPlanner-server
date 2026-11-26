package service

import (
	"github.com/google/uuid"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/scenes/model"
)

type SceneRepo interface {
	ListApprovedScenes(offset int) ([]model.Scene, error)
	ListPendingScenes(offset int) ([]model.Scene, error)
	GetApprovedSceneCount() (int, error)
	GetPendingSceneCount() (int, error)
	GetSceneByID(id uuid.UUID) (*model.Scene, error)
	AddScene(request model.Scene) error
	UpdateSceneModerationStatus(id uuid.UUID, status dto.ModerationStatus) (bool, error)
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

func (s *SceneService) ListScenes(offset int) ([]model.Scene, int) {
	total, err := s.repo.GetApprovedSceneCount()
	if err != nil {
		total = 0
	}
	scenes, err := s.repo.ListApprovedScenes(offset)
	if err != nil {
		return []model.Scene{}, total
	}
	if scenes == nil {
		return []model.Scene{}, total
	}
	return scenes, total
}

func (s *SceneService) ListPendingScenes(offset int) ([]model.Scene, int, error) {
	total, err := s.repo.GetPendingSceneCount()
	if err != nil {
		return nil, 0, err
	}
	scenes, err := s.repo.ListPendingScenes(offset)
	if err != nil {
		return nil, 0, err
	}
	if scenes == nil {
		return []model.Scene{}, total, nil
	}
	return scenes, total, nil
}

func (s *SceneService) GetSceneByID(id uuid.UUID) *model.Scene {
	scene, err := s.repo.GetSceneByID(id)
	if err != nil || scene == nil {
		return nil
	}
	return scene
}

func (s *SceneService) AddScene(request model.Scene) error {
	return s.repo.AddScene(request)
}

func (s *SceneService) UpdateSceneModerationStatus(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
	return s.repo.UpdateSceneModerationStatus(id, status)
}
