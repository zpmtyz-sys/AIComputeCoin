import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const querySuccess = new Rate('query_success');
const queryLatency = new Trend('query_latency');

export const options = {
  stages: [
    { duration: '30s', target: 500 },     // ramp up
    { duration: '2m', target: 5000 },      // moderate load
    { duration: '3m', target: 50000 },     // push to 50K/s
    { duration: '1m', target: 0 },         // ramp down
  ],
  thresholds: {
    'query_latency': ['p(95)<50', 'p(99)<100'],
    'query_success': ['rate>0.999'],
    'http_req_duration': ['p(95)<50'],
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'load-test-token';

const PAIRS = ['BTC-USDT', 'ETH-USDT', 'SOL-USDT', 'AVAX-USDT', 'DOT-USDT'];
const DEPTHS = [10, 20, 50, 100];

export default function () {
  const pair = PAIRS[Math.floor(Math.random() * PAIRS.length)];
  const depth = DEPTHS[Math.floor(Math.random() * DEPTHS.length)];

  const params = {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
  };

  const startTime = Date.now();
  const res = http.get(`${BASE_URL}/markets/${pair}/orderbook?depth=${depth}`, params);
  const latency = Date.now() - startTime;

  queryLatency.add(latency);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'has bids': (r) => {
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body.bids);
      } catch (e) {
        return false;
      }
    },
    'has asks': (r) => {
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body.asks);
      } catch (e) {
        return false;
      }
    },
    'response time < 50ms': (r) => r.timings.duration < 50,
  });

  querySuccess.add(success ? 1 : 0);

  sleep(0.001); // 1ms between queries per VU
}
