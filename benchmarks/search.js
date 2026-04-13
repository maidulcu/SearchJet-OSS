import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const searchLatency = new Trend('search_latency_ms');
const indexLatency = new Trend('index_latency_ms');

export const options = {
  scenarios: {
    search_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 },   // Ramp up
        { duration: '1m', target: 1000 },   // Sustain
        { duration: '30s', target: 5000 },   // Peak
        { duration: '30s', target: 0 },      // Ramp down
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<100'],  // p95 < 100ms
    http_req_duration: ['p(99)<200'],  // p99 < 200ms
    errors: ['rate<0.01'],              // Error rate < 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'masterKey';

export default function () {
  const headers = {
    'Content-Type': 'application/json',
    'X-PDPL-Consent': 'functional',
  };

  // Search queries
  const queries = [
    'مطعم دبي',
    'restaurant dubai',
    'hotel abu dhabi',
    'shopping mall sharjah',
    ' clinic near me',
  ];

  const query = queries[Math.floor(Math.random() * queries.length)];

  // Search test
  const searchRes = http.get(`${BASE_URL}/v1/search?q=${encodeURIComponent(query)}&limit=20`, {
    headers,
  });

  searchLatency.add(searchRes.timings.duration);
  check(searchRes, {
    'search status 200': (r) => r.status === 200,
    'search returns hits': (r) => r.json('hits') !== undefined,
  });
  errorRate.add(searchRes.status !== 200);

  sleep(Math.random() * 0.5 + 0.1);

  // Index test (every 10th iteration)
  if (__ITER % 10 === 0) {
    const doc = {
      documents: [
        {
          id: `doc-${Date.now()}-${__VU}-${__ITER}`,
          title: `Test Document ${Math.random()}`,
          body: 'This is a test document for benchmarking purposes.',
          lang: 'en',
          emirate: 'dubai',
          source: 'benchmark',
        },
      ],
    };

    const indexRes = http.post(`${BASE_URL}/v1/index`, JSON.stringify(doc), { headers });
    indexLatency.add(indexRes.timings.duration);
    check(indexRes, {
      'index status 202': (r) => r.status === 202,
    });
    errorRate.add(indexRes.status !== 202);
  }

  sleep(Math.random() * 1 + 0.5);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'summary.json': JSON.stringify(data),
  };
}

function textSummary(data, opts) {
  const indent = opts.indent || '';
  const enableColors = opts.enableColors || false;

  let output = `\n${indent}Test Results:\n`;
  output += `${indent}=============\n`;

  if (data.metrics.http_req_duration) {
    const duration = data.metrics.http_req_duration;
    output += `\n${indent}HTTP Request Duration:\n`;
    output += `${indent}  p95: ${duration.values['p(95)'].toFixed(2)}ms\n`;
    output += `${indent}  p99: ${duration.values['p(99)'].toFixed(2)}ms\n`;
    output += `${indent}  avg: ${duration.values.avg.toFixed(2)}ms\n`;
  }

  if (data.metrics.errors) {
    output += `\n${indent}Error Rate: ${(data.metrics.errors.values.rate * 100).toFixed(2)}%\n`;
  }

  return output;
}
