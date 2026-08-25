package queries

import (
	"fmt"
	"net/http"

	"dungeonplannerserver/dto"
)

const baseUrl = "http://localhost:8080"

func GetAllScenes() []dto.SceneResponse {
	var scenes []dto.SceneResponse
	offset := 0

	for {
		response, err := http.Get(fmt.Sprintf("%s/v1/scenes/list/%d", baseUrl, offset))
		offset += response.PageSize
		if response.TotalCount <= offset {
			break
		}
	}
	return scenes
}