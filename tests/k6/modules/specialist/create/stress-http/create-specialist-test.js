import { HttpClient } from "../../../../commom/http-client.js";
import { validateSpecialistResponse } from "./validations.js";
import { generateSpecialistData } from "../../../../commom/factories/specialist.js";
import { stressTestConfig } from "./config.js";
import { httpStressTestRunner } from "../../../../commom/http-stress-runner.js";

export const options = stressTestConfig;

const httpClient = new HttpClient();

export default function runStressTest() {
  const requestData = generateSpecialistData();

  httpStressTestRunner(
    httpClient,
    requestData,
    "/api/v1/specialists",
    validateSpecialistResponse,
  );
}
