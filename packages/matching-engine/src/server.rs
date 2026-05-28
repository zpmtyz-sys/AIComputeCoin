use std::pin::Pin;
use std::sync::Arc;

use rust_decimal::Decimal;
use tokio::sync::{broadcast, Mutex};
use tokio_stream::wrappers::BroadcastStream;
use tokio_stream::Stream;
use tokio_stream::StreamExt;
use tonic::{Request, Response, Status};
use uuid::Uuid;

use crate::engine::MatchingEngine;
use crate::types::{OrderType, Side};

pub mod proto {
    tonic::include_proto!("computecoin.v1");
}

use proto::matching_service_server::MatchingService;

pub struct MatchingServiceImpl {
    engine: Arc<Mutex<MatchingEngine>>,
    trade_tx: broadcast::Sender<proto::Trade>,
    orderbook_tx: broadcast::Sender<proto::GetOrderBookResponse>,
}

impl MatchingServiceImpl {
    pub fn new(pair: String) -> Self {
        let engine = Arc::new(Mutex::new(MatchingEngine::new(pair)));
        let (trade_tx, _) = broadcast::channel(1024);
        let (orderbook_tx, _) = broadcast::channel(1024);
        Self {
            engine,
            trade_tx,
            orderbook_tx,
        }
    }
}

#[tonic::async_trait]
impl MatchingService for MatchingServiceImpl {
    type SubscribeOrderBookStream =
        Pin<Box<dyn Stream<Item = Result<proto::GetOrderBookResponse, Status>> + Send>>;
    type SubscribeTradesStream = Pin<Box<dyn Stream<Item = Result<proto::Trade, Status>> + Send>>;

    async fn submit_order(
        &self,
        request: Request<proto::SubmitOrderRequest>,
    ) -> Result<Response<proto::SubmitOrderResponse>, Status> {
        let req = request.into_inner();
        let proto_order = req
            .order
            .ok_or_else(|| Status::invalid_argument("order is required"))?;

        let mut internal_order = proto_order_to_internal(&proto_order)?;

        let mut engine = self.engine.lock().await;
        let trades = engine
            .process_order(internal_order.clone())
            .map_err(|e| Status::internal(e.to_string()))?;

        // Update filled_quantity from the engine result
        // The process_order mutates the order in place via &mut, but we cloned before.
        // Re-read the filled quantity from trades.
        let total_filled: Decimal = trades.iter().map(|t| t.quantity).sum();
        internal_order.filled_quantity = total_filled;

        let proto_trades: Vec<proto::Trade> = trades.iter().map(internal_trade_to_proto).collect();

        // Broadcast trades
        for t in &proto_trades {
            let _ = self.trade_tx.send(t.clone());
        }

        // Broadcast order book update
        let (bids, asks) = engine.orderbook.depth(20);
        let ob_response = proto::GetOrderBookResponse {
            pair: engine.orderbook.pair.clone(),
            bids: bids
                .iter()
                .map(|(price, qty)| proto::PriceLevel {
                    price: price.to_string(),
                    quantity: qty.to_string(),
                    order_count: 0,
                })
                .collect(),
            asks: asks
                .iter()
                .map(|(price, qty)| proto::PriceLevel {
                    price: price.to_string(),
                    quantity: qty.to_string(),
                    order_count: 0,
                })
                .collect(),
            sequence_id: 0,
        };
        let _ = self.orderbook_tx.send(ob_response);

        let response_order = internal_order_to_proto(&internal_order);
        Ok(Response::new(proto::SubmitOrderResponse {
            order: Some(response_order),
            trades: proto_trades,
        }))
    }

    async fn cancel_order(
        &self,
        request: Request<proto::CancelOrderRequest>,
    ) -> Result<Response<proto::CancelOrderResponse>, Status> {
        let req = request.into_inner();
        let order_id = Uuid::parse_str(&req.order_id)
            .map_err(|e| Status::invalid_argument(format!("invalid order_id: {}", e)))?;

        let mut engine = self.engine.lock().await;
        let order = engine
            .cancel_order(order_id)
            .map_err(|e| Status::not_found(e.to_string()))?;

        Ok(Response::new(proto::CancelOrderResponse {
            order: Some(internal_order_to_proto(&order)),
        }))
    }

