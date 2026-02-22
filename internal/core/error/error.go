package errors

import "errors"

var (
	// Global
	ErrUnknown        = errors.New("ErrUnknown")        // Erro desconhecido
	ErrMissingToken   = errors.New("ErrMissingToken")   // Access Token faltando
	ErrInvalidToken   = errors.New("ErrInvalidToken")   // Access Token inválido
	ErrInvalidRequest = errors.New("ErrInvalidRequest") // Requisição inválida

	// Auth
	ErrEmailAlreadyUsed   = errors.New("ErrEmailAlreadyUsed")   // Email jé em uso
	ErrUserNotFound       = errors.New("ErrUserNotFound")       // Usuário não encontrado
	ErrInvalidCredentials = errors.New("ErrInvalidCredentials") // Credenciais inválidas

	// ViaCEP
	ErrInvalidCEP              = errors.New("ErrInvalidCEP")              // CEP com formato inválido
	ErrCEPNotFound             = errors.New("ErrCEPNotFound")             // CEP não encontrado na base do ViaCEP
	ErrExternalServiceFailure  = errors.New("ErrExternalServiceFailure")  // Falha ao se comunicar com serviço externo
	ErrInvalidExternalResponse = errors.New("ErrInvalidExternalResponse") // Resposta do serviço externo não pôde ser interpretada
	ErrExternalBadRequest      = errors.New("ErrExternalBadRequest")      // Serviço externo rejeitou a requisição (400)

	// PropertyAds
	ErrInvalidAdType       = errors.New("ErrInvalidAdType")       // Tipo inválido (não SALE/RENT)
	ErrInvalidPrice        = errors.New("ErrInvalidPrice")        // Preço inválido (deve ser > 0)
	ErrMissingAddressField = errors.New("ErrMissingAddressField") // Campo de endereço obrigatório ausente
	ErrInvalidImageType    = errors.New("ErrInvalidImageType")    // Tipo de imagem inválido (só JPEG/PNG)
	ErrImageTooLarge       = errors.New("ErrImageTooLarge")       // Imagem excede 5MB
	ErrPropertyAdNotFound  = errors.New("ErrPropertyAdNotFound")  // Anúncio não encontrado ou já deletado

	// ExchangeRates
	ErrInvalidRate = errors.New("ErrInvalidRate") // Taxa de câmbio inválida (deve ser > 0)
)
