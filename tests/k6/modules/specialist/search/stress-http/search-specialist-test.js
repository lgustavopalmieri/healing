import { sleep } from 'k6';
import { HttpClient } from '../../../../commom/http-client.js';
import { validateSearchResponse } from './validations.js';
import { generateSearchPayload } from './factory.js';
import { stressTestConfig } from './config.js';

export const options = stressTestConfig;

const httpClient = new HttpClient();

export default function runStressTest() {
  const requestData = generateSearchPayload();

  const response = httpClient.post('/api/v1/specialists/search', requestData);

  validateSearchResponse(response);

  if (response.status !== 200) {
    console.error(`✗ Error: ${response.status} - ${response.body}`);
  }

  sleep(0.5);
}
