import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

// Configuração do teste de stress
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up para 10 usuários em 30s
    { duration: '1m', target: 10 },   // Mantém 10 usuários por 1 minuto
    { duration: '30s', target: 50 },  // Ramp up para 50 usuários em 30s
    { duration: '1m', target: 50 },   // Mantém 50 usuários por 1 minuto
    { duration: '30s', target: 0 },   // Ramp down para 0
  ],
  thresholds: {
    'grpc_req_duration': ['p(95)<500'], // 95% das requisições devem ser < 500ms
    'checks': ['rate>0.95'],             // 95% das verificações devem passar
  },
};

const client = new grpc.Client();
client.load(['./proto'], 'specialist.proto');

export default function () {
  // Conecta ao serviço gRPC
  client.connect('healing-specialist:50051', {
    plaintext: true,
  });

  // Gera dados aleatórios para cada requisição
  const randomId = Math.floor(Math.random() * 1000000);
  const request = {
    name: `Dr. Test User ${randomId}`,
    email: `test.user.${randomId}@example.com`,
    phone: '+5511999999999',
    specialty: 'Cardiology',
    license_number: `CRM-SP-${randomId}`,
    description: 'Experienced cardiologist for stress testing',
    keywords: ['cardiology', 'heart', 'stress-test'],
    agreed_to_share: true,
  };

  // Faz a chamada gRPC
  const response = client.invoke('pb.SpecialistService/CreateSpecialist', request);

  // Verifica a resposta
  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
    'specialist created': (r) => r && r.message && r.message.specialist,
    'has valid ID': (r) => r && r.message && r.message.specialist && r.message.specialist.id,
    'email matches': (r) => r && r.message && r.message.specialist && r.message.specialist.email === request.email,
  });

  // Log de erro se houver
  if (response.status !== grpc.StatusOK) {
    console.error(`Error: ${response.status} - ${response.error ? response.error.message : 'Unknown error'}`);
  }

  client.close();
  
  // Pequeno delay entre requisições
  sleep(1);
}
