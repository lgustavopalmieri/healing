import grpc from 'k6/net/grpc';
import { sleep } from 'k6';
import { GrpcClient } from '../../../../commom/grpc-client.js';
import { expectedResponse } from './checks.js';
import { requestData } from './request-data.js';
import { stressOptions } from './options.js';

export const options = stressOptions;

const client = new GrpcClient();

export default function () {
  client.connect();

  const request = requestData();
  console.log(`Creating specialist: ${request.email}`);

  const response = client.invoke('pb.SpecialistService/CreateSpecialist', request);

  expectedResponse(response, request);

  if (response.status === grpc.StatusOK) {
    console.log(`✓ Successfully created specialist: ${response.message.specialist.id}`);
  } else {
    console.error(`✗ Error: ${response.status} - ${response.error ? response.error.message : 'Unknown error'}`);
  }

  client.close();
  sleep(0.5);
}
