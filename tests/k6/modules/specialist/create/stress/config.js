// 🔥 CONFIGURAÇÃO STRESS TEST EXTREMO - CreateSpecialist 🔥
export const stressTestConfig = {
  // 🚀 CENÁRIO DE CARGA PROGRESSIVA VIOLENTA
  stages: [
    // Fase 1: Aquecimento suave (preparar conexões)
    { duration: '30s', target: 50 },   // Sobe para 50 VUs em 30s
    
    // Fase 2: Carga moderada (testar estabilidade inicial)
    { duration: '1m', target: 100 },   // Sobe para 100 VUs e mantém por 1min
    
    // Fase 3: Stress pesado (começar a pressionar)
    { duration: '1m', target: 300 },   // SALTA para 300 VUs em 1min
    
    // Fase 4: CARGA EXTREMA (teste de limite)
    { duration: '2m', target: 500 },   // 🔥 500 VUs SIMULTÂNEOS por 2min
    
    // Fase 5: PICO MÁXIMO (quebrar tudo que for fraco)
    { duration: '1m', target: 800 },   // 🚀 800 VUs - MÁXIMO STRESS
    
    // Fase 6: Sustentação do caos (manter pressão)
    { duration: '3m', target: 800 },   // Mantém 800 VUs por 3min completos
    
    // Fase 7: Descida controlada (verificar recuperação)
    { duration: '1m', target: 300 },   // Volta para 300 VUs
    { duration: '30s', target: 100 },  // Reduz para 100 VUs
    { duration: '30s', target: 0 },    // Finaliza graciosamente
  ],

  // 📊 MÉTRICAS E THRESHOLDS RIGOROSOS
  thresholds: {
    // ⚡ LATÊNCIA - Deve aguentar a pancada
    'grpc_req_duration': [
      'p(50)<100',    // 50% das requests < 100ms (mediana agressiva)
      'p(90)<300',    // 90% das requests < 300ms (boa performance)
      'p(95)<500',    // 95% das requests < 500ms (aceitável sob stress)
      'p(99)<1000',   // 99% das requests < 1s (limite máximo)
    ],

    // ✅ TAXA DE SUCESSO - Não pode falhar muito
    'checks': [
      'rate>0.98',    // 98% de sucesso mínimo (mais rigoroso que 95%)
    ],

    // 🔄 THROUGHPUT - Quantas requests por segundo
    'http_reqs': [
      'rate>100',     // Mínimo 100 requests/segundo
    ],

    // 📈 ITERAÇÕES - Quantos ciclos completos
    'iterations': [
      'rate>80',      // Mínimo 80 iterações/segundo
    ],

    // 🌐 REDE - Controle de dados transferidos
    'data_received': [
      'rate<10000',   // Máximo 10KB/s recebidos (evitar sobrecarga)
    ],
    'data_sent': [
      'rate<15000',   // Máximo 15KB/s enviados (controlar upload)
    ],

    // ⏱️ DURAÇÃO DE ITERAÇÃO - Tempo total por ciclo
    'iteration_duration': [
      'p(95)<600',    // 95% das iterações < 600ms (incluindo sleep)
    ],

    // 🎯 VUS ATIVOS - Controle de usuários virtuais
    'vus': [
      'value<=800',   // Nunca exceder 800 VUs
    ],
  },

  // 🛠️ CONFIGURAÇÕES AVANÇADAS DE EXECUÇÃO
  setupTimeout: '60s',        // Tempo para setup inicial
  teardownTimeout: '60s',     // Tempo para limpeza final
  
  // 🚫 LIMITES DE SEGURANÇA (evitar travamento)
  maxRedirects: 0,           // Sem redirects (gRPC não usa)
  batch: 20,                 // Processa 20 requests em lote
  batchPerHost: 10,          // Máximo 10 por host simultaneamente
  
  // 🔧 CONFIGURAÇÕES DE REDE OTIMIZADAS
  noConnectionReuse: false,   // Reutilizar conexões (mais eficiente)
  noVUConnectionReuse: false, // Reutilizar conexões por VU
  
  // 📝 TAGS PARA ANÁLISE DETALHADA
  tags: {
    test_type: 'extreme_stress',
    service: 'specialist_creation',
    environment: 'test',
    max_vus: '800',
  },

  // 🎛️ CONFIGURAÇÕES DE SISTEMA
  systemTags: [
    'check',
    'error',
    'error_code', 
    'expected_response',
    'group',
    'method',
    'name',
    'proto',
    'scenario',
    'status',
    'subproto',
    'tls_version',
    'url',
    'vu',
    'iter',
  ],
};