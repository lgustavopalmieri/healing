import http from "k6/http";

export class HttpClient {
  constructor(config = {}) {
    this.baseUrl =
      config.baseUrl ||
      "http://k8s-healingqa-60d8597078-1797570892.us-east-1.elb.amazonaws.com";
  }

  post(path, body, params = {}) {
    const url = `${this.baseUrl}${path}`;
    const payload = JSON.stringify(body);
    const headers = Object.assign(
      { "Content-Type": "application/json" },
      params.headers || {},
    );
    return http.post(url, payload, { headers });
  }
}
