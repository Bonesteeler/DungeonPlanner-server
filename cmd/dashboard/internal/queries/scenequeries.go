package queries

import (
	"encoding/json"

	"fmt"
	"net/http"

	"dungeonplannerserver/dto"
	"dungeonplannerserver-dashboard/internal"
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
		if err != nil {
			fmt.Println("Error fetching scenes:", err)
			break
		}
		defer response.Body.Close()

		scenes = append(scenes, result.Scenes...)
		offset += result.PageSize
		if result.TotalCount <= offset {
			break
		}
	}
	return scenes
}