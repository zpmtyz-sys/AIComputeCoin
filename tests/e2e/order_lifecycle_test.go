package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderSubmission(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Place a limit order
	order := placeOrder(t, token, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    49000.00,
		Quantity: 0.5,
	})

	require.NotEmpty(t, order.ID, "order should have an ID")
	assert.Equal(t, "BTC-USDT", order.Pair)
	assert.Equal(t, "buy", order.Side)
	assert.Equal(t, "limit", order.Type)
	assert.Equal(t, 49000.00, order.Price)
	assert.Equal(t, 0.5, order.Quantity)
	assert.NotEmpty(t, order.CreatedAt)
}

func TestOrderCancellation(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Place an order
	order := placeOrder(t, token, OrderRequest{
		Pair:     "ETH-USDT",
		Side:     "sell",
		Type:     "limit",
		Price:    3500.00,
		Quantity: 2.0,
	})
	require.NotEmpty(t, order.ID)

	// Cancel the order
	resp := cancelOrder(t, token, order.ID)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "cancel should return 200")

	// Verify order is cancelled
	getResp := doGet(t, "/orders/"+order.ID, token)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var cancelled OrderResponse
	decodeJSON(t, getResp.Body, &cancelled)
	assert.Equal(t, "cancelled", cancelled.Status)
}

func TestOrderFill(t *testing.T) {
	// Register two users for crossing orders
	email1 := uniqueEmail()
	tokens1 := registerUser(t, email1, "SecurePass123!")
	token1 := tokens1.AccessToken

	email2 := uniqueEmail()
	tokens2 := registerUser(t, email2, "SecurePass123!")
	token2 := tokens2.AccessToken

	// User 1 places a limit buy
	buyOrder := placeOrder(t, token1, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    50000.00,
		Quantity: 1.0,
	})
	require.NotEmpty(t, buyOrder.ID)

	// User 2 places a limit sell at same price (crossing order)
	sellOrder := placeOrder(t, token2, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "sell",
		Type:     "limit",
		Price:    50000.00,
		Quantity: 1.0,
	})
	require.NotEmpty(t, sellOrder.ID)

	// Verify buy order is filled
	buyResp := doGet(t, "/orders/"+buyOrder.ID, token1)
	defer buyResp.Body.Close()
	var filledBuy OrderResponse
	decodeJSON(t, buyResp.Body, &filledBuy)
	assert.Equal(t, "filled", filledBuy.Status, "buy order should be filled")

	// Verify sell order is filled
	sellResp := doGet(t, "/orders/"+sellOrder.ID, token2)
	defer sellResp.Body.Close()
	var filledSell OrderResponse
	decodeJSON(t, sellResp.Body, &filledSell)
	assert.Equal(t, "filled", filledSell.Status, "sell order should be filled")
}
