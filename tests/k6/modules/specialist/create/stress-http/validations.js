import { check } from 'k6';

export function validateSpecialistResponse(response, request = null) {
  let body = null;
  try {
    body = JSON.parse(response.body);
  } catch (e) {
    // noop
  }

  const specialist = body && body.specialist ? body.specialist : null;

  const validations = {
    'http status 201': (r) => r.status === 201,
    'specialist exists': () => specialist !== null,
    'id present': () => specialist && specialist.id,
    'name present': () => specialist && specialist.name,
    'email present': () => specialist && specialist.email,
    'phone present': () => specialist && specialist.phone,
    'specialty present': () => specialist && specialist.specialty,
    'license present': () => specialist && specialist.license_number,
    'description present': () => specialist && specialist.description,
    'keywords array': () => specialist && Array.isArray(specialist.keywords),
    'agreement boolean': () => specialist && typeof specialist.agreed_to_share === 'boolean',
    'status present': () => specialist && specialist.status,
  };

  return check(response, validations);
}
