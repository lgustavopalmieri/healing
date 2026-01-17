import { check } from 'k6';
import grpc from 'k6/net/grpc';

export function expectedResponse(response, request = null) {
  const checks = {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
    'specialist created': (r) => r && r.message && r.message.specialist,
    'has valid ID': (r) => r && r.message && r.message.specialist && r.message.specialist.id,
    'has name': (r) => r && r.message && r.message.specialist && r.message.specialist.name,
    'has email': (r) => r && r.message && r.message.specialist && r.message.specialist.email,
    'has phone': (r) => r && r.message && r.message.specialist && r.message.specialist.phone,
    'has specialty': (r) => r && r.message && r.message.specialist && r.message.specialist.specialty,
    'has licenseNumber': (r) => r && r.message && r.message.specialist && r.message.specialist.licenseNumber,
    'has description': (r) => r && r.message && r.message.specialist && r.message.specialist.description,
    'has keywords': (r) => r && r.message && r.message.specialist && Array.isArray(r.message.specialist.keywords),
    'has agreedToShare': (r) => r && r.message && r.message.specialist && typeof r.message.specialist.agreedToShare === 'boolean',
  };

  // if (request) {
  //   if (request.name) {
  //     checks['name matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.name === request.name;
  //   }
  //   if (request.email) {
  //     checks['email matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.email === request.email;
  //   }
  //   if (request.phone) {
  //     checks['phone matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.phone === request.phone;
  //   }
  //   if (request.specialty) {
  //     checks['specialty matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.specialty === request.specialty;
  //   }
  //   if (request.license_number) {
  //     checks['licenseNumber matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.licenseNumber === request.license_number;
  //   }
  //   if (request.description) {
  //     checks['description matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.description === request.description;
  //   }
  //   if (request.keywords && Array.isArray(request.keywords)) {
  //     checks['keywords match'] = (r) => 
  //       r && r.message && r.message.specialist && 
  //       Array.isArray(r.message.specialist.keywords) &&
  //       r.message.specialist.keywords.length === request.keywords.length &&
  //       request.keywords.every(keyword => r.message.specialist.keywords.includes(keyword));
  //   }
  //   if (typeof request.agreed_to_share === 'boolean') {
  //     checks['agreedToShare matches'] = (r) => 
  //       r && r.message && r.message.specialist && r.message.specialist.agreedToShare === request.agreed_to_share;
  //   }
  // }

  return check(response, checks);
}
