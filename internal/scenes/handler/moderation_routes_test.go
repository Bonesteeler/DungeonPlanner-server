package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/auth"
	"dungeonplannerserver/internal/scenes/model"
)

func newModerationTestServer(service SceneService) (*echo.Echo, *auth.TokenManager) {
	e := echo.New()
	tokenManager := auth.NewTokenManager([]byte("test-access-secret"), []byte("test-refresh-secret"))
	SetupRoutes(e, NewSceneHandler(service), tokenManager)
	return e, tokenManager
}

func moderationRequest(t *testing.T, e *echo.Echo, tokenManager *auth.TokenManager, method, path, role string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	token, err := tokenManager.GenerateAccessToken(uuid.NewString(), role)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestModerationRoutes_ModeratorCanListAndReview(t *testing.T) {
	sceneID := uuid.New()
	var statuses []dto.ModerationStatus
	service := &mockSceneService{
		listPendingScenesFn: func(offset int) ([]model.Scene, int, error) {
			if offset != 0 {
				t.Errorf("offset = %d, want 0", offset)
			}
			return []model.Scene{{ID: sceneID.String(), Name: "Pending", Author: "Author"}}, 1, nil
		},
		updateSceneModerationStatusFn: func(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
			if id != sceneID {
				t.Errorf("scene ID = %v, want %v", id, sceneID)
			}
			statuses = append(statuses, status)
			return true, nil
		},
	}
	e, tokens := newModerationTestServer(service)

	rec := moderationRequest(t, e, tokens, http.MethodGet, "/v1/scenes/moderation/pending/0", "moderator", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("pending status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var list dto.SceneListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("invalid pending response: %v", err)
	}
	if len(list.Scenes) != 1 || list.Scenes[0].ID != sceneID.String() || list.TotalCount != 1 || list.PageSize != pageSize {
		t.Errorf("unexpected pending response: %+v", list)
	}

	for _, status := range []dto.ModerationStatus{dto.ModerationStatusApproved, dto.ModerationStatusRejected} {
		body := []byte(fmt.Sprintf(`{"status":%d}`, status))
		rec = moderationRequest(t, e, tokens, http.MethodPatch, "/v1/scenes/moderation/"+sceneID.String(), "moderator", body)
		if rec.Code != http.StatusOK {
			t.Fatalf("moderation status = %d, body = %s", rec.Code, rec.Body.String())
		}
	}
	if len(statuses) != 2 || statuses[0] != dto.ModerationStatusApproved || statuses[1] != dto.ModerationStatusRejected {
		t.Errorf("moderation statuses = %v", statuses)
	}
}

func TestModerationRoutes_Authorization(t *testing.T) {
	for _, test := range []struct {
		role string
		want int
	}{
		{role: "user", want: http.StatusForbidden},
		{role: "moderator", want: http.StatusOK},
		{role: "admin", want: http.StatusOK},
	} {
		t.Run(test.role, func(t *testing.T) {
			service := &mockSceneService{
				listPendingScenesFn: func(offset int) ([]model.Scene, int, error) { return []model.Scene{}, 0, nil },
			}
			e, tokens := newModerationTestServer(service)
			rec := moderationRequest(t, e, tokens, http.MethodGet, "/v1/scenes/moderation/pending/0", test.role, nil)
			if rec.Code != test.want {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, test.want, rec.Body.String())
			}
		})
	}
}

func TestModerationRoutes_UserCannotReview(t *testing.T) {
	service := &mockSceneService{
		updateSceneModerationStatusFn: func(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
			t.Error("ordinary user reached moderation service")
			return true, nil
		},
	}
	e, tokens := newModerationTestServer(service)
	body := []byte(fmt.Sprintf(`{"status":%d}`, dto.ModerationStatusApproved))
	rec := moderationRequest(t, e, tokens, http.MethodPatch, "/v1/scenes/moderation/"+uuid.NewString(), "user", body)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestModerationRoutes_InvalidJSON(t *testing.T) {
	e, tokens := newModerationTestServer(&mockSceneService{})
	rec := moderationRequest(t, e, tokens, http.MethodPatch, "/v1/scenes/moderation/"+uuid.NewString(), "moderator", []byte(`{"status":`))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestModerationRoutes_InvalidPendingOffset(t *testing.T) {
	service := &mockSceneService{
		listPendingScenesFn: func(offset int) ([]model.Scene, int, error) {
			t.Error("service should not be called for an invalid offset")
			return nil, 0, nil
		},
	}
	e, tokens := newModerationTestServer(service)
	rec := moderationRequest(t, e, tokens, http.MethodGet, "/v1/scenes/moderation/pending/not-a-number", "moderator", nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
