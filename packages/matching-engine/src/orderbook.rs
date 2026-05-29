use std::collections::{BTreeMap, VecDeque};

use rust_decimal::Decimal;

use crate::types::{Order, Side};

/// Aggregated order book snapshot for market data feeds.
#[derive(Debug, Clone)]
pub struct OrderBookSnapshot {
    pub bids: Vec<(Decimal, Decimal, usize)>,
    pub asks: Vec<(Decimal, Decimal, usize)>,
}

/// Order book maintaining price-time priority for bids and asks.
pub struct OrderBook {
    pub bids: BTreeMap<Decimal, VecDeque<Order>>,
    pub asks: BTreeMap<Decimal, VecDeque<Order>>,
    pub pair: String,
}

impl OrderBook {
    pub fn new(pair: String) -> Self {
        Self {
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            pair,
        }
    }

    pub fn add_order(&mut self, order: Order) {
        let book = match order.side {
            Side::Bid => &mut self.bids,
            Side::Ask => &mut self.asks,
        };
        book.entry(order.price).or_default().push_back(order);
    }

    pub fn cancel_order(&mut self, order_id: uuid::Uuid, side: Side) -> Option<Order> {
        let book = match side {
            Side::Bid => &mut self.bids,
            Side::Ask => &mut self.asks,
        };

        let mut found_price = None;
        let mut found_order = None;

        for (price, orders) in book.iter_mut() {
            if let Some(pos) = orders.iter().position(|o| o.id == order_id) {
                found_order = orders.remove(pos);
                if orders.is_empty() {
                    found_price = Some(*price);
                }
                break;
            }
        }

        // Remove empty price level
        if let Some(price) = found_price {
            book.remove(&price);
        }

        found_order
    }

    pub fn best_bid(&self) -> Option<Decimal> {
        self.bids.keys().next_back().copied()
    }

    pub fn best_ask(&self) -> Option<Decimal> {
        self.asks.keys().next().copied()
    }

    #[allow(clippy::type_complexity)]
    pub fn depth(&self, levels: usize) -> (Vec<(Decimal, Decimal)>, Vec<(Decimal, Decimal)>) {
        let bids: Vec<(Decimal, Decimal)> = self
            .bids
            .iter()
            .rev()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: Decimal =
                    orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty)
            })
            .collect();

        let asks: Vec<(Decimal, Decimal)> = self
            .asks
            .iter()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: Decimal =
                    orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty)
            })
            .collect();

        (bids, asks)
    }

    /// Returns an aggregated snapshot of the order book with (price, total_quantity, order_count) tuples.
    pub fn snapshot(&self) -> OrderBookSnapshot {
        let bids: Vec<(Decimal, Decimal, usize)> = self
            .bids
            .iter()
            .rev()
            .map(|(price, orders)| {
                let total_qty: Decimal =
                    orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty, orders.len())
            })
            .collect();

        let asks: Vec<(Decimal, Decimal, usize)> = self
            .asks
            .iter()
            .map(|(price, orders)| {
                let total_qty: Decimal =
                    orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty, orders.len())
            })
            .collect();

        OrderBookSnapshot { bids, asks }
    }

    /// Returns the total number of orders across both sides.
    pub fn order_count(&self) -> usize {
        let bid_count: usize = self.bids.values().map(|q| q.len()).sum();
        let ask_count: usize = self.asks.values().map(|q| q.len()).sum();
        bid_count + ask_count
    }
}
