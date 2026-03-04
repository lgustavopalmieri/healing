export const stressTestConfig = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 300 },
    { duration: '1m', target: 500 },
    { duration: '1m', target: 800 },
    { duration: '3m', target: 800 },
    { duration: '1m', target: 0 },
  ],

  thresholds: {
    'http_req_duration': [
      'p(50)<100',
      'p(90)<300',
      'p(95)<500',
      'p(99)<1000',
    ],

    'checks': [
      'rate>0.98',
    ],

    'iteration_duration': [
      'p(95)<1100',
    ],
  },

  setupTimeout: '60s',
  teardownTimeout: '60s',

  noConnectionReuse: false,
  noVUConnectionReuse: false,

  tags: {
    test_type: 'stress',
    service: 'specialist_creation_http',
    environment: 'k8s',
    max_vus: '800',
  },
};
