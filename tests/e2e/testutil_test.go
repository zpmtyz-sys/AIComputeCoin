package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	baseURL string
	wsURL   string
	client  *http.Client
)

func init() {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000/api/v1"
	}
	wsURL = os.Getenv("WS_BASE_URL")
	if wsURL == "" {
		wsURL = "ws://localhost:3000/ws"
	}
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OrderRequest struct {
	Pair     string  `json:"pair"`
	Side     string  `json:"side"`
	Type     string  `json:"type"`
	Price    float64 `json:"price,omitempty"`
	Quantity float64 `json:"quantity"`
}

type OrderResponse struct {
	ID        string  `json:"id"`
	Pair      string  `json:"pair"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type BalanceResponse struct {
	Asset     string  `json:"asset"`
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func registerUser(t *testing.T, email, password string) *AuthTokens {
	t.Helper()
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := doPost(t, "/auth/register", body, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("register failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var tokens AuthTokens
	decodeJSON(t, resp.Body, &tokens)
	return &tokens
}

func loginUser(t *testing.T, email, password string) *AuthTokens {
	t.Helper()
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := doPost(t, "/auth/login", body, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var tokens AuthTokens
	decodeJSON(t, resp.Body, &tokens)
	return &tokens
}

func placeOrder(t *testing.T, token string, order OrderRequest) *OrderResponse {
	t.Helper()
	resp := doPost(t, "/orders", order, token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("place order failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var orderResp OrderResponse
	decodeJSON(t, resp.Body, &orderResp)
	return &orderResp
}

func cancelOrder(t *testing.T, token string, orderID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/orders/"+orderID, nil)
	if err != nil {
		t.Fatalf("create cancel request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("cancel order request: %v", err)
	}
	return resp
}

func getBalances(t *testing.T, token string) []BalanceResponse {
	t.Helper()
	resp := doGet(t, "/account/balances", token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("get balances failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var balances []BalanceResponse
	decodeJSON(t, resp.Body, &balances)
	return balances
}

func doGet(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		t.Fatalf("create GET request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func doPost(t *testing.T, path string, body interface{}, token string) *http.Response {
	t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func decodeJSON(t *testing.T, r io.Reader, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func uniqueEmail() string {
	return fmt.Sprintf("testuser_%d@example.com", time.Now().UnixNano())
}
