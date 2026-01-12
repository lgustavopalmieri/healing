package application

// TEST CASES
// 1-validate license number
// o repository c.repository.ValidateUniqueness(ctx, id, email, licenseNumber) retorna true
// o método c.externalGateway.ValidateLicenseNumber(ctx, licenseNumber) retorna true (válida)
// o c.repository.Save(ctx, specialist) é chamado e salvo
// o restante (trace, eventos, logs e etc) são chamados com sucesso

// 2-validate license number
// o repository c.repository.ValidateUniqueness(ctx, id, email, licenseNumber) retorna true
// o método c.externalGateway.ValidateLicenseNumber(ctx, licenseNumber) retorna false (inválida)
// o c.repository.Save(ctx, specialist) NÃO é chamado nem salvo
// o restante (trace, logs e etc) são chamados com erros e mensagens corretas
// o evento não deve ser publicado
// o commando deve retornar o erro correto deste caso

// 3-validate license number
// o repository c.repository.ValidateUniqueness(ctx, id, email, licenseNumber) retorna true
// o método c.externalGateway.ValidateLicenseNumber(ctx, licenseNumber) retorna um erro
// o c.repository.Save(ctx, specialist) NÃO é chamado nem salvo
// o restante (trace, logs e etc) são chamados com erros e mensagens corretas
// o evento não deve ser publicado
// o commando deve retornar o erro correto deste caso

// 4-validate license number
// o repository c.repository.ValidateUniqueness(ctx, id, email, licenseNumber) retorna true
// o método c.externalGateway.ValidateLicenseNumber(ctx, licenseNumber) excede o período de 800ms
// o c.repository.Save(ctx, specialist) NÃO é chamado nem salvo
// o restante (trace, eventos, logs e etc) são chamados com erros e mensagens corretas
// o evento não deve ser publicado
// o commando deve retornar o erro correto deste caso
