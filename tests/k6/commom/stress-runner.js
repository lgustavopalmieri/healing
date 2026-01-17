import grpc from 'k6/net/grpc';
import { sleep } from 'k6';

export function stressTestRunner(grpcClient, requestData, serviceName, validateResponse) {    
    grpcClient.connect();
    
    const response = grpcClient.invoke(serviceName, requestData);
    
    validateResponse(response, requestData);
    
    if (response.status === grpc.StatusOK) {
        console.log(`✓ Successfully created: ${response.message?.specialist?.id || 'resource'}`);
    } else {
        console.error(`✗ Error: ${response.status} - ${response.error ? response.error.message : 'Unknown error'}`);
    }
    
    grpcClient.close();
    sleep(0.5);
}