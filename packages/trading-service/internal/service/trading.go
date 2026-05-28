package service

import (
	"context"
	"log"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

// TradingService implements the trading gRPC service.
type TradingService struct {
	config *config.Config
}

// NewTradingService creates a new TradingService.
func NewTradingService(cfg *config.Config) *TradingService {
	return &TradingService{
		config: cfg,
	}
}

// SubmitOrder processes a new order submission.
func (s *TradingService) SubmitOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	log.Printf("SubmitOrder: %s %s %s @ %s", order.Side, order.Quantity, order.Pair, order.Price)
	// Stub: forward to matching engine
	return order, nil
}

// CancelOrder cancels an existing order.
func (s *TradingService) CancelOrder(ctx context.Context, orderID string) error {
	log.Printf("CancelOrder: %s", orderID)
	// Stub: forward to matching engine
	return nil
}

// GetPositions returns positions for a user.
func (s *TradingService) GetPositions(ctx context.Context, userID string) ([]*model.Position, error) {
	log.Printf("GetPositions: %s", userID)
	// Stub: query database
	return nil, nil
}
