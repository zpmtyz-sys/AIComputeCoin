use ordered_float::OrderedFloat;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Side {
    Bid,
    Ask,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderType {
    Limit,
    Market,
    StopLimit,
    IOC,
    FOK,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    New,
    PartiallyFilled,
    Filled,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub user_id: String,
    pub pair: String,
    pub side: Side,
    pub order_type: OrderType,
    pub price: OrderedFloat<f64>,
    pub quantity: f64,
    pub filled_quantity: f64,
    pub timestamp: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Trade {
    pub id: Uuid,
    pub maker_order_id: Uuid,
    pub taker_order_id: Uuid,
    pub pair: String,
    pub price: f64,
    pub quantity: f64,
    pub side: Side,
    pub timestamp: u64,
}
