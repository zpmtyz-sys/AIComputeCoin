use computecoin_matching_engine::engine::MatchingEngine;
use computecoin_matching_engine::types::{Order, OrderType, Side};
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use uuid::Uuid;

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
        pair: "CU/USDT".to_string(),
        side,
        order_type,
        price,
        quantity,
        filled_quantity: Decimal::ZERO,
        timestamp: 1000,
        stop_price: None,
    }
}

fn make_order_with_ts(
    user_id: &str,
    side: Side,
    order_type: OrderType,
    price: Decimal,
    quantity: Decimal,
    timestamp: u64,
) -> Order {
    Order {
        id: Uuid::new_v4(),
        user_id: user_id.to_string(),
        pair: "CU/USDT".to_string(),
        side,
        order_type,
        price,
        quantity,
        filled_quantity: Decimal::ZERO,
        timestamp,
        stop_price: None,
    }
}

#[test]
fn test_limit_buy_sell_crossing() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    let buy = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(5));
    engine.process_order(buy).unwrap();

    let sell = make_order("bob", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
    let trades = engine.process_order(sell).unwrap();

    assert_eq!(trades.len(), 1);
    assert_eq!(trades[0].price, dec!(100));
    assert_eq!(trades[0].quantity, dec!(5));
}

#[test]
fn test_partial_fill_leaves_remainder() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    let buy = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(10));
    let buy_id = buy.id;
    engine.process_order(buy).unwrap();

    // Sell 3 against the buy of 10
    let sell = make_order("bob", Side::Ask, OrderType::Limit, dec!(100), dec!(3));
    let trades = engine.process_order(sell).unwrap();

    assert_eq!(trades.len(), 1);
    assert_eq!(trades[0].quantity, dec!(3));

    // The buy should still be on the book with 1 order remaining
    assert_eq!(engine.orderbook.order_count(), 1);
    // Best bid should still be 100
    assert_eq!(engine.orderbook.best_bid(), Some(dec!(100)));

    // Now sell the remaining 7
    let sell2 = make_order("charlie", Side::Ask, OrderType::Limit, dec!(100), dec!(7));
    let trades2 = engine.process_order(sell2).unwrap();

    assert_eq!(trades2.len(), 1);
    assert_eq!(trades2[0].quantity, dec!(7));
    assert_eq!(trades2[0].maker_order_id, buy_id);

    // Book should now be empty
    assert_eq!(engine.orderbook.order_count(), 0);
}

#[test]
fn test_price_time_priority() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place 3 buy orders at the same price with different timestamps
    let buy1 = make_order_with_ts(
        "alice",
        Side::Bid,
        OrderType::Limit,
        dec!(100),
        dec!(1),
        100,
    );
    let buy2 = make_order_with_ts("bob", Side::Bid, OrderType::Limit, dec!(100), dec!(1), 200);
    let buy3 = make_order_with_ts(
        "charlie",
        Side::Bid,
        OrderType::Limit,
        dec!(100),
        dec!(1),
        300,
    );

    let id1 = buy1.id;
    let id2 = buy2.id;
    let id3 = buy3.id;

    engine.process_order(buy1).unwrap();
    engine.process_order(buy2).unwrap();
    engine.process_order(buy3).unwrap();

    // Sell 3 units at 100 - should fill in FIFO order
    let sell = make_order("dave", Side::Ask, OrderType::Limit, dec!(100), dec!(3));
    let trades = engine.process_order(sell).unwrap();

    assert_eq!(trades.len(), 3);
    assert_eq!(trades[0].maker_order_id, id1);
    assert_eq!(trades[1].maker_order_id, id2);
    assert_eq!(trades[2].maker_order_id, id3);
}

#[test]
fn test_multiple_price_levels() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place buys at different prices
    let buy_100 = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(1));
    let buy_101 = make_order("bob", Side::Bid, OrderType::Limit, dec!(101), dec!(1));
    let buy_102 = make_order("charlie", Side::Bid, OrderType::Limit, dec!(102), dec!(1));

    let id_100 = buy_100.id;
    let id_101 = buy_101.id;
    let id_102 = buy_102.id;

    engine.process_order(buy_100).unwrap();
    engine.process_order(buy_101).unwrap();
    engine.process_order(buy_102).unwrap();

    // Sell 3 at 100 (will cross all bids)
    let sell = make_order("dave", Side::Ask, OrderType::Limit, dec!(100), dec!(3));
    let trades = engine.process_order(sell).unwrap();

    assert_eq!(trades.len(), 3);
    // Highest bid fills first
    assert_eq!(trades[0].maker_order_id, id_102);
    assert_eq!(trades[0].price, dec!(102));
    assert_eq!(trades[1].maker_order_id, id_101);
    assert_eq!(trades[1].price, dec!(101));
    assert_eq!(trades[2].maker_order_id, id_100);
    assert_eq!(trades[2].price, dec!(100));
}

