use std::collections::{BTreeMap, VecDeque};

use rust_decimal::Decimal;

use crate::types::{Order, Side};

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

        for (_price, orders) in book.iter_mut() {
            if let Some(pos) = orders.iter().position(|o| o.id == order_id) {
                return orders.remove(pos);
            }
        }
        None
    }

    pub fn best_bid(&self) -> Option<Decimal> {
        self.bids.keys().next_back().copied()
    }

    pub fn best_ask(&self) -> Option<Decimal> {
        self.asks.keys().next().copied()
    }

    pub fn depth(&self, levels: usize) -> (Vec<(Decimal, Decimal)>, Vec<(Decimal, Decimal)>) {
        let bids: Vec<(Decimal, Decimal)> = self
            .bids
            .iter()
            .rev()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: Decimal = orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty)
            })
            .collect();

        let asks: Vec<(Decimal, Decimal)> = self
            .asks
            .iter()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: Decimal = orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (*price, total_qty)
            })
            .collect();

        (bids, asks)
    }
}