    async fn get_order_book(
        &self,
        request: Request<proto::GetOrderBookRequest>,
    ) -> Result<Response<proto::GetOrderBookResponse>, Status> {
        let req = request.into_inner();
        let depth = if req.depth == 0 {
            20
        } else {
            req.depth as usize
        };

        let engine = self.engine.lock().await;
        let (bids, asks) = engine.orderbook.depth(depth);

        let response = proto::GetOrderBookResponse {
            pair: req.pair,
            bids: bids
                .iter()
                .map(|(price, qty)| proto::PriceLevel {
                    price: price.to_string(),
                    quantity: qty.to_string(),
                    order_count: 0,
                })
                .collect(),
            asks: asks
                .iter()
                .map(|(price, qty)| proto::PriceLevel {
                    price: price.to_string(),
                    quantity: qty.to_string(),
                    order_count: 0,
                })
                .collect(),
            sequence_id: 0,
        };

        Ok(Response::new(response))
    }

    async fn get_trades(
        &self,
        request: Request<proto::GetTradesRequest>,
    ) -> Result<Response<proto::GetTradesResponse>, Status> {
        let req = request.into_inner();
        let limit = if req.limit == 0 {
            100
        } else {
            req.limit as usize
        };

        let engine = self.engine.lock().await;
        let trades = engine.get_trades(limit);

        let proto_trades: Vec<proto::Trade> = trades.iter().map(internal_trade_to_proto).collect();

        Ok(Response::new(proto::GetTradesResponse {
            trades: proto_trades,
        }))
    }

    async fn subscribe_order_book(
        &self,
        _request: Request<proto::GetOrderBookRequest>,
    ) -> Result<Response<Self::SubscribeOrderBookStream>, Status> {
        let rx = self.orderbook_tx.subscribe();
        let stream = BroadcastStream::new(rx).filter_map(|result| match result {
            Ok(msg) => Some(Ok(msg)),
            Err(tokio_stream::wrappers::errors::BroadcastStreamRecvError::Lagged(_)) => None,
        });
        Ok(Response::new(Box::pin(stream)))
    }

    async fn subscribe_trades(
        &self,
        _request: Request<proto::SubscribeTradesRequest>,
    ) -> Result<Response<Self::SubscribeTradesStream>, Status> {
        let rx = self.trade_tx.subscribe();
        let stream = BroadcastStream::new(rx).filter_map(|result| match result {
            Ok(msg) => Some(Ok(msg)),
            Err(tokio_stream::wrappers::errors::BroadcastStreamRecvError::Lagged(_)) => None,
        });
        Ok(Response::new(Box::pin(stream)))
    }
}

fn internal_order_to_proto(order: &crate::types::Order) -> proto::Order {
    let side = match order.side {
        Side::Bid => proto::Side::Buy as i32,
        Side::Ask => proto::Side::Sell as i32,
    };

    let order_type = match order.order_type {
        OrderType::Limit => proto::OrderType::Limit as i32,
        OrderType::Market => proto::OrderType::Market as i32,
        OrderType::StopLimit => proto::OrderType::StopLimit as i32,
        OrderType::IOC => proto::OrderType::Ioc as i32,
        OrderType::FOK => proto::OrderType::Fok as i32,
    };

    let status = {
        let s = order.status();
        match s {
            crate::types::OrderStatus::New => proto::OrderStatus::New as i32,
            crate::types::OrderStatus::PartiallyFilled => {
                proto::OrderStatus::PartiallyFilled as i32
            }
            crate::types::OrderStatus::Filled => proto::OrderStatus::Filled as i32,
            crate::types::OrderStatus::Cancelled => proto::OrderStatus::Cancelled as i32,
        }
    };

    proto::Order {
        id: order.id.to_string(),
        user_id: order.user_id.clone(),
        pair: order.pair.clone(),
        side,
        order_type,
        price: order.price.to_string(),
        quantity: order.quantity.to_string(),
        filled_quantity: order.filled_quantity.to_string(),
        status,
        created_at: order.timestamp as i64,
        updated_at: order.timestamp as i64,
    }
}

