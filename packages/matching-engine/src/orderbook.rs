use std::collections::{BTreeMap, VecDeque};

use ordered_float::OrderedFloat;

use crate::types::{Order, Side};

/// Order book maintaining price-time priority for bids and asks.
pub struct OrderBook {
    pub bids: BTreeMap<OrderedFloat<f64>, VecDeque<Order>>,
    pub asks: BTreeMap<OrderedFloat<f64>, VecDeque<Order>>,
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

    pub fn best_bid(&self) -> Option<OrderedFloat<f64>> {
        self.bids.keys().next_back().copied()
    }

    pub fn best_ask(&self) -> Option<OrderedFloat<f64>> {
        self.asks.keys().next().copied()
    }

    pub fn depth(&self, levels: usize) -> (Vec<(f64, f64)>, Vec<(f64, f64)>) {
        let bids: Vec<(f64, f64)> = self
            .bids
            .iter()
            .rev()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: f64 = orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (price.into_inner(), total_qty)
            })
            .collect();

        let asks: Vec<(f64, f64)> = self
            .asks
            .iter()
            .take(levels)
            .map(|(price, orders)| {
                let total_qty: f64 = orders.iter().map(|o| o.quantity - o.filled_quantity).sum();
                (price.into_inner(), total_qty)
            })
            .collect();

        (bids, asks)
    }
}
