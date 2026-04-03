import { check } from 'k6';

export function validateSearchResponse(response) {
  let body = null;
  try {
    body = JSON.parse(response.body);
  } catch (e) {
    // noop
  }

  const validations = {
    'http status 200': (r) => r.status === 200,
    'body parsed': () => body !== null,
    'specialists is array': () => body && Array.isArray(body.specialists),
    'pagination present': () => body && body.pagination !== undefined,
    'has_next_page present': () => body && typeof body.pagination.has_next_page === 'boolean',
    'has_previous_page present': () => body && typeof body.pagination.has_previous_page === 'boolean',
    'total_items_in_page present': () => body && typeof body.pagination.total_items_in_page === 'number',
  };

  return check(response, validations);
}
