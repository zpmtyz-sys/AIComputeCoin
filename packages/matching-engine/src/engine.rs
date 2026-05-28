use rust_decimal::Decimal;
use uuid::Uuid;

use crate::error::EngineError;
use crate::orderbook::OrderBook;
use crate::types::{Order, OrderStatus, OrderType, Side, Trade};

/// Maximum number of trades retained in history to prevent unbounded memory growth.
const MAX_TRADE_HISTORY: usize = 100_000;

/// Matching engine implementing price-time priority order matching.
pub struct MatchingEngine {
    pub orderbook: OrderBook,
    pub trades: Vec<Trade>,
    pub pending_stops: Vec<Order>,
    /// Monotonically increasing sequence number, incremented on every process_order and cancel_order call.
    pub sequence_id: u64,
}

impl MatchingEngine {
    pub fn new(pair: String) -> Self {
        Self {
            orderbook: OrderBook::new(pair),
            trades: Vec::new(),
            pending_stops: Vec::new(),
            sequence_id: 0,
        }
    }

    /// Returns the most recent `limit` trades.
    pub fn get_trades(&self, limit: usize) -> Vec<Trade> {
        let len = self.trades.len();
        if limit >= len {
            self.trades.clone()
        } else {
            self.trades[len - limit..].to_vec()
        }
    }

    /// Process an incoming order against the order book.
    /// Returns a list of trades that resulted from matching.
    pub fn process_order(&mut self, mut order: Order) -> Result<Vec<Trade>, EngineError> {
        if order.quantity <= Decimal::ZERO {
            return Err(EngineError::InvalidOrder(
                "quantity must be positive".to_string(),
            ));
        }

        self.sequence_id += 1;

        let trades = match order.order_type {
            OrderType::Market => self.match_order(&mut order),
            OrderType::Limit => {
                let trades = self.match_order(&mut order);
                // Place remaining quantity on the book
                let remaining = order.quantity - order.filled_quantity;
                if remaining > Decimal::ZERO {
                    self.orderbook.add_order(order);
                }
                trades
            }
            OrderType::IOC => {
                // Cancel any unfilled portion (do not place on book)
                self.match_order(&mut order)
            }
            OrderType::FOK => {
                // Fill or Kill: only execute if full quantity can be filled
                if self.can_fill_fully(&order) {
                    self.match_order(&mut order)
                } else {
                    return Ok(vec![]);
                }
            }
            OrderType::StopLimit => {
                // Store stop-limit orders as pending until stop price is triggered
                self.pending_stops.push(order);
                return Ok(vec![]);
            }
        };

        // Store trades in history
        self.trades.extend(trades.clone());

        // Check if any stop orders should be activated
        if !trades.is_empty() {
            let stop_trades = self.check_stop_orders();
            self.trades.extend(stop_trades.clone());
            let mut all_trades = trades;
            all_trades.extend(stop_trades);
            self.cap_trade_history();
            return Ok(all_trades);
        }

        self.cap_trade_history();
        Ok(trades)
    }

    /// Cancel an order by ID, searching both the orderbook and pending stops.
    pub fn cancel_order(&mut self, order_id: Uuid) -> Result<Order, EngineError> {
        self.sequence_id += 1;

        // Search bids side
        if let Some(order) = self.orderbook.cancel_order(order_id, Side::Bid) {
            return Ok(order);
        }

        // Search asks side
        if let Some(order) = self.orderbook.cancel_order(order_id, Side::Ask) {
            return Ok(order);
        }

        // Search pending stops
        if let Some(pos) = self.pending_stops.iter().position(|o| o.id == order_id) {
            let order = self.pending_stops.remove(pos);
            return Ok(order);
        }

        Err(EngineError::OrderNotFound(order_id.to_string()))
    }

    /// Cancel an order by ID, but only if it belongs to the given user_id.
    /// Returns PermissionDenied error if the order belongs to a different user.
    pub fn cancel_order_by_user(
        &mut self,
        order_id: Uuid,
        user_id: &str,
    ) -> Result<Order, EngineError> {
        // First, find the order without removing it to check ownership
        let owner = self.find_order_owner(order_id);
        match owner {
            Some(found_user_id) => {
                if found_user_id != user_id {
                    return Err(EngineError::PermissionDenied(
                        "user_id does not match order owner".to_string(),
                    ));
                }
                self.cancel_order(order_id)
            }
            None => Err(EngineError::OrderNotFound(order_id.to_string())),
        }
    }

