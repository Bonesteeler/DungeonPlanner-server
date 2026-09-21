package queries

import (
	"encoding/json"

	"fmt"
	"net/http"

	"dungeonplannerserver/dto"
)

func GetAllScenes() []dto.SceneResponse {
	var scenes []dto.SceneResponse
	offset := 0

	for {
		response, err := http.Get(fmt.Sprintf("%s/v1/scenes/list/%d", APIBaseURL, offset))
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