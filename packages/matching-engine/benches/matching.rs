use computecoin_matching_engine::engine::MatchingEngine;
use computecoin_matching_engine::types::{Order, OrderType, Side};
use criterion::{black_box, criterion_group, criterion_main, Criterion};
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

fn bench_single_insert(c: &mut Criterion) {
    c.bench_function("single_insert", |b| {
        b.iter_batched(
            || {
                let engine = MatchingEngine::new("CU/USDT".to_string());
                let order = make_order("alice", Side::Bid, OrderType::Limit, dec!(100), dec!(1));
                (engine, order)
            },
            |(mut engine, order)| {
                black_box(engine.process_order(order).unwrap());
            },
            criterion::BatchSize::SmallInput,
        );
    });
}

fn bench_match_single(c: &mut Criterion) {
    c.bench_function("match_single", |b| {
        b.iter_batched(
            || {
                let mut engine = MatchingEngine::new("CU/USDT".to_string());
                let sell = make_order("alice", Side::Ask, OrderType::Limit, dec!(100), dec!(1));
                engine.process_order(sell).unwrap();
                let buy = make_order("bob", Side::Bid, OrderType::Limit, dec!(100), dec!(1));
                (engine, buy)
            },
            |(mut engine, buy)| {
                black_box(engine.process_order(buy).unwrap());
            },
            criterion::BatchSize::SmallInput,
        );
    });
}

fn bench_cancel_order(c: &mut Criterion) {
    c.bench_function("cancel_order_from_100", |b| {
        b.iter_batched(
            || {
                let mut engine = MatchingEngine::new("CU/USDT".to_string());
                let mut ids = Vec::new();
                for i in 0..100u64 {
                    let order = Order {
                        id: Uuid::new_v4(),
                        user_id: format!("user_{}", i),
                        pair: "CU/USDT".to_string(),
                        side: Side::Bid,
                        order_type: OrderType::Limit,
                        price: Decimal::from(100 + i),
                        quantity: dec!(1),
                        filled_quantity: Decimal::ZERO,
                        timestamp: 1000 + i,
                        stop_price: None,
                    };
                    ids.push(order.id);
                    engine.process_order(order).unwrap();
                }
                let target_id = ids[50];
                (engine, target_id)
            },
            |(mut engine, target_id)| {
                black_box(engine.cancel_order(target_id).unwrap());
            },
            criterion::BatchSize::SmallInput,
        );
    });
}

fn bench_match_deep_book(c: &mut Criterion) {
    c.bench_function("match_deep_book_1000_levels", |b| {
        b.iter_batched(
            || {
                let mut engine = MatchingEngine::new("CU/USDT".to_string());
                // Place 1000 sell orders at prices 100..1099
                for i in 0..1000u64 {
                    let order = Order {
                        id: Uuid::new_v4(),
                        user_id: format!("seller_{}", i),
                        pair: "CU/USDT".to_string(),
                        side: Side::Ask,
                        order_type: OrderType::Limit,
                        price: Decimal::from(100 + i),
                        quantity: dec!(1),
                        filled_quantity: Decimal::ZERO,
                        timestamp: 1000 + i,
                        stop_price: None,
                    };
                    engine.process_order(order).unwrap();
                }
                // A market buy that fills against the best level (price 100)
                let buy = make_order("buyer", Side::Bid, OrderType::Market, dec!(0), dec!(1));
                (engine, buy)
            },
            |(mut engine, buy)| {
                black_box(engine.process_order(buy).unwrap());
            },
            criterion::BatchSize::SmallInput,
        );
    });
}

fn bench_orderbook_depth(c: &mut Criterion) {
    c.bench_function("orderbook_depth_10", |b| {
        let mut engine = MatchingEngine::new("CU/USDT".to_string());
        // Place 100 bid levels and 100 ask levels
        for i in 0..100u64 {
            let bid = Order {
                id: Uuid::new_v4(),
                user_id: format!("bidder_{}", i),
                pair: "CU/USDT".to_string(),
                side: Side::Bid,
                order_type: OrderType::Limit,
                price: Decimal::from(100 + i),
                quantity: dec!(5),
                filled_quantity: Decimal::ZERO,
                timestamp: 1000 + i,
                stop_price: None,
            };
            engine.process_order(bid).unwrap();

            let ask = Order {
                id: Uuid::new_v4(),
                user_id: format!("asker_{}", i),
                pair: "CU/USDT".to_string(),
                side: Side::Ask,
                order_type: OrderType::Limit,
                price: Decimal::from(200 + i),
                quantity: dec!(5),
                filled_quantity: Decimal::ZERO,
                timestamp: 1000 + i,
                stop_price: None,
            };
            engine.process_order(ask).unwrap();
        }

        b.iter(|| {
            black_box(engine.orderbook.depth(10));
        });
    });
}

criterion_group!(
    benches,
    bench_single_insert,
    bench_match_single,
    bench_cancel_order,
    bench_match_deep_book,
    bench_orderbook_depth,
);
criterion_main!(benches);
