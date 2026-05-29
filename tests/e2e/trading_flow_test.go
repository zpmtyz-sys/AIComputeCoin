package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullTradingFlow(t *testing.T) {
	// Step 1: Register a new user
	email := uniqueEmail()
	password := "SecurePass123!"
	tokens := registerUser(t, email, password)
	require.NotEmpty(t, tokens.AccessToken, "should receive access token after registration")

	// Step 2: Login with the registered user
	loginTokens := loginUser(t, email, password)
	require.NotEmpty(t, loginTokens.AccessToken, "should receive access token after login")
	token := loginTokens.AccessToken

	// Step 3: Get initial balances
	balances := getBalances(t, token)
	assert.NotEmpty(t, balances, "should have initial balances")

	// Step 4: Place a limit buy order
	buyOrder := placeOrder(t, token, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    50000.00,
		Quantity: 0.1,
	})
	require.NotEmpty(t, buyOrder.ID, "buy order should have an ID")
	assert.Equal(t, "buy", buyOrder.Side)
	assert.Equal(t, "limit", buyOrder.Type)

	// Step 5: Place a market sell order that should cross with the buy
	sellOrder := placeOrder(t, token, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "sell",
		Type:     "market",
		Quantity: 0.1,
	})
	require.NotEmpty(t, sellOrder.ID, "sell order should have an ID")
	assert.Equal(t, "sell", sellOrder.Side)

	// Step 6: Verify order statuses (both should be filled)
	resp := doGet(t, "/orders/"+buyOrder.ID, token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var filledBuy OrderResponse
	decodeJSON(t, resp.Body, &filledBuy)
	assert.Equal(t, "filled", filledBuy.Status, "buy order should be filled")

	// Step 7: Check updated balances
	updatedBalances := getBalances(t, token)
	assert.NotEmpty(t, updatedBalances, "should have updated balances after trade")
}
