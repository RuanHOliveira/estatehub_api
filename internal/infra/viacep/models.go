package viacep

type ViaCEPAddress struct {
	ZipCode      string
	Street       string
	Neighborhood string
	City         string
	State        string
	Complement   string
}

type viaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	Complemento string `json:"complemento"`
	Erro        bool   `json:"erro"`
}