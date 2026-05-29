package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/events"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/margin"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/position"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/settlement"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/validation"
)

// ErrOrderNotFound is returned when an order cannot be found.
var ErrOrderNotFound = errors.New("order not found")

// ErrDuplicateOrder is returned when a duplicate client order ID is submitted.
var ErrDuplicateOrder = errors.New("duplicate order: client order ID already exists")

// TradingService implements the trading business logic.
type TradingService struct {
	config          *config.Config
	validator       *validation.OrderValidator
	rateLimiter     *validation.RateLimiter
	idempotency     *validation.IdempotencyStore
	positionTracker *position.PositionTracker
	settlement      *settlement.SettlementEngine
	marginSystem    *margin.MarginSystem
	eventProducer   events.EventProducer

	mu       sync.RWMutex
	orders   map[string]*model.Order   // orderID -> order
	balances map[string]*model.Balance // userID -> balance
}

// NewTradingService creates a new TradingService with all components wired together.
func NewTradingService(cfg *config.Config) *TradingService {
	return &TradingService{
		config:          cfg,
		validator:       validation.NewOrderValidator(),
		rateLimiter:     validation.NewRateLimiter(),
		idempotency:     validation.NewIdempotencyStore(),
		positionTracker: position.NewPositionTracker(),
		settlement:      settlement.NewSettlementEngine(),
		marginSystem:    margin.NewMarginSystem(),
		eventProducer:   events.NewInMemoryProducer(),
		orders:          make(map[string]*model.Order),
		balances:        make(map[string]*model.Balance),
	}
}

// NewTradingServiceWithDeps creates a TradingService with injected dependencies (for testing).
func NewTradingServiceWithDeps(
	cfg *config.Config,
	validator *validation.OrderValidator,
	rateLimiter *validation.RateLimiter,
	idempotency *validation.IdempotencyStore,
	posTracker *position.PositionTracker,
	settle *settlement.SettlementEngine,
	marginSys *margin.MarginSystem,
	producer events.EventProducer,
) *TradingService {
	return &TradingService{
		config:          cfg,
		validator:       validator,
		rateLimiter:     rateLimiter,
		idempotency:     idempotency,
		positionTracker: posTracker,
		settlement:      settle,
		marginSystem:    marginSys,
		eventProducer:   producer,
		orders:          make(map[string]*model.Order),
		balances:        make(map[string]*model.Balance),
	}
}

// SetBalance sets or creates a balance for a user (used for setup/testing).
func (s *TradingService) SetBalance(userID string, available decimal.Decimal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] = model.NewBalance(userID, available)
}

// SubmitOrder validates, checks idempotency, locks margin, and stores the order.
func (s *TradingService) SubmitOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(order.UserID, validation.TierDefault) {
		order.Status = model.OrderStatusRejected
		order.RejectionReason = "rate limit exceeded"
		_ = s.eventProducer.Publish(events.EventOrderRejected, events.OrderRejectedPayload{
			OrderID: order.ID,
			UserID:  order.UserID,
			Reason:  order.RejectionReason,
		})
		return order, errors.New("rate limit exceeded")
	}

	// Check idempotency
	if order.ClientOrderID != "" {
		if exists, existingID := s.idempotency.Check(order.ClientOrderID); exists {
			s.mu.RLock()
			existing, ok := s.orders[existingID]
			s.mu.RUnlock()
			if ok {
				return existing, ErrDuplicateOrder
			}
		}
	}

	// Get or create balance
	s.mu.Lock()
	balance, exists := s.balances[order.UserID]
	if !exists {
		balance = model.NewBalance(order.UserID, decimal.Zero)
		s.balances[order.UserID] = balance
	}
	s.mu.Unlock()

	// Validate order
	result := s.validator.ValidateOrder(order, balance)
	if !result.Valid {
		order.Status = model.OrderStatusRejected
		order.RejectionReason = result.Errors[0]
		_ = s.eventProducer.Publish(events.EventOrderRejected, events.OrderRejectedPayload{
			OrderID: order.ID,
			UserID:  order.UserID,
			Reason:  order.RejectionReason,
		})
		return order, errors.New(order.RejectionReason)
	}

	// Assign ID if not set
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Lock initial margin
	requiredMargin := s.marginSystem.CalculateInitialMargin(order)
	if err := balance.Lock(requiredMargin); err != nil {
		order.Status = model.OrderStatusRejected
		order.RejectionReason = "insufficient margin"
		_ = s.eventProducer.Publish(events.EventOrderRejected, events.OrderRejectedPayload{
			OrderID: order.ID,
			UserID:  order.UserID,
			Reason:  order.RejectionReason,
		})
		return order, err
	}

	// Store order
	order.Status = model.OrderStatusNew
	order.FilledQuantity = decimal.Zero
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	s.mu.Lock()
	s.orders[order.ID] = order
	s.mu.Unlock()

	// Track open order count
	s.validator.IncrementOpenOrders(order.UserID)

	// Store idempotency key
	if order.ClientOrderID != "" {
		s.idempotency.Store(order.ClientOrderID, order.ID)
	}

	// Publish event
	_ = s.eventProducer.Publish(events.EventOrderSubmitted, events.OrderSubmittedPayload{
		OrderID:       order.ID,
		UserID:        order.UserID,
		Pair:          order.Pair,
		Side:          string(order.Side),
		OrderType:     string(order.OrderType),
		Price:         order.Price,
		Quantity:      order.Quantity,
		ClientOrderID: order.ClientOrderID,
	})

	return order, nil
}