#[test]
fn test_cancel_order_removes_from_book() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    let buy = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(5));
    let order_id = buy.id;
    engine.process_order(buy).unwrap();

    assert_eq!(engine.orderbook.order_count(), 1);

    let cancelled = engine.cancel_order(order_id).unwrap();
    assert_eq!(cancelled.id, order_id);
    assert_eq!(engine.orderbook.order_count(), 0);
    assert_eq!(engine.orderbook.best_bid(), None);
}

#[test]
fn test_ioc_partial_fill_remainder_discarded() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place a resting sell for qty 5 at 100
    let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
    engine.process_order(sell).unwrap();

    // Submit IOC buy for qty 10 at 100
    let ioc_buy = make_order("bob", Side::Bid, OrderType::IOC, dec!(100), dec!(10));
    let trades = engine.process_order(ioc_buy).unwrap();

    // Should trade qty 5
    assert_eq!(trades.len(), 1);
    assert_eq!(trades[0].quantity, dec!(5));

    // Nothing remains on the book (the unfilled 5 of the IOC is discarded)
    assert_eq!(engine.orderbook.order_count(), 0);
}

#[test]
fn test_fok_rejected_insufficient_liquidity() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place sell for qty 3 at 100
    let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(3));
    engine.process_order(sell).unwrap();

    // Submit FOK buy for qty 10 at 100 - insufficient liquidity
    let fok_buy = make_order("bob", Side::Bid, OrderType::FOK, dec!(100), dec!(10));
    let trades = engine.process_order(fok_buy).unwrap();

    // No trades produced
    assert!(trades.is_empty());

    // Sell still on book
    assert_eq!(engine.orderbook.order_count(), 1);
    assert_eq!(engine.orderbook.best_ask(), Some(dec!(100)));
}

#[test]
fn test_fok_filled_sufficient_liquidity() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place sell for qty 10 at 100
    let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(10));
    engine.process_order(sell).unwrap();

    // Submit FOK buy for qty 10 at 100
    let fok_buy = make_order("bob", Side::Bid, OrderType::FOK, dec!(100), dec!(10));
    let trades = engine.process_order(fok_buy).unwrap();

    assert_eq!(trades.len(), 1);
    assert_eq!(trades[0].quantity, dec!(10));
    assert_eq!(engine.orderbook.order_count(), 0);
}

#[test]
fn test_stop_limit_activation() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place a resting sell at 100
    let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
    engine.process_order(sell).unwrap();

    // Submit a stop-limit buy with stop_price 99 and price 100, qty 5
    let mut stop_buy = make_order(
        "charlie",
        Side::Bid,
        OrderType::StopLimit,
        dec!(100),
        dec!(5),
    );
    stop_buy.stop_price = Some(dec!(99));
    engine.process_order(stop_buy).unwrap();

    // The stop order should be pending
    assert_eq!(engine.pending_stops.len(), 1);

    // Place a regular sell at 99 and a regular buy at 99 to generate a trade at price 99
    let sell_99 = make_order("alice", Side::Ask, OrderType::Limit, dec!(99), dec!(2));
    engine.process_order(sell_99).unwrap();

    let buy_99 = make_order("bob", Side::Bid, OrderType::Limit, dec!(99), dec!(2));
    let trades = engine.process_order(buy_99).unwrap();

    // A trade at price 99 should activate the stop-limit buy (stop_price 99 <= last_trade 99)
    assert!(!trades.is_empty());
    // The stop order should have been activated
    assert_eq!(engine.pending_stops.len(), 0);

    // The activated stop order should have traded against the resting sell at 100
    // Check total trades in history
    let all_trades = engine.get_trades(100);
    // We should have: the trade at 99 + the stop-limit trade at 100
    assert!(all_trades.len() >= 2);
}

#[test]
fn test_self_trade_prevention() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place a buy with user "alice"
    let buy = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(5));
    engine.process_order(buy).unwrap();

    assert_eq!(engine.orderbook.order_count(), 1);

    // Submit a sell with user "alice" at matching price
    let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(5));
    let trades = engine.process_order(sell).unwrap();

    // No trade should be generated (self-trade prevented)
    assert!(trades.is_empty());

    // The maker order was removed from the book by self-trade prevention
    // The sell should be placed on the book since it could not match
    // Verify the original buy is gone (popped by STP)
    assert_eq!(engine.orderbook.best_bid(), None);
}

