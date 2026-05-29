import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const orderSuccess = new Rate('order_success');
const orderLatency = new Trend('order_latency');

export const options = {
  stages: [
    { duration: '30s', target: 100 },    // ramp up
    { duration: '2m', target: 1000 },     // hold at 1K/s
    { duration: '2m', target: 10000 },    // push to 10K/s
    { duration: '1m', target: 0 },        // ramp down
  ],
  thresholds: {
    'order_latency': ['p(95)<100', 'p(99)<500'],
    'order_success': ['rate>0.99'],
    'http_req_duration': ['p(95)<200'],
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';

// Pre-generated auth tokens for load testing
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'load-test-token';

const PAIRS = ['BTC-USDT', 'ETH-USDT', 'SOL-USDT', 'AVAX-USDT', 'DOT-USDT'];
const SIDES = ['buy', 'sell'];

function randomPrice(base, variance) {
  return base + (Math.random() - 0.5) * 2 * variance;
}

function randomQuantity(min, max) {
  return min + Math.random() * (max - min);
}

function generateOrder() {
  const pair = PAIRS[Math.floor(Math.random() * PAIRS.length)];
  const side = SIDES[Math.floor(Math.random() * SIDES.length)];

  let basePrice;
  switch (pair) {
    case 'BTC-USDT': basePrice = 50000; break;
    case 'ETH-USDT': basePrice = 3500; break;
    case 'SOL-USDT': basePrice = 100; break;
    case 'AVAX-USDT': basePrice = 35; break;
    case 'DOT-USDT': basePrice = 7; break;
    default: basePrice = 100;
  }

  return {
    pair: pair,
    side: side,
    type: 'limit',
    price: randomPrice(basePrice, basePrice * 0.02),
    quantity: randomQuantity(0.01, 10.0),
  };
}

export default function () {
  const order = generateOrder();

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
  };

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/orders`, JSON.stringify(order), params);
  const latency = Date.now() - startTime;

  orderLatency.add(latency);

  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'has order id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  orderSuccess.add(success ? 1 : 0);

  sleep(0.01); // 10ms between orders per VU
}