#[allow(clippy::result_large_err)]
fn proto_order_to_internal(proto_order: &proto::Order) -> Result<crate::types::Order, Status> {
    let side = match proto::Side::try_from(proto_order.side) {
        Ok(proto::Side::Buy) => Side::Bid,
        Ok(proto::Side::Sell) => Side::Ask,
        _ => return Err(Status::invalid_argument("invalid side")),
    };

    let order_type = match proto::OrderType::try_from(proto_order.order_type) {
        Ok(proto::OrderType::Limit) => OrderType::Limit,
        Ok(proto::OrderType::Market) => OrderType::Market,
        Ok(proto::OrderType::StopLimit) => OrderType::StopLimit,
        Ok(proto::OrderType::Ioc) => OrderType::IOC,
        Ok(proto::OrderType::Fok) => OrderType::FOK,
        _ => return Err(Status::invalid_argument("invalid order_type")),
    };

    let price: Decimal = proto_order
        .price
        .parse()
        .map_err(|e| Status::invalid_argument(format!("invalid price: {}", e)))?;

    let quantity: Decimal = proto_order
        .quantity
        .parse()
        .map_err(|e| Status::invalid_argument(format!("invalid quantity: {}", e)))?;

    let filled_quantity: Decimal = if proto_order.filled_quantity.is_empty() {
        Decimal::ZERO
    } else {
        proto_order
            .filled_quantity
            .parse()
            .map_err(|e| Status::invalid_argument(format!("invalid filled_quantity: {}", e)))?
    };

    let id = if proto_order.id.is_empty() {
        Uuid::new_v4()
    } else {
        Uuid::parse_str(&proto_order.id)
            .map_err(|e| Status::invalid_argument(format!("invalid id: {}", e)))?
    };

    Ok(crate::types::Order {
        id,
        user_id: proto_order.user_id.clone(),
        pair: proto_order.pair.clone(),
        side,
        order_type,
        price,
        quantity,
        filled_quantity,
        stop_price: None,
        timestamp: proto_order.created_at as u64,
    })
}

