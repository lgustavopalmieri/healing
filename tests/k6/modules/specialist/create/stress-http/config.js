export const stressTestConfig = {
  stages: [
    { duration: '15s', target: 10 },
    { duration: '30s', target: 30 },
    { duration: '30s', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '15s', target: 0 },
  ],

  thresholds: {
    'http_req_duration': [
      'p(50)<300',
      'p(90)<800',
      'p(95)<1200',
      'p(99)<2500',
    ],

    'checks': [
      'rate>0.95',
    ],

    'iteration_duration': [
      'p(95)<3000',
    ],
  },

  setupTimeout: '30s',
  teardownTimeout: '30s',

  noConnectionReuse: false,
  noVUConnectionReuse: false,

  tags: {
    test_type: 'stress',
    service: 'specialist_creation_http',
    environment: 'development',
    max_vus: '50',
  },
};
