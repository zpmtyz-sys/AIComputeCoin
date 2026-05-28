use uuid::Uuid;

use crate::error::EngineError;
use crate::orderbook::OrderBook;
use crate::types::{Order, OrderStatus, OrderType, Side, Trade};

/// Matching engine implementing price-time priority order matching.
pub struct MatchingEngine {
    pub orderbook: OrderBook,
}

impl MatchingEngine {
    pub fn new(pair: String) -> Self {
        Self {
            orderbook: OrderBook::new(pair),
        }
    }

    /// Process an incoming order against the order book.
    /// Returns a list of trades that resulted from matching.
    pub fn process_order(&mut self, mut order: Order) -> Result<Vec<Trade>, EngineError> {
        if order.quantity <= 0.0 {
            return Err(EngineError::InvalidOrder(
                "quantity must be positive".to_string(),
            ));
        }

        let trades = match order.order_type {
            OrderType::Market => self.match_order(&mut order),
            OrderType::Limit => {
                let trades = self.match_order(&mut order);
                // Place remaining quantity on the book
                let remaining = order.quantity - order.filled_quantity;
                if remaining > 0.0 {
                    self.orderbook.add_order(order);
                }
                trades
            }
            OrderType::IOC => {
                let trades = self.match_order(&mut order);
                // Cancel any unfilled portion (do not place on book)
                trades
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
                // Stop-limit orders are placed on the book as limit orders
                // once the stop price is triggered (simplified here as immediate placement)
                let trades = self.match_order(&mut order);
                let remaining = order.quantity - order.filled_quantity;
                if remaining > 0.0 {
                    self.orderbook.add_order(order);
                }
                trades
            }
        };

        Ok(trades)
    }

    fn match_order(&mut self, order: &mut Order) -> Vec<Trade> {
        let mut trades = Vec::new();
        let timestamp = order.timestamp;

        loop {
            let remaining = order.quantity - order.filled_quantity;
            if remaining <= 0.0 {
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
                while !queue.is_empty() && (order.quantity - order.filled_quantity) > 0.0 {
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
                        price: best_price.into_inner(),
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
        let opposite_book = match order.side {
            Side::Bid => &self.orderbook.asks,
            Side::Ask => &self.orderbook.bids,
        };

        let mut available = 0.0;
        for (price, queue) in opposite_book.iter() {
            let price_ok = match order.side {
                Side::Bid => *price <= order.price,
                Side::Ask => *price >= order.price,
            };
            if !price_ok {
                break;
            }
            for o in queue.iter() {
                available += o.quantity - o.filled_quantity;
                if available >= order.quantity {
                    return true;
                }
            }
        }
        available >= order.quantity
    }
}

impl Order {
    pub fn status(&self) -> OrderStatus {
        if self.filled_quantity == 0.0 {
            OrderStatus::New
        } else if self.filled_quantity >= self.quantity {
            OrderStatus::Filled
        } else {
            OrderStatus::PartiallyFilled
        }
    }
}