    /// Find the user_id of an order by its ID without removing it.
    fn find_order_owner(&self, order_id: Uuid) -> Option<String> {
        // Search bids
        for queue in self.orderbook.bids.values() {
            for order in queue.iter() {
                if order.id == order_id {
                    return Some(order.user_id.clone());
                }
            }
        }

        // Search asks
        for queue in self.orderbook.asks.values() {
            for order in queue.iter() {
                if order.id == order_id {
                    return Some(order.user_id.clone());
                }
            }
        }

        // Search pending stops
        for order in &self.pending_stops {
            if order.id == order_id {
                return Some(order.user_id.clone());
            }
        }

        None
    }

    /// Check pending stop orders against the last trade price.
    /// Activates stop orders whose stop_price has been crossed.
    pub fn check_stop_orders(&mut self) -> Vec<Trade> {
        let last_trade_price = match self.trades.last() {
            Some(trade) => trade.price,
            None => return vec![],
        };

        let mut activated = Vec::new();
        let mut remaining = Vec::new();

        for order in self.pending_stops.drain(..) {
            let stop_price = order.stop_price.unwrap_or(order.price);
            let should_activate = match order.side {
                Side::Bid => stop_price <= last_trade_price,
                Side::Ask => stop_price >= last_trade_price,
            };

            if should_activate {
                activated.push(order);
            } else {
                remaining.push(order);
            }
        }

        self.pending_stops = remaining;

        let mut all_trades = Vec::new();
        for mut order in activated {
            // Process as a limit order
            order.order_type = OrderType::Limit;
            let trades = self.match_order(&mut order);
            let rem = order.quantity - order.filled_quantity;
            if rem > Decimal::ZERO {
                self.orderbook.add_order(order);
            }
            all_trades.extend(trades);
        }

        all_trades
    }

    /// Cap trade history to MAX_TRADE_HISTORY entries to prevent unbounded memory growth.
    fn cap_trade_history(&mut self) {
        if self.trades.len() > MAX_TRADE_HISTORY {
            let excess = self.trades.len() - MAX_TRADE_HISTORY;
            self.trades.drain(..excess);
        }
    }

    fn match_order(&mut self, order: &mut Order) -> Vec<Trade> {
        let mut trades = Vec::new();
        let timestamp = order.timestamp;

        loop {
            let remaining = order.quantity - order.filled_quantity;
            if remaining <= Decimal::ZERO {
                break;
            }

            let opposite_book = match order.side {
                Side::Bid => &mut self.orderbook.asks,
                Side::Ask => &mut self.orderbook.bids,
            };

            let best_price = match order.side {
                Side::Bid => opposite_book.keys().next().copied(),
                Side::Ask => opposite_book.keys().next_back().copied(),
            };

            let best_price = match best_price {
                Some(p) => p,
                None => break,
            };

            // Check price compatibility
            let price_compatible = match order.side {
                Side::Bid => order.price >= best_price || order.order_type == OrderType::Market,
                Side::Ask => order.price <= best_price || order.order_type == OrderType::Market,
            };

            if !price_compatible {
                break;
            }

            let opposite_book = match order.side {
                Side::Bid => &mut self.orderbook.asks,
                Side::Ask => &mut self.orderbook.bids,
            };

            if let Some(queue) = opposite_book.get_mut(&best_price) {
                while !queue.is_empty() && (order.quantity - order.filled_quantity) > Decimal::ZERO
                {
                    let maker = queue.front().unwrap();

                    // Self-trade prevention: if same user, remove maker and skip
                    // TODO: Production should emit cancel events for removed maker orders
                    // so subscribers and the maker are notified of the cancellation.
                    if order.user_id == maker.user_id {
                        queue.pop_front();
                        continue;
                    }

                    let maker = queue.front_mut().unwrap();
                    let maker_remaining = maker.quantity - maker.filled_quantity;
                    let taker_remaining = order.quantity - order.filled_quantity;
                    let fill_qty = maker_remaining.min(taker_remaining);

                    maker.filled_quantity += fill_qty;
                    order.filled_quantity += fill_qty;

                    let trade = Trade {
                        id: Uuid::new_v4(),
                        maker_order_id: maker.id,
                        taker_order_id: order.id,
                        pair: order.pair.clone(),
                        price: best_price,
                        quantity: fill_qty,
                        side: order.side,
                        timestamp,
                    };
                    trades.push(trade);

                    if maker.filled_quantity >= maker.quantity {
                        queue.pop_front();
                    }
                }

                if queue.is_empty() {
                    let opposite_book = match order.side {
                        Side::Bid => &mut self.orderbook.asks,
                        Side::Ask => &mut self.orderbook.bids,
                    };
                    opposite_book.remove(&best_price);
                }
            }
        }

        trades
    }

