package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsufficientBalance(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Attempt to place an order exceeding available balance
	order := OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    50000.00,
		Quantity: 10000.0, // Extremely large quantity
	}
	resp := doPost(t, "/orders", order, token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "should reject order with insufficient balance")

	var errResp ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.NotEmpty(t, errResp.Message, "error response should have a message")
}

func TestInvalidOrderParams(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Submit malformed order - negative quantity
	malformedOrder := map[string]interface{}{
		"pair":     "BTC-USDT",
		"side":     "buy",
		"type":     "limit",
		"price":    -100.0,
		"quantity": -1.0,
	}
	resp := doPost(t, "/orders", malformedOrder, token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "should reject invalid order parameters")

	var errResp ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.NotEmpty(t, errResp.Message, "should contain error details")
}

func TestRateLimiting(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Send many requests quickly to trigger rate limiting
	var rateLimited bool
	for i := 0; i < 200; i++ {
		resp := doGet(t, "/markets/BTC-USDT/orderbook", token)
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			resp.Body.Close()
			break
		}
		resp.Body.Close()
		time.Sleep(5 * time.Millisecond)
	}

	assert.True(t, rateLimited, "should eventually receive 429 Too Many Requests")
}

func TestUnauthenticatedAccess(t *testing.T) {
	// Access protected endpoint without token
	req, err := http.NewRequest(http.MethodGet, baseURL+"/account/balances", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "should return 401 for unauthenticated access")
}
