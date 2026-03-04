import { sleep } from 'k6';

export function httpStressTestRunner(httpClient, requestData, path, validateResponse) {
  const response = httpClient.post(path, requestData);

  validateResponse(response, requestData);

  if (response.status !== 201) {
    console.error(`✗ Error: ${response.status} - ${response.body}`);
  }

  sleep(0.5);
}
