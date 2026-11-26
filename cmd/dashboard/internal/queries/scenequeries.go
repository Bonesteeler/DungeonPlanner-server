package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"dungeonplannerserver-dashboard/internal"
	"dungeonplannerserver/dto"
)

const authorizationKey = "Authorization"
const prefix = "Bearer "

func AddAuthHeader(req *http.Request, token string) {
	req.Header.Set(authorizationKey, prefix+token)
}

func GetAllScenes() []dto.SceneResponse {
	var token = internal.GetAccessToken()

	var scenes []dto.SceneResponse
	offset := 0

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/scenes/list/%d", APIBaseURL, offset), nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			break
		}
		AddAuthHeader(req, token)

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error fetching scenes:", err)
			break
		}
		defer response.Body.Close()

		var result dto.SceneListResponse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			fmt.Println("Error decoding response:", err)
			break
		}

		scenes = append(scenes, result.Scenes...)
		offset += result.PageSize
		if result.TotalCount <= offset {
			break
		}
	}
	return scenes
}

func GetPendingScenes() ([]dto.SceneResponse, error) {
	var scenes []dto.SceneResponse
	offset := 0

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/scenes/moderation/pending/%d", APIBaseURL, offset), nil)
		if err != nil {
			return nil, fmt.Errorf("create pending scenes request: %w", err)
		}
		AddAuthHeader(req, internal.GetAccessToken())

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch pending scenes: %w", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, moderationResponseError(response, "fetch pending scenes")
		}

		var result dto.SceneListResponse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode pending scenes: %w", err)
		}
		scenes = append(scenes, result.Scenes...)
		offset += result.PageSize
		if result.TotalCount <= offset {
			break
		}
	}
	return scenes, nil
}

func UpdateSceneModerationStatus(sceneID string, status dto.ModerationStatus) error {
	body, err := json.Marshal(dto.UpdateSceneModerationRequest{Status: status})
	if err != nil {
		return fmt.Errorf("encode moderation request: %w", err)
	}
	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/v1/scenes/moderation/%s", APIBaseURL, sceneID),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create moderation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	AddAuthHeader(req, internal.GetAccessToken())

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("update scene moderation status: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return moderationResponseError(response, "update scene moderation status")
	}
	return nil
}

func moderationResponseError(response *http.Response, action string) error {
	var result dto.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err == nil && result.Error != "" {
		return fmt.Errorf("%s: %s", action, result.Error)
	}
	return fmt.Errorf("%s: %s", action, response.Status)
}
