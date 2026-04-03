export const stressTestConfig = {
  stages: [
    { duration: '20s', target: 100 },
    { duration: '30s', target: 200 },
    { duration: '30s', target: 300 },
    { duration: '30s', target: 400 },
    { duration: '2m30s', target: 400 },
    { duration: '20s', target: 0 },
  ],

  thresholds: {
    'http_req_duration': [
      'p(50)<200',
      'p(90)<500',
      'p(95)<1000',
      'p(99)<2000',
    ],

    'checks': [
      'rate>0.95',
    ],

    'iteration_duration': [
      'p(95)<2500',
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
    max_vus: '400',
  },
};
