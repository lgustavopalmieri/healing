import { GrpcClient } from '../../../../commom/grpc-client.js';
import { validateSpecialistResponse } from './validations.js';
import { generateSpecialistData } from './generate-data.js';
import { stressTestConfig } from './config.js';
import { stressTestRunner } from '../../../../commom/stress-runner.js';

export const options = stressTestConfig;

const grpcClient = new GrpcClient();

export default function runStressTest() {
  
  const requestData = generateSpecialistData();
  
  stressTestRunner(
    grpcClient,
    requestData,
    'pb.SpecialistService/CreateSpecialist',
    validateSpecialistResponse
  );
}