    fn can_fill_fully(&self, order: &Order) -> bool {
        let mut available = Decimal::ZERO;

        match order.side {
            Side::Bid => {
                for (price, queue) in self.orderbook.asks.iter() {
                    if *price > order.price {
                        break;
                    }
                    for o in queue.iter() {
                        available += o.quantity - o.filled_quantity;
                        if available >= order.quantity {
                            return true;
                        }
                    }
                }
            }
            Side::Ask => {
                for (price, queue) in self.orderbook.bids.iter().rev() {
                    if *price < order.price {
                        break;
                    }
                    for o in queue.iter() {
                        available += o.quantity - o.filled_quantity;
                        if available >= order.quantity {
                            return true;
                        }
                    }
                }
            }
        }

        available >= order.quantity
    }
}

impl Order {
    pub fn status(&self) -> OrderStatus {
        if self.filled_quantity == Decimal::ZERO {
            OrderStatus::New
        } else if self.filled_quantity >= self.quantity {
            OrderStatus::Filled
        } else {
            OrderStatus::PartiallyFilled
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal_macros::dec;

    fn make_order(
        user_id: &str,
        side: Side,
        order_type: OrderType,
        price: Decimal,
        quantity: Decimal,
    ) -> Order {
        Order {
            id: Uuid::new_v4(),
            user_id: user_id.to_string(),
            pair: "BTC/USDT".to_string(),
            side,
            order_type,
            price,
            quantity,
            filled_quantity: Decimal::ZERO,
            stop_price: None,
            timestamp: 1000,
        }
    }

    #[test]
    fn test_basic_limit_match() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
        let buy = make_order("user_b", Side::Bid, OrderType::Limit, dec!(100), dec!(5));

        engine.process_order(sell).unwrap();
        let trades = engine.process_order(buy).unwrap();

        assert_eq!(trades.len(), 1);
        assert_eq!(trades[0].price, dec!(100));
        assert_eq!(trades[0].quantity, dec!(5));
    }

    #[test]
    fn test_self_trade_prevention() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
        let buy = make_order("user_a", Side::Bid, OrderType::Limit, dec!(100), dec!(5));

        engine.process_order(sell).unwrap();
        let trades = engine.process_order(buy).unwrap();

        // No trade should be generated (same user)
        assert!(trades.is_empty());
    }

