package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ldez/mimetype"
)

func ExampleAPIHandler_HandleAPIRequest() {
	type shortenRequest struct {
		URL string `json:"url"`
	}

	type shortenResponse struct {
		Result string `json:"result"`
	}

	request := shortenRequest{
		URL: "https://practicum.yandex.ru/",
	}

	data, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}

	responseData, err := http.Post("http://localhost:8080/api/shorten", mimetype.ApplicationJSON, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer responseData.Body.Close()

	response := shortenResponse{}
	if err := json.NewDecoder(responseData.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Status code: %v\n", responseData.StatusCode)
	fmt.Printf("Short URL: %v\n", response.Result)

	// Не используем Output, чтобы тесты не требовали запущенного сервиса.
}
