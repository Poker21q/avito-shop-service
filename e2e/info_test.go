package e2e

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
)

const (
	infoURL = baseURL + "info"
)

type AuthResponse struct {
	Token string `json:"token"`
}

type UserInfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory struct {
		Received []CoinTransfer `json:"received"`
		Sent     []CoinTransfer `json:"sent"`
	} `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinTransfer struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Amount   int    `json:"amount"`
}

func getAuthToken(username, password string) (string, error) {
	resp, err := sendAuthRequest(username, password)
	if err != nil {
		return "", fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	return authResp.Token, nil
}

func createAuthHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func sendInfoRequest(headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	return client.Do(req)
}

func testInfo(t *testing.T, token string, expectedStatus int, expectedUserInfo UserInfoResponse, stepDescription string) {
	log.Println(stepDescription)

	headers := createAuthHeader(token)
	resp, err := sendInfoRequest(headers)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if expectedStatus == http.StatusOK {
		var userInfo UserInfoResponse
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		checkCoins(t, userInfo.Coins, expectedUserInfo.Coins)
		checkInventory(t, userInfo.Inventory, expectedUserInfo.Inventory)
		checkCoinHistory(t, userInfo.CoinHistory, expectedUserInfo.CoinHistory)
	}
}

func checkCoins(t *testing.T, actualCoins, expectedCoins int) {
	if actualCoins != expectedCoins {
		t.Errorf("expected coin balance %d, got %d", expectedCoins, actualCoins)
	}
}

func checkInventory(t *testing.T, actualInventory, expectedInventory []InventoryItem) {
	expectedMap := make(map[string]int)
	for _, item := range expectedInventory {
		expectedMap[item.Type] = item.Quantity
	}

	for _, item := range actualInventory {
		expectedQty, exists := expectedMap[item.Type]
		if !exists {
			t.Errorf("unexpected inventory item: %+v", item)
			continue
		}
		if item.Quantity != expectedQty {
			t.Errorf("inventory mismatch for item %s: expected quantity %d, got %d", item.Type, expectedQty, item.Quantity)
		}
		// Убираем элемент из мапы, чтобы проверить, что все элементы были проверены
		delete(expectedMap, item.Type)
	}

	for remainingItemType, remainingQty := range expectedMap {
		t.Errorf("expected inventory item %s with quantity %d, but not found", remainingItemType, remainingQty)
	}
}

// Проверка истории монет
func checkCoinHistory(t *testing.T, actualHistory struct {
	Received []CoinTransfer `json:"received"`
	Sent     []CoinTransfer `json:"sent"`
}, expectedHistory struct {
	Received []CoinTransfer `json:"received"`
	Sent     []CoinTransfer `json:"sent"`
}) {
	// Проверка полученных монет
	if len(actualHistory.Received) != len(expectedHistory.Received) {
		t.Errorf("expected received coin history length %d, got %d", len(expectedHistory.Received), len(actualHistory.Received))
	} else {
		for i := range actualHistory.Received {
			if actualHistory.Received[i].FromUser != expectedHistory.Received[i].FromUser ||
				actualHistory.Received[i].Amount != expectedHistory.Received[i].Amount {
				t.Errorf("received coin history mismatch at index %d: expected %+v, got %+v", i, expectedHistory.Received[i], actualHistory.Received[i])
			}
		}
	}

	// Проверка отправленных монет
	if len(actualHistory.Sent) != len(expectedHistory.Sent) {
		t.Errorf("expected sent coin history length %d, got %d", len(expectedHistory.Sent), len(actualHistory.Sent))
	} else {
		for i := range actualHistory.Sent {
			if actualHistory.Sent[i].ToUser != expectedHistory.Sent[i].ToUser ||
				actualHistory.Sent[i].Amount != expectedHistory.Sent[i].Amount {
				t.Errorf("sent coin history mismatch at index %d: expected %+v, got %+v", i, expectedHistory.Sent[i], actualHistory.Sent[i])
			}
		}
	}
}

func TestGetUserInfo(t *testing.T) {
	log.Println("Starting test: GET /api/info")
	token, err := getAuthToken("infotest", "infotest")
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}
	expectedUserInfo := UserInfoResponse{Coins: 1000}
	testInfo(t, token, http.StatusOK, expectedUserInfo, "Fetching user info with valid token")
	testInfo(t, "invalid_token", http.StatusUnauthorized, UserInfoResponse{}, "Fetching user info with invalid token")
}