    #[test]
    fn test_partial_fill() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(3));
        let buy = make_order("user_b", Side::Bid, OrderType::Limit, dec!(100), dec!(10));

        engine.process_order(sell).unwrap();
        let trades = engine.process_order(buy).unwrap();

        assert_eq!(trades.len(), 1);
        assert_eq!(trades[0].quantity, dec!(3));
        // Remaining 7 should be on the book
        assert_eq!(engine.orderbook.order_count(), 1);
    }

    #[test]
    fn test_market_order_empty_book() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let buy = make_order("user_a", Side::Bid, OrderType::Market, dec!(0), dec!(5));
        let trades = engine.process_order(buy).unwrap();

        assert!(trades.is_empty());
    }

    #[test]
    fn test_ioc_unfilled() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        // Place a small sell order
        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(2));
        engine.process_order(sell).unwrap();

        // IOC buy for 10, only 2 available
        let buy = make_order("user_b", Side::Bid, OrderType::IOC, dec!(100), dec!(10));
        let trades = engine.process_order(buy).unwrap();

        assert_eq!(trades.len(), 1);
        assert_eq!(trades[0].quantity, dec!(2));
        // Remainder is cancelled - not placed on book
        assert_eq!(engine.orderbook.order_count(), 0);
    }

    #[test]
    fn test_fok_rejected() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        // Place a small sell order
        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(2));
        engine.process_order(sell).unwrap();

        // FOK buy for 10, only 2 available - should be rejected
        let buy = make_order("user_b", Side::Bid, OrderType::FOK, dec!(100), dec!(10));
        let trades = engine.process_order(buy).unwrap();

        assert!(trades.is_empty());
        // The existing sell should still be on the book
        assert_eq!(engine.orderbook.order_count(), 1);
    }

    #[test]
    fn test_fok_filled() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        // Place enough sell liquidity
        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(10));
        engine.process_order(sell).unwrap();

        // FOK buy for 5 with enough liquidity
        let buy = make_order("user_b", Side::Bid, OrderType::FOK, dec!(100), dec!(5));
        let trades = engine.process_order(buy).unwrap();

        assert_eq!(trades.len(), 1);
        assert_eq!(trades[0].quantity, dec!(5));
    }

    #[test]
    fn test_stop_limit_activation() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        // Place a sell order on the book
        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
        engine.process_order(sell).unwrap();

        // Place a stop-limit buy that activates when price >= 100
        let mut stop_buy = make_order(
            "user_c",
            Side::Bid,
            OrderType::StopLimit,
            dec!(105),
            dec!(3),
        );
        stop_buy.stop_price = Some(dec!(100));
        engine.process_order(stop_buy).unwrap();

        // The stop order should be pending
        assert_eq!(engine.pending_stops.len(), 1);

        // Trigger a trade at price 100 by normal matching
        let sell2 = make_order("user_a", Side::Ask, OrderType::Limit, dec!(99), dec!(2));
        engine.process_order(sell2).unwrap();

        let buy = make_order("user_b", Side::Bid, OrderType::Limit, dec!(99), dec!(2));
        let _trades = engine.process_order(buy).unwrap();

        // The trade at 99 should have activated the stop buy (stop_price 100 <= last_trade 99 is false)
        // Actually stop_price 100 <= 99 is false, so let's adjust the test
        // For a buy stop: activates when stop_price <= last_trade_price
        // So we need last trade price >= 100

        // Let's use a fresh engine for clarity
        let mut engine2 = MatchingEngine::new("BTC/USDT".to_string());

        // Place sell liquidity at 105
        let sell_at_105 = make_order("user_a", Side::Ask, OrderType::Limit, dec!(105), dec!(10));
        engine2.process_order(sell_at_105).unwrap();

        // Place stop-limit buy: activates when price reaches 100, then buys at 105
        let mut stop_buy = make_order(
            "user_c",
            Side::Bid,
            OrderType::StopLimit,
            dec!(105),
            dec!(3),
        );
        stop_buy.stop_price = Some(dec!(100));
        engine2.process_order(stop_buy).unwrap();
        assert_eq!(engine2.pending_stops.len(), 1);

        // Place a sell and a buy to generate a trade at price 100
        let sell_at_100 = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(2));
        engine2.process_order(sell_at_100).unwrap();

        let buy_at_100 = make_order("user_b", Side::Bid, OrderType::Limit, dec!(100), dec!(2));
        let trades = engine2.process_order(buy_at_100).unwrap();

        // Trade at 100 generated, stop order with stop_price=100 should activate (100 <= 100)
        assert!(!trades.is_empty());
        // The stop order should have been activated
        assert_eq!(engine2.pending_stops.len(), 0);
    }

    #[test]
    fn test_cancel_order() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let sell = make_order("user_a", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
        let order_id = sell.id;
        engine.process_order(sell).unwrap();

        let cancelled = engine.cancel_order(order_id).unwrap();
        assert_eq!(cancelled.id, order_id);
        assert_eq!(engine.orderbook.order_count(), 0);
    }

    #[test]
    fn test_cancel_nonexistent() {
        let mut engine = MatchingEngine::new("BTC/USDT".to_string());

        let result = engine.cancel_order(Uuid::new_v4());
        assert!(result.is_err());
    }
}
