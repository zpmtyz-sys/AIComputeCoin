import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';

const connectionSuccess = new Rate('ws_connection_success');
const messagesReceived = new Counter('ws_messages_received');
const connectionLatency = new Trend('ws_connection_latency');

export const options = {
  stages: [
    { duration: '1m', target: 1000 },      // ramp up
    { duration: '3m', target: 10000 },      // moderate load
    { duration: '5m', target: 100000 },     // push to 100K concurrent
    { duration: '2m', target: 0 },          // ramp down
  ],
  thresholds: {
    'ws_connection_success': ['rate>0.95'],
    'ws_connection_latency': ['p(95)<1000'],
  },
};

const WS_URL = __ENV.WS_URL || 'ws://localhost:3000/ws';
const PAIRS = ['BTC-USDT', 'ETH-USDT', 'SOL-USDT', 'AVAX-USDT', 'DOT-USDT'];
const CHANNELS = ['orderbook', 'trades', 'kline'];

export default function () {
  const pair = PAIRS[Math.floor(Math.random() * PAIRS.length)];
  const channel = CHANNELS[Math.floor(Math.random() * CHANNELS.length)];

  const startTime = Date.now();

  const res = ws.connect(WS_URL, {}, function (socket) {
    const latency = Date.now() - startTime;
    connectionLatency.add(latency);

    socket.on('open', function () {
      connectionSuccess.add(1);

      // Subscribe to a channel
      socket.send(JSON.stringify({
        type: 'subscribe',
        channel: channel,
        pair: pair,
      }));
    });

    socket.on('message', function (msg) {
      messagesReceived.add(1);

      try {
        const data = JSON.parse(msg);
        check(data, {
          'has type field': (d) => d.type !== undefined,
          'has valid structure': (d) => d.channel !== undefined || d.type === 'connected',
        });
      } catch (e) {
        // Non-JSON message
      }
    });

    socket.on('error', function () {
      connectionSuccess.add(0);
    });

    // Keep connection alive for 30 seconds
    socket.setTimeout(function () {
      socket.send(JSON.stringify({ type: 'unsubscribe', channel: channel, pair: pair }));
      socket.close();
    }, 30000);
  });

  check(res, {
    'ws connection status is 101': (r) => r && r.status === 101,
  });

  sleep(1);
}