#[test]
fn test_depth_and_snapshot() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());

    // Place multiple buy orders at different price levels
    let buy1 = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(5));
    let buy2 = make_order("bob", Side::Bid, OrderType::Limit, dec!(100), dec!(3));
    let buy3 = make_order("charlie", Side::Bid, OrderType::Limit, dec!(99), dec!(7));
    let buy4 = make_order("dave", Side::Bid, OrderType::Limit, dec!(98), dec!(2));
    let buy5 = make_order("eve", Side::Bid, OrderType::Limit, dec!(97), dec!(4));

    engine.process_order(buy1).unwrap();
    engine.process_order(buy2).unwrap();
    engine.process_order(buy3).unwrap();
    engine.process_order(buy4).unwrap();
    engine.process_order(buy5).unwrap();

    // Place multiple sell orders
    let sell1 = make_order("frank", Side::Ask, OrderType::Limit, dec!(101), dec!(6));
    let sell2 = make_order("grace", Side::Ask, OrderType::Limit, dec!(102), dec!(4));
    let sell3 = make_order("heidi", Side::Ask, OrderType::Limit, dec!(103), dec!(8));

    engine.process_order(sell1).unwrap();
    engine.process_order(sell2).unwrap();
    engine.process_order(sell3).unwrap();

    // Test depth(3) - should return top 3 levels each side
    let (bids, asks) = engine.orderbook.depth(3);

    assert_eq!(bids.len(), 3);
    // Bids are returned from highest to lowest
    assert_eq!(bids[0], (dec!(100), dec!(8))); // 5 + 3
    assert_eq!(bids[1], (dec!(99), dec!(7)));
    assert_eq!(bids[2], (dec!(98), dec!(2)));

    assert_eq!(asks.len(), 3);
    // Asks are returned from lowest to highest
    assert_eq!(asks[0], (dec!(101), dec!(6)));
    assert_eq!(asks[1], (dec!(102), dec!(4)));
    assert_eq!(asks[2], (dec!(103), dec!(8)));

    // Test snapshot
    let snapshot = engine.orderbook.snapshot();

    // Snapshot includes all levels
    assert_eq!(snapshot.bids.len(), 4); // 100, 99, 98, 97
    assert_eq!(snapshot.asks.len(), 3); // 101, 102, 103

    // Verify aggregation with order counts
    // Price 100 has 2 orders totalling qty 8
    assert_eq!(snapshot.bids[0], (dec!(100), dec!(8), 2));
    assert_eq!(snapshot.bids[1], (dec!(99), dec!(7), 1));

    assert_eq!(snapshot.asks[0], (dec!(101), dec!(6), 1));
    assert_eq!(snapshot.asks[1], (dec!(102), dec!(4), 1));
    assert_eq!(snapshot.asks[2], (dec!(103), dec!(8), 1));
}

#[test]
fn test_stress_1000_orders() {
    let mut engine = MatchingEngine::new("CU/USDT".to_string());
    let mut total_trades = 0;

    // Insert 500 buy limit orders at price 100 (all at same price for guaranteed matching)
    for i in 0..500u64 {
        let buy = Order {
            id: Uuid::new_v4(),
            user_id: format!("buyer_{}", i),
            pair: "CU/USDT".to_string(),
            side: Side::Bid,
            order_type: OrderType::Limit,
            price: dec!(100),
            quantity: dec!(1),
            filled_quantity: Decimal::ZERO,
            timestamp: 1000 + i,
            stop_price: None,
        };
        let trades = engine.process_order(buy).unwrap();
        total_trades += trades.len();
    }

    // No trades should have happened yet (all buys, no sells)
    assert_eq!(total_trades, 0);
    assert_eq!(engine.orderbook.order_count(), 500);

    // Insert 500 sell orders at price 100 that will cross with the buys
    for i in 0..500u64 {
        let sell = Order {
            id: Uuid::new_v4(),
            user_id: format!("seller_{}", i),
            pair: "CU/USDT".to_string(),
            side: Side::Ask,
            order_type: OrderType::Limit,
            price: dec!(100),
            quantity: dec!(1),
            filled_quantity: Decimal::ZERO,
            timestamp: 2000 + i,
            stop_price: None,
        };
        let trades = engine.process_order(sell).unwrap();
        total_trades += trades.len();
    }

    // All 500 sells should have matched against buys
    assert_eq!(total_trades, 500);

    // Verify engine consistency
    let all_trades = engine.get_trades(1000);
    assert_eq!(all_trades.len(), 500);

    // Book should be empty since all orders matched
    assert_eq!(engine.orderbook.order_count(), 0);
}
