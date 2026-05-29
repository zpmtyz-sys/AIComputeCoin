package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WSMessage struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

func connectWS(t *testing.T) *websocket.Conn {
	t.Helper()
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err, "should connect to WebSocket")
	return conn
}

func TestWebSocketConnection(t *testing.T) {
	conn := connectWS(t)
	defer conn.Close()

	// Read the initial connection message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err, "should receive initial message")

	var wsMsg WSMessage
	err = json.Unmarshal(msg, &wsMsg)
	require.NoError(t, err, "should parse connection message")
	assert.Equal(t, "connected", wsMsg.Type, "should receive connected message type")
}

func TestOrderbookSubscription(t *testing.T) {
	// Connect to WebSocket
	conn := connectWS(t)
	defer conn.Close()

	// Read initial connection message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err := conn.ReadMessage()
	require.NoError(t, err)

	// Subscribe to orderbook channel
	subMsg := map[string]string{
		"type":    "subscribe",
		"channel": "orderbook",
		"pair":    "BTC-USDT",
	}
	err = conn.WriteJSON(subMsg)
	require.NoError(t, err, "should send subscribe message")

	// Read subscription confirmation
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err, "should receive subscription confirmation")

	var wsMsg WSMessage
	err = json.Unmarshal(msg, &wsMsg)
	require.NoError(t, err)
	assert.Equal(t, "subscribed", wsMsg.Type)
	assert.Equal(t, "orderbook", wsMsg.Channel)

	// Place an order via REST to trigger an update
	email := uniqueEmail()
	tokens := registerUser(t, email, "SecurePass123!")
	placeOrder(t, tokens.AccessToken, OrderRequest{
		Pair:     "BTC-USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    48000.00,
		Quantity: 0.5,
	})

	// Wait for WebSocket update
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, updateMsg, err := conn.ReadMessage()
	require.NoError(t, err, "should receive orderbook update via WebSocket")

	var update WSMessage
	err = json.Unmarshal(updateMsg, &update)
	require.NoError(t, err)
	assert.Equal(t, "orderbook", update.Channel, "update should be on orderbook channel")
}

func TestKlineSubscription(t *testing.T) {
	conn := connectWS(t)
	defer conn.Close()

	// Read initial connection message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err := conn.ReadMessage()
	require.NoError(t, err)

	// Subscribe to kline channel
	subMsg := map[string]string{
		"type":     "subscribe",
		"channel":  "kline",
		"pair":     "BTC-USDT",
		"interval": "1m",
	}
	err = conn.WriteJSON(subMsg)
	require.NoError(t, err, "should send kline subscribe message")

	// Read subscription confirmation
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err, "should receive kline subscription confirmation")

	var wsMsg WSMessage
	err = json.Unmarshal(msg, &wsMsg)
	require.NoError(t, err)
	assert.Equal(t, "subscribed", wsMsg.Type)
	assert.Equal(t, "kline", wsMsg.Channel)

	// Wait for a kline data update
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, dataMsg, err := conn.ReadMessage()
	require.NoError(t, err, "should receive kline data stream")

	var klineUpdate WSMessage
	err = json.Unmarshal(dataMsg, &klineUpdate)
	require.NoError(t, err)
	assert.Equal(t, "kline", klineUpdate.Channel)
}
