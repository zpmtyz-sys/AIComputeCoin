import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics per workload type
const orderSubmitLatency = new Trend('order_submit_latency');
const orderbookQueryLatency = new Trend('orderbook_query_latency');
const tradeHistoryLatency = new Trend('trade_history_latency');
const overallSuccess = new Rate('overall_success');
const requestsByType = new Counter('requests_by_type');

export const options = {
  scenarios: {
    order_submission: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 },
        { duration: '3m', target: 400 },
        { duration: '1m', target: 0 },
      ],
      exec: 'orderSubmission',
    },
    orderbook_queries: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 },
        { duration: '3m', target: 400 },
        { duration: '1m', target: 0 },
      ],
      exec: 'orderbookQuery',
    },
    trade_history: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '3m', target: 200 },
        { duration: '1m', target: 0 },
      ],
      exec: 'tradeHistory',
    },
  },
  thresholds: {
    'overall_success': ['rate>0.99'],
    'order_submit_latency': ['p(95)<100', 'p(99)<500'],
    'orderbook_query_latency': ['p(95)<50'],
    'trade_history_latency': ['p(95)<200'],
    'http_req_duration': ['p(95)<200'],
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'load-test-token';

const PAIRS = ['BTC-USDT', 'ETH-USDT', 'SOL-USDT', 'AVAX-USDT', 'DOT-USDT'];
const SIDES = ['buy', 'sell'];

function getHeaders() {
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
  };
}

function randomPair() {
  return PAIRS[Math.floor(Math.random() * PAIRS.length)];
}

function randomPrice(pair) {
  const basePrices = {
    'BTC-USDT': 50000,
    'ETH-USDT': 3500,
    'SOL-USDT': 100,
    'AVAX-USDT': 35,
    'DOT-USDT': 7,
  };
  const base = basePrices[pair] || 100;
  return base + (Math.random() - 0.5) * base * 0.04;
}

// 40% of traffic - Order submission
export function orderSubmission() {
  const pair = randomPair();
  const order = {
    pair: pair,
    side: SIDES[Math.floor(Math.random() * SIDES.length)],
    type: Math.random() > 0.3 ? 'limit' : 'market',
    price: randomPrice(pair),
    quantity: 0.01 + Math.random() * 5,
  };

  group('order_submission', function () {
    const startTime = Date.now();
    const res = http.post(`${BASE_URL}/orders`, JSON.stringify(order), getHeaders());
    orderSubmitLatency.add(Date.now() - startTime);
    requestsByType.add(1, { type: 'order' });

    const success = check(res, {
      'order created': (r) => r.status === 201,
    });
    overallSuccess.add(success ? 1 : 0);
  });

  sleep(0.05 + Math.random() * 0.1);
}

// 40% of traffic - Orderbook queries
export function orderbookQuery() {
  const pair = randomPair();

  group('orderbook_query', function () {
    const startTime = Date.now();
    const res = http.get(`${BASE_URL}/markets/${pair}/orderbook?depth=20`, getHeaders());
    orderbookQueryLatency.add(Date.now() - startTime);
    requestsByType.add(1, { type: 'orderbook' });

    const success = check(res, {
      'orderbook returned': (r) => r.status === 200,
      'has bid/ask data': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.bids && body.asks;
        } catch (e) {
          return false;
        }
      },
    });
    overallSuccess.add(success ? 1 : 0);
  });

  sleep(0.01 + Math.random() * 0.05);
}

// 20% of traffic - Trade history queries
export function tradeHistory() {
  const pair = randomPair();

  group('trade_history', function () {
    const startTime = Date.now();
    const res = http.get(`${BASE_URL}/markets/${pair}/trades?limit=50`, getHeaders());
    tradeHistoryLatency.add(Date.now() - startTime);
    requestsByType.add(1, { type: 'history' });

    const success = check(res, {
      'trade history returned': (r) => r.status === 200,
      'has trades array': (r) => {
        try {
          const body = JSON.parse(r.body);
          return Array.isArray(body.trades || body);
        } catch (e) {
          return false;
        }
      },
    });
    overallSuccess.add(success ? 1 : 0);
  });

  sleep(0.05 + Math.random() * 0.1);
}
