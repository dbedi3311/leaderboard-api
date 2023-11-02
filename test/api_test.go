package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const (
	baseURL       = "http://localhost:8080"
	usernameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Generate a random string of given length
func randomString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = usernameChars[rand.Intn(len(usernameChars))]
	}
	return string(result)
}

// TestSubmitScoreHandler performs A/B and fuzzy testing on /submit-score endpoint
func TestSubmitScoreHandler(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// A/B Testing
	for i := 0; i < 100; i++ {
		username := randomString(10)
		score := float64(rand.Intn(1000))

		data := map[string]interface{}{
			"username": username,
			"score":    score,
		}
		payload, _ := json.Marshal(data)

		resp, err := http.Post(baseURL+"/submit-score", "application/json", bytes.NewBuffer(payload))
		if err != nil || resp.StatusCode != http.StatusCreated {
			t.Errorf("Failed to submit score: %v", err)
		}
		resp.Body.Close()
	}

	// Fuzzy Testing
	invalidPayloads := []string{
		// `{"username": "", "score": 100}`,
		// `{"username": "testUser"}`,
		// `{"score": 100}`,
		// `{}`, --> these should all work.
		`invalid json`,
	}

	for _, payload := range invalidPayloads {
		resp, err := http.Post(baseURL+"/submit-score", "application/json", bytes.NewBufferString(payload))
		if err != nil || resp.StatusCode == http.StatusCreated {
			t.Errorf("Expected failure for payload: %v", payload)
		}
		resp.Body.Close()
	}
}

// TestGetLeaderboardHandler performs testing on /leaderboard endpoint
func TestGetLeaderboardHandler(t *testing.T) {
	resp, err := http.Get(baseURL + "/leaderboard")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to get leaderboard: %v", err)
	}
	resp.Body.Close()
}

// TestGetRankHandler performs testing on /rank/{username} endpoint
func TestGetRankHandler(t *testing.T) {
	// Test with a valid username
	username := "testUser"
	resp, err := http.Get(fmt.Sprintf("%s/rank/%s", baseURL, username))
	if err != nil {
		t.Errorf("Failed to get rank: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected status code: %v", resp.StatusCode)
	}

	resp.Body.Close()

	// Fuzzy Testing with invalid usernames
	invalidUsernames := []string{
		"",                  // empty username
		randomString(1000),  // very long username
		"non_existent_user", // user not in the leaderboard
	}

	for _, invalidUsername := range invalidUsernames {
		resp, err := http.Get(fmt.Sprintf("%s/rank/%s", baseURL, invalidUsername))
		if err != nil {
			t.Errorf("Failed to get rank for invalid username: %v", err)
		}
		resp.Body.Close()
	}
}

func main() {
	t := &testing.T{}
	TestSubmitScoreHandler(t)
	TestGetLeaderboardHandler(t)
	TestGetRankHandler(t)
}