fn internal_trade_to_proto(trade: &crate::types::Trade) -> proto::Trade {
    let side = match trade.side {
        Side::Bid => proto::Side::Buy as i32,
        Side::Ask => proto::Side::Sell as i32,
    };

    proto::Trade {
        id: trade.id.to_string(),
        order_id: trade.taker_order_id.to_string(),
        maker_order_id: trade.maker_order_id.to_string(),
        taker_order_id: trade.taker_order_id.to_string(),
        pair: trade.pair.clone(),
        price: trade.price.to_string(),
        quantity: trade.quantity.to_string(),
        side,
        traded_at: trade.timestamp as i64,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal_macros::dec;

    #[test]
    fn test_proto_to_internal_roundtrip() {
        let proto_order = proto::Order {
            id: "550e8400-e29b-41d4-a716-446655440000".to_string(),
            user_id: "user123".to_string(),
            pair: "CU/USDT".to_string(),
            side: proto::Side::Buy as i32,
            order_type: proto::OrderType::Limit as i32,
            price: "100.50".to_string(),
            quantity: "10.25".to_string(),
            filled_quantity: "0".to_string(),
            status: proto::OrderStatus::New as i32,
            created_at: 1700000000,
            updated_at: 1700000000,
        };

        let internal = proto_order_to_internal(&proto_order).unwrap();
        assert_eq!(
            internal.id.to_string(),
            "550e8400-e29b-41d4-a716-446655440000"
        );
        assert_eq!(internal.user_id, "user123");
        assert_eq!(internal.pair, "CU/USDT");
        assert_eq!(internal.side, Side::Bid);
        assert_eq!(internal.order_type, OrderType::Limit);
        assert_eq!(internal.price, dec!(100.50));
        assert_eq!(internal.quantity, dec!(10.25));
        assert_eq!(internal.filled_quantity, Decimal::ZERO);
        assert_eq!(internal.timestamp, 1700000000);

        // Convert back to proto
        let back_to_proto = internal_order_to_proto(&internal);
        assert_eq!(back_to_proto.id, proto_order.id);
        assert_eq!(back_to_proto.user_id, proto_order.user_id);
        assert_eq!(back_to_proto.pair, proto_order.pair);
        assert_eq!(back_to_proto.side, proto_order.side);
        assert_eq!(back_to_proto.order_type, proto_order.order_type);
        assert_eq!(back_to_proto.price, "100.50");
        assert_eq!(back_to_proto.quantity, "10.25");
        assert_eq!(back_to_proto.filled_quantity, "0");
        assert_eq!(back_to_proto.status, proto::OrderStatus::New as i32);
    }

    #[test]
    fn test_internal_to_proto_conversion() {
        let order = crate::types::Order {
            id: Uuid::parse_str("550e8400-e29b-41d4-a716-446655440000").unwrap(),
            user_id: "user456".to_string(),
            pair: "BTC/USDT".to_string(),
            side: Side::Ask,
            order_type: OrderType::Market,
            price: dec!(50000),
            quantity: dec!(1.5),
            filled_quantity: dec!(0.5),
            stop_price: None,
            timestamp: 1700000001,
        };

        let proto_order = internal_order_to_proto(&order);
        assert_eq!(proto_order.id, "550e8400-e29b-41d4-a716-446655440000");
        assert_eq!(proto_order.user_id, "user456");
        assert_eq!(proto_order.pair, "BTC/USDT");
        assert_eq!(proto_order.side, proto::Side::Sell as i32);
        assert_eq!(proto_order.order_type, proto::OrderType::Market as i32);
        assert_eq!(proto_order.price, "50000");
        assert_eq!(proto_order.quantity, "1.5");
        assert_eq!(proto_order.filled_quantity, "0.5");
        assert_eq!(
            proto_order.status,
            proto::OrderStatus::PartiallyFilled as i32
        );
    }

    #[test]
    fn test_internal_trade_to_proto() {
        let trade = crate::types::Trade {
            id: Uuid::parse_str("660e8400-e29b-41d4-a716-446655440000").unwrap(),
            maker_order_id: Uuid::parse_str("770e8400-e29b-41d4-a716-446655440000").unwrap(),
            taker_order_id: Uuid::parse_str("880e8400-e29b-41d4-a716-446655440000").unwrap(),
            pair: "CU/USDT".to_string(),
            price: dec!(200.75),
            quantity: dec!(5),
            side: Side::Bid,
            timestamp: 1700000002,
        };

        let proto_trade = internal_trade_to_proto(&trade);
        assert_eq!(proto_trade.id, "660e8400-e29b-41d4-a716-446655440000");
        assert_eq!(
            proto_trade.maker_order_id,
            "770e8400-e29b-41d4-a716-446655440000"
        );
        assert_eq!(
            proto_trade.taker_order_id,
            "880e8400-e29b-41d4-a716-446655440000"
        );
        assert_eq!(proto_trade.pair, "CU/USDT");
        assert_eq!(proto_trade.price, "200.75");
        assert_eq!(proto_trade.quantity, "5");
        assert_eq!(proto_trade.side, proto::Side::Buy as i32);
        assert_eq!(proto_trade.traded_at, 1700000002);
    }

    #[test]
    fn test_proto_order_empty_id_generates_uuid() {
        let proto_order = proto::Order {
            id: "".to_string(),
            user_id: "user789".to_string(),
            pair: "CU/USDT".to_string(),
            side: proto::Side::Sell as i32,
            order_type: proto::OrderType::Ioc as i32,
            price: "75.00".to_string(),
            quantity: "20".to_string(),
            filled_quantity: "".to_string(),
            status: proto::OrderStatus::New as i32,
            created_at: 0,
            updated_at: 0,
        };

        let internal = proto_order_to_internal(&proto_order).unwrap();
        // Should have generated a new UUID (not nil)
        assert!(!internal.id.is_nil());
        assert_eq!(internal.side, Side::Ask);
        assert_eq!(internal.order_type, OrderType::IOC);
        assert_eq!(internal.filled_quantity, Decimal::ZERO);
    }

    #[test]
    fn test_proto_order_invalid_side_returns_error() {
        let proto_order = proto::Order {
            id: "".to_string(),
            user_id: "user".to_string(),
            pair: "CU/USDT".to_string(),
            side: 0, // SIDE_UNSPECIFIED
            order_type: proto::OrderType::Limit as i32,
            price: "100".to_string(),
            quantity: "10".to_string(),
            filled_quantity: "0".to_string(),
            status: proto::OrderStatus::New as i32,
            created_at: 0,
            updated_at: 0,
        };

        let result = proto_order_to_internal(&proto_order);
        assert!(result.is_err());
    }

    #[test]
    fn test_proto_order_invalid_price_returns_error() {
        let proto_order = proto::Order {
            id: "".to_string(),
            user_id: "user".to_string(),
            pair: "CU/USDT".to_string(),
            side: proto::Side::Buy as i32,
            order_type: proto::OrderType::Limit as i32,
            price: "not_a_number".to_string(),
            quantity: "10".to_string(),
            filled_quantity: "0".to_string(),
            status: proto::OrderStatus::New as i32,
            created_at: 0,
            updated_at: 0,
        };

        let result = proto_order_to_internal(&proto_order);
        assert!(result.is_err());
    }
}
