package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"dungeonplannerserver/dto"
)

func Login(username, password string) (*dto.GeneratedTokensResponse, error) {
	requestBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	response, err := http.Post(fmt.Sprintf("%s/auth/login", APIBaseURL), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %s", response.Status)
	}

	var result dto.GeneratedTokensResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
