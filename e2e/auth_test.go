package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"
)

const authURL = baseURL + "auth"

func createRequestBody(username, password string) (io.Reader, error) {
	body, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(body), nil
}

func sendAuthRequest(username, password string) (*http.Response, error) {
	body, err := createRequestBody(username, password)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(authURL, "application/json", body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func testAuth(t *testing.T, username, password string, expectedStatus int, stepDescription string) {
	log.Println(stepDescription)

	resp, err := sendAuthRequest(username, password)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func TestAuthFlow(t *testing.T) {
	log.Println("Starting test: Auth flow...")
	testAuth(t, "ivan", "password", http.StatusOK, "Attempting authentication with valid credentials...")
	testAuth(t, "ivan", "invalidpassword", http.StatusUnauthorized, "Attempting authentication with invalid password...")
	testAuth(t, "ivan", "password", http.StatusOK, "Attempting authentication again with valid credentials...")
	log.Println("Test completed successfully.")
}
