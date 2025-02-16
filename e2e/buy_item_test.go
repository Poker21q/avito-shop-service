package e2e

import (
	"log"
	"net/http"
	"testing"
)

const (
	buyURL = baseURL + "buy/"
)

func sendBuyRequest(item, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", buyURL+item, nil)
	if err != nil {
		return nil, err
	}

	headers := createAuthHeader(token)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	return client.Do(req)
}

func testBuyItem(t *testing.T, token, item string, expectedStatus int, expectedCoins int, expectedInventory []InventoryItem, stepDescription string) {
	log.Println(stepDescription)

	// Отправляем запрос на покупку
	resp, err := sendBuyRequest(item, token)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if expectedStatus == http.StatusOK {
		expectedUserInfo := UserInfoResponse{
			Coins:     expectedCoins,
			Inventory: expectedInventory,
		}
		testInfo(t, token, expectedStatus, expectedUserInfo, "Verifying user info after buying "+item)
	}
}

func TestBuyItem(t *testing.T) {
	log.Println("Starting test: GET /api/buy/{item}")
	token, err := getAuthToken("buyitem", "buyitem")
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	// Test with sufficient coins
	testBuyItem(t, token, "t-shirt", http.StatusOK, 920, []InventoryItem{{Type: "t-shirt", Quantity: 1}}, "Buying t-shirt (80 coins)")
	testBuyItem(t, token, "hoody", http.StatusOK, 620, []InventoryItem{{Type: "t-shirt", Quantity: 1}, {Type: "hoody", Quantity: 1}}, "Buying hoody (300 coins)")
	testBuyItem(t, token, "powerbank", http.StatusOK, 420, []InventoryItem{{Type: "t-shirt", Quantity: 1}, {Type: "hoody", Quantity: 1}, {Type: "powerbank", Quantity: 1}}, "Buying powerbank (200 coins)")
	testBuyItem(t, token, "powerbank", http.StatusOK, 220, []InventoryItem{{Type: "t-shirt", Quantity: 1}, {Type: "hoody", Quantity: 1}, {Type: "powerbank", Quantity: 2}}, "Buying powerbank (200 coins)")

	// Test invalid item
	testBuyItem(t, token, "invalid-item", http.StatusBadRequest, 1000, nil, "Buying an invalid item")

	// Test with insufficient funds
	testBuyItem(t, token, "hoody", http.StatusBadRequest, 220, []InventoryItem{{Type: "t-shirt", Quantity: 1}, {Type: "hoody", Quantity: 1}, {Type: "powerbank", Quantity: 1}}, "Buying hoody (insufficient funds)")

	// Test unauthorized user (invalid token)
	testBuyItem(t, "invalid_token", "t-shirt", http.StatusUnauthorized, 1000, nil, "Attempting to buy t-shirt with invalid token")
}
