package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func PostToGoogleChat(message string, webhookURL string) error {
	payload := map[string]string{
		"text": message,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error when marshal json: %v", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(webhookURL, "application/json; charset=UTF-8", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error when send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("google return error: %d", resp.StatusCode)
	}

	return nil
}
