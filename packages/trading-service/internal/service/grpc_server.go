package service

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

// GRPCServer wraps TradingService to provide a gRPC-compatible interface.
// Once proto compilation is set up, this struct will implement the generated interface.
type GRPCServer struct {
	tradingService *TradingService
}

// NewGRPCServer creates a new GRPCServer.
func NewGRPCServer(ts *TradingService) *GRPCServer {
	return &GRPCServer{
		tradingService: ts,
	}
}

// SubmitOrderRequest represents a gRPC request to submit an order.
type SubmitOrderRequest struct {
	UserID        string `json:"user_id"`
	Pair          string `json:"pair"`
	Side          string `json:"side"`
	OrderType     string `json:"order_type"`
	Price         string `json:"price"`
	Quantity      string `json:"quantity"`
	ClientOrderID string `json:"client_order_id"`
}

// SubmitOrderResponse represents a gRPC response for order submission.
type SubmitOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// SubmitOrder handles a gRPC SubmitOrder request.
func (s *GRPCServer) SubmitOrder(ctx context.Context, req *SubmitOrderRequest) (*SubmitOrderResponse, error) {
	price, _ := decimal.NewFromString(req.Price)
	quantity, _ := decimal.NewFromString(req.Quantity)

	order := &model.Order{
		UserID:        req.UserID,
		Pair:          req.Pair,
		Side:          model.Side(req.Side),
		OrderType:     model.OrderType(req.OrderType),
		Price:         price,
		Quantity:      quantity,
		ClientOrderID: req.ClientOrderID,
	}

	result, err := s.tradingService.SubmitOrder(ctx, order)
	if err != nil {
		return &SubmitOrderResponse{
			OrderID: result.ID,
			Status:  string(result.Status),
		}, err
	}

	return &SubmitOrderResponse{
		OrderID: result.ID,
		Status:  string(result.Status),
	}, nil
}

// CancelOrderRequest represents a gRPC cancel order request.
type CancelOrderRequest struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

// CancelOrder handles a gRPC CancelOrder request.
func (s *GRPCServer) CancelOrder(ctx context.Context, req *CancelOrderRequest) error {
	return s.tradingService.CancelOrder(ctx, req.OrderID, req.UserID)
}

// GetPositionsRequest represents a gRPC get positions request.
type GetPositionsRequest struct {
	UserID string `json:"user_id"`
}

// GetPositions handles a gRPC GetPositions request.
func (s *GRPCServer) GetPositions(ctx context.Context, req *GetPositionsRequest) []*model.TradingPosition {
	return s.tradingService.GetPositions(ctx, req.UserID)
}

// GetBalanceRequest represents a gRPC get balance request.
type GetBalanceRequest struct {
	UserID string `json:"user_id"`
}

// GetBalance handles a gRPC GetBalance request.
func (s *GRPCServer) GetBalance(ctx context.Context, req *GetBalanceRequest) *model.Balance {
	return s.tradingService.GetBalance(ctx, req.UserID)
}

// GetOpenOrdersRequest represents a gRPC get open orders request.
type GetOpenOrdersRequest struct {
	UserID string `json:"user_id"`
}

// GetOpenOrders handles a gRPC GetOpenOrders request.
func (s *GRPCServer) GetOpenOrders(ctx context.Context, req *GetOpenOrdersRequest) []*model.Order {
	return s.tradingService.GetOpenOrders(ctx, req.UserID)
}
