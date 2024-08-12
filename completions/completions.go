package completions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/maxquinn/draftbot/instructions"
)

type TradeDeal struct {
	Time             string `json:"time"`
	Status           string `json:"status"`
	TeamOffering     string `json:"teamOffering"`
	TeamReceiving    string `json:"teamReceiving"`
	PlayersOffered   string `json:"playersOffered"`
	PlayersRequested string `json:"playersRequested"`
}

type OpenAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Body struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

func CreateCompletion(tradeDeal TradeDeal) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	tradeDealData, err := json.Marshal(tradeDeal)
	if err != nil {
		return "", err
	}

	body := Body{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{
				Role:    "system",
				Content: instructions.AssistantInstructions,
			},
			{
				Role:    "user",
				Content: string(tradeDealData),
			},
		},
	}
	postBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(postBody))
	resp.Header.Add("Content-Type", "application/json")
	resp.Header.Add("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	var openAPIResponse OpenAPIResponse
	json.NewDecoder(response.Body).Decode(&openAPIResponse)

	return openAPIResponse.Choices[0].Message.Content, nil
}
