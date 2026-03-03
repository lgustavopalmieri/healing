export const stressTestConfig = {
  stages: [
    { duration: '30s', target: 100 },  // Fast ramp to 100 VUs
    { duration: '1m', target: 300 },   // Push to 300 VUs
    { duration: '1m', target: 500 },   // 🔥 500 VUs
    { duration: '1m', target: 800 },   // 🚀 800 VUs - peak stress
    { duration: '3m', target: 800 },   // Sustain peak for 3min
    { duration: '1m', target: 0 },     // Ramp down
  ],

  thresholds: {
    'grpc_req_duration': [
      'p(50)<100',
      'p(90)<300',
      'p(95)<500',
      'p(99)<1000',
    ],

    'checks': [
      'rate>0.98',
    ],

    // Accounts for sleep(0.5) in the runner (~500ms baseline)
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
    service: 'specialist_creation',
    environment: 'k8s',
    max_vus: '800',
  },
};
