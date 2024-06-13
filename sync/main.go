package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AnkiConnect API URL
const ankiConnectURL = "http://localhost:8765"

func main() {
	// Replace "YourDeckName" with the actual deck name
	deckName := "var"

	// Prepare the request payload
	requestData := map[string]interface{}{
		"action":  "findCards",
		"version": 6,
		"params": map[string]interface{}{
			"query": fmt.Sprintf("deck:%s", deckName),
		},
	}

	// Convert the payload to JSON
	payload, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Send the HTTP POST request to AnkiConnect
	resp, err := http.Post(ankiConnectURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return
	}

	// Check if the request was successful
	if errMsg, ok := responseBody["error"]; ok {
		fmt.Println("AnkiConnect error:", errMsg)
		return
	}

	// Extract card IDs from the response
	cardIDs, ok := responseBody["result"].([]interface{})
	if !ok {
		fmt.Println("Error extracting card IDs from response")
		return
	}

	// Print the retrieved card IDs
	fmt.Println("Card IDs for deck", deckName, ":", cardIDs)
}
