export const stressTestConfig = {
  stages: [
    { duration: '20s', target: 25 },
    { duration: '30s', target: 50 },
    { duration: '30s', target: 75 },
    { duration: '30s', target: 100 },
    { duration: '1m30s', target: 100 },
    { duration: '15s', target: 0 },
  ],

  thresholds: {
    'http_req_duration': [
      'p(50)<500',
      'p(90)<1000',
      'p(95)<1500',
      'p(99)<3000',
    ],

    'checks': [
      'rate>0.95',
    ],

    'iteration_duration': [
      'p(95)<3500',
    ],
  },

  setupTimeout: '30s',
  teardownTimeout: '30s',

  noConnectionReuse: false,
  noVUConnectionReuse: false,

  tags: {
    test_type: 'stress',
    service: 'specialist_search_http',
    environment: 'development',
    max_vus: '100',
  },
};
