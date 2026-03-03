import grpc from 'k6/net/grpc';

/**
 * GrpcClient - Reusable gRPC client for K6 tests
 */
export class GrpcClient {
  constructor(config = {}) {
    this.client = new grpc.Client();
    this.address = config.address || 'healing-specialist:50051';
    this.protoPath = config.protoPath || ['./proto'];
    this.protoFile = config.protoFile || 'specialist.proto';
    this.plaintext = config.plaintext !== undefined ? config.plaintext : true;
    this.isConnected = false;
    
    // Load the proto file
    this.client.load(this.protoPath, this.protoFile);
  }

  connect() {
    if (this.isConnected) {
      console.warn('Client already connected');
      return;
    }

    this.client.connect(this.address, {
      plaintext: this.plaintext,
    });

    this.isConnected = true;
  }

  invoke(method, request, params = {}) {
    if (!this.isConnected) {
      throw new Error('Client not connected. Call connect() first.');
    }

    return this.client.invoke(method, request, params);
  }

  close() {
    if (!this.isConnected) {
      console.warn('Client not connected');
      return;
    }

    this.client.close();
    this.isConnected = false;
  }

  getClient() {
    return this.client;
  }
}
