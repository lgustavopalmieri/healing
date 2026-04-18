import http from "k6/http";

export class HttpClient {
  constructor(config = {}) {
    this.baseUrl =
      config.baseUrl ||
      "";
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
