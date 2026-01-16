import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

// Configuração básica - começa bem simples
export const options = {
  vus: 5,              // 5 usuários virtuais
  duration: '30s',     // Duração de 30 segundos
  thresholds: {
    'grpc_req_duration': ['p(95)<1000'], // 95% das requisições < 1s
    'checks': ['rate>0.9'],               // 90% das verificações devem passar
  },
};

const client = new grpc.Client();
client.load(['./proto'], 'specialist.proto');

export default function () {
  // Conecta ao serviço gRPC (nome do container na rede Docker)
  client.connect('healing-specialist:50051', {
    plaintext: true,
  });

  // Gera dados únicos para cada requisição
  const timestamp = Date.now();
  const randomId = Math.floor(Math.random() * 100000);
  const uniqueId = `${timestamp}-${randomId}`;
  
  const request = {
    name: `Dr. Stress Test ${uniqueId}`,
    email: `stress.test.${uniqueId}@example.com`,
    phone: '+5511987654321',
    specialty: 'Cardiology',
    license_number: `CRM-SP-${uniqueId}`,
    description: 'Stress test specialist',
    keywords: ['cardiology', 'test'],
    agreed_to_share: true,
  };

  console.log(`Creating specialist: ${request.email}`);

  // Faz a chamada gRPC
  const response = client.invoke('pb.SpecialistService/CreateSpecialist', request);

  // Verifica a resposta
  const checkResult = check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
    'has specialist': (r) => r && r.message && r.message.specialist,
    'has ID': (r) => r && r.message && r.message.specialist && r.message.specialist.id !== '',
  });

  if (!checkResult) {
    console.error(`Failed request for ${request.email}`);
    console.error(`Status: ${response.status}`);
    if (response.error) {
      console.error(`Error: ${response.error.message}`);
    }
  } else {
    console.log(`✓ Successfully created specialist: ${response.message.specialist.id}`);
  }

  client.close();
  
  // Pequeno delay entre requisições
  sleep(0.5);
}
