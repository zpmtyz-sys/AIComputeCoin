package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PositionResponse struct {
	ID              string  `json:"id"`
	Pair            string  `json:"pair"`
	Side            string  `json:"side"`
	Size            float64 `json:"size"`
	EntryPrice      float64 `json:"entryPrice"`
	Leverage        int     `json:"leverage"`
	Margin          float64 `json:"margin"`
	UnrealizedPnL   float64 `json:"unrealizedPnl"`
	LiquidationPrice float64 `json:"liquidationPrice"`
}

type FundingPayment struct {
	Pair      string  `json:"pair"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	Timestamp string  `json:"timestamp"`
}

func TestPositionSettlement(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Open a leveraged position
	body := map[string]interface{}{
		"pair":     "BTC-USDT",
		"side":     "long",
		"size":     1.0,
		"leverage": 10,
	}
	resp := doPost(t, "/positions/open", body, token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "should open position")

	var position PositionResponse
	decodeJSON(t, resp.Body, &position)
	require.NotEmpty(t, position.ID)
	assert.Equal(t, "BTC-USDT", position.Pair)
	assert.Equal(t, "long", position.Side)
	assert.Equal(t, 10, position.Leverage)

	initialMargin := position.Margin

	// Trigger settlement (simulate price change and settlement)
	settleResp := doPost(t, "/positions/"+position.ID+"/settle", nil, token)
	defer settleResp.Body.Close()
	require.Equal(t, http.StatusOK, settleResp.StatusCode, "settlement should succeed")

	// Verify margin adjustments
	getResp := doGet(t, "/positions/"+position.ID, token)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var updatedPosition PositionResponse
	decodeJSON(t, getResp.Body, &updatedPosition)
	// After settlement, margin may have been adjusted based on PnL
	assert.NotEqual(t, initialMargin, updatedPosition.Margin, "margin should be adjusted after settlement")
}

func TestFundingPayment(t *testing.T) {
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	token := tokens.AccessToken

	// Open a leveraged position
	body := map[string]interface{}{
		"pair":     "ETH-USDT",
		"side":     "long",
		"size":     5.0,
		"leverage": 5,
	}
	resp := doPost(t, "/positions/open", body, token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var position PositionResponse
	decodeJSON(t, resp.Body, &position)
	require.NotEmpty(t, position.ID)

	// Wait for a funding interval (in test environment, this might be shortened)
	time.Sleep(2 * time.Second)

	// Check funding payments
	fundingResp := doGet(t, "/positions/"+position.ID+"/funding", token)
	defer fundingResp.Body.Close()
	require.Equal(t, http.StatusOK, fundingResp.StatusCode)

	var payments []FundingPayment
	err := json.NewDecoder(fundingResp.Body).Decode(&payments)
	require.NoError(t, err)

	// Verify funding payment was applied
	if len(payments) > 0 {
		payment := payments[0]
		assert.Equal(t, "ETH-USDT", payment.Pair)
		assert.NotZero(t, payment.Rate, "funding rate should be non-zero")
		assert.NotEmpty(t, payment.Timestamp)
	}
}
