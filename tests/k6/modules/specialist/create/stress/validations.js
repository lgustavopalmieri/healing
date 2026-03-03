import { check } from 'k6';
import grpc from 'k6/net/grpc';

export function validateSpecialistResponse(response, request = null) {
  const validations = {
    'grpc status ok': (r) => r && r.status === grpc.StatusOK,
    'specialist exists': (r) => r && r.message && r.message.specialist,
    'id present': (r) => r && r.message && r.message.specialist && r.message.specialist.id,
    'name present': (r) => r && r.message && r.message.specialist && r.message.specialist.name,
    'email present': (r) => r && r.message && r.message.specialist && r.message.specialist.email,
    'phone present': (r) => r && r.message && r.message.specialist && r.message.specialist.phone,
    'specialty present': (r) => r && r.message && r.message.specialist && r.message.specialist.specialty,
    'license present': (r) => r && r.message && r.message.specialist && r.message.specialist.licenseNumber,
    'description present': (r) => r && r.message && r.message.specialist && r.message.specialist.description,
    'keywords array': (r) => r && r.message && r.message.specialist && Array.isArray(r.message.specialist.keywords),
    'agreement boolean': (r) => r && r.message && r.message.specialist && typeof r.message.specialist.agreedToShare === 'boolean',
    'status present': (r) => r && r.message && r.message.specialist && r.message.specialist.status,
  };
  return check(response, validations);
}