// CancelOrder cancels an existing order and unlocks its margin.
func (s *TradingService) CancelOrder(ctx context.Context, orderID, userID string) error {
	s.mu.Lock()
	order, exists := s.orders[orderID]
	if !exists {
		s.mu.Unlock()
		return ErrOrderNotFound
	}
	if order.UserID != userID {
		s.mu.Unlock()
		return ErrOrderNotFound
	}
	if order.Status != model.OrderStatusNew && order.Status != model.OrderStatusPartiallyFilled {
		s.mu.Unlock()
		return errors.New("order cannot be cancelled in current status")
	}

	order.Status = model.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	balance := s.balances[order.UserID]
	s.mu.Unlock()

	// Unlock margin for unfilled portion
	unfilledQty := order.Quantity.Sub(order.FilledQuantity)
	if !unfilledQty.IsZero() && balance != nil {
		marginToUnlock := order.Price.Mul(unfilledQty).Mul(decimal.NewFromFloat(0.10))
		_ = balance.Unlock(marginToUnlock)
	}

	// Decrement open orders
	s.validator.DecrementOpenOrders(order.UserID)

	// Publish event
	_ = s.eventProducer.Publish(events.EventOrderCancelled, events.OrderCancelledPayload{
		OrderID: orderID,
		UserID:  userID,
		Reason:  "user requested cancellation",
	})

	return nil
}

// GetBalance returns the balance for a user.
func (s *TradingService) GetBalance(ctx context.Context, userID string) *model.Balance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	balance, exists := s.balances[userID]
	if !exists {
		return model.NewBalance(userID, decimal.Zero)
	}
	return balance
}

// GetPositions returns open positions for a user.
func (s *TradingService) GetPositions(ctx context.Context, userID string) []*model.TradingPosition {
	return s.positionTracker.GetPositions(userID)
}

// GetOpenOrders returns all open orders for a user.
func (s *TradingService) GetOpenOrders(ctx context.Context, userID string) []*model.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Order
	for _, order := range s.orders {
		if order.UserID == userID &&
			(order.Status == model.OrderStatusNew || order.Status == model.OrderStatusPartiallyFilled) {
			result = append(result, order)
		}
	}
	return result
}

// GetOrder returns a specific order by ID.
func (s *TradingService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// GetEventProducer returns the event producer (for testing).
func (s *TradingService) GetEventProducer() events.EventProducer {
	return s.eventProducer
}

// GetPositionTracker returns the position tracker (for testing).
func (s *TradingService) GetPositionTracker() *position.PositionTracker {
	return s.positionTracker
}

// Stop shuts down background goroutines.
func (s *TradingService) Stop() {
	s.idempotency.Stop()
}
