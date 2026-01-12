go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

### Primeiro passo é criar o arquivo .proto

```go
syntax = "proto3";
package pb;
option go_package = "internal/modules/specialist/feature/create/infra/grpc/pb";

import "google/protobuf/timestamp.proto";

message Specialist {
    string id = 1;
    string name = 2;
    string email = 3;
    string phone = 4;
    string specialty = 5;
    string license_number = 6;
    string description = 7;
    repeated string keywords = 8;
    bool agreed_to_share = 9;
    google.protobuf.Timestamp created_at = 10;
    google.protobuf.Timestamp updated_at = 11;
}

message CreateSpecialistRequest {
    string name = 1;
    string email = 2;
    string phone = 3;
    string specialty = 4;
    string license_number = 5;
    string description = 6;
    repeated string keywords = 7;
    bool agreed_to_share = 8;
}

message CreateSpecialistResponse {
    Specialist specialist = 1;
}

service SpecialistService {
    rpc CreateSpecialist(CreateSpecialistRequest) returns (CreateSpecialistResponse) {}
}
```

### Segundo passo é gerar o código com protoc

```bash
protoc --go_out=. --go-grpc_out=. proto/specialist.proto
```

### Terceiro passo é fazer o código, desde a camada de application até o service e dependency injection.

### Quarto passo é fazer o servidor e o arquivo main.go

### Quinto passo é interagir com o servidor via **Evans - https://github.com/ktr0731/evans**

```bash
go install github.com/ktr0731/evans@latest
```

```bash
evans -r rpl

# primeiro é preciso selecionar o package
# para listar:
show package
package pb #(selecionar o nome do package)

# com o package selecionado, listamos os services registrados:
show service

# agora selecionar o service:
service SpecialistService

# agora para fazer a chamada para nosso server: 
call CreateSpecialist
```
call --dig-manually CreateSpecialist

Exemplo de payload:
```json
{
  "name": "Dr. João Silva",
  "email": "joao@exemplo.com",
  "phone": "+5511999999999",
  "specialty": "Cardiology",
  "license_number": "CRM-123456",
  "description": "Especialista em cardiologia clínica",
//   "keywords": ["heart", "cardiology"],
  "agreed_to_share": true
}
```
