# EstateHub API

API REST para gerenciamento de anúncios imobiliários, com autenticação JWT, conversão de preços BRL → USD, upload de imagens e integração com o serviço ViaCEP para consulta de endereços brasileiros.

---

## Tecnologias

| Tecnologia | Versão | Função |
|------------|--------|--------|
| Go | 1.25 | Linguagem principal |
| Chi | v5 | Framework HTTP |
| PostgreSQL | 15 | Banco de dados |
| pgx | v5 | Driver PostgreSQL |
| SQLC | — | Geração de código SQL type-safe |
| golang-migrate | — | Migrations de banco de dados |
| golang-jwt | v5 | Autenticação JWT HS256 |
| bcrypt | — | Hash de senhas (custo 14) |
| Docker + Docker Compose | — | Containerização e orquestração |

---

## Arquitetura

### Estrutura de pastas

```
estatehub_api/
├── cmd/
│   └── api/
│       └── main.go                         # Entry point — wire-up e servidor HTTP
├── internal/
│   ├── core/
│   │   ├── config/                         # Carregamento de variáveis de ambiente
│   │   ├── error/                          # Erros
│   │   ├── json/                           # Helpers de serialização JSON
│   │   ├── middlewares/                    # AuthMiddleware (validação JWT)
│   │   ├── security/                       # JwtService (geração e validação de tokens)
│   │   └── testutil/                       # Utilitários compartilhados de teste
│   ├── domain/
│   │   ├── auth/                           # Registro e login de usuários
│   │   ├── property_ads/                   # Anúncios imobiliários (CRUD + upload)
│   │   ├── exchange_rates/                 # Cotações de câmbio BRL → USD
│   │   └── viacep/                         # Handler de consulta de endereço por CEP
│   ├── infra/
│   │   ├── database/postgresql/
│   │   │   ├── connection.go               # Conexão com PostgreSQL via pgx
│   │   │   ├── txmanager.go                # Gerenciador de transações
│   │   │   ├── migration/                  # Arquivos SQL de migration
│   │   │   └── sqlc/
│   │   │       ├── query/                  # Queries SQL (fonte para sqlc generate)
│   │   │       └── generated/              # Código Go gerado pelo SQLC
│   │   └── viacep/                         # Client HTTP para a API ViaCEP
│   └── router/
│       └── router.go                       # Definição de rotas e middlewares
└── docker-compose.yaml
```

### Camadas por domínio

```
Handler (HTTP) → Usecase (regras de negócio) → Repository (PostgreSQL via SQLC)
```

Cada domínio (`auth`, `property_ads`, `exchange_rates`) é organizado em três arquivos:

- **handler.go** — recebe requisições HTTP, valida o input, chama o usecase e serializa a resposta
- **usecase.go** — contém as regras de negócio; define a interface e sua implementação
- **models.go** — structs de input/output para o usecase e de request/response para o handler

As operações de banco de dados são sempre executadas dentro de transações gerenciadas pelo `TxManager`.

---

## Pré-requisitos

- [Docker](https://www.docker.com/) e [Docker Compose](https://docs.docker.com/compose/) instalados

---

## Como executar

### 1. Clonar o repositório

```bash
git clone <url-do-repositorio>
cd estatehub_api
```

### 2. Subir os serviços

```bash
docker-compose up --build
```

O Docker Compose executa os seguintes serviços **em ordem**:

| Ordem | Serviço | Função |
|-------|---------|--------|
| 1 | `postgres` | Inicia o PostgreSQL e aguarda health check |
| 2 | `test` | Executa a suite de testes (`go test ./internal/...`) |
| 3 | `migrate` | Aplica as migrations no banco de dados |
| 4 | `api` | Inicia o servidor HTTP |

> O serviço `api` só sobe se `test` e `migrate` concluírem com sucesso. Isso garante que a API nunca sobe com testes falhando ou schema desatualizado.

### 3. Verificar disponibilidade

```bash
curl http://localhost:8080/v1/health
# Resposta: API Online!
```

A API estará disponível em `http://localhost:<APP_PORT>`.

---

## Executando testes

Os testes são executados automaticamente pelo Docker Compose. Para executar manualmente (sem Docker):

```bash
go test ./internal/...
```

### Organização dos testes

- **Testes unitários** — sem dependência de banco de dados, rede ou clock externo
- **Table-driven** com `t.Run` para cobertura de múltiplos cenários por função
- **Stdlib apenas** — `testing` e `net/http/httptest`
- Testes de handler e de usecase são separados dentro de cada domínio
- Mocks manuais com campos de função; sem geração automática de mocks

**Cobertura:**

| Domínio | Handler | Usecase | Subtests |
|---------|---------|---------|----------|
| `auth` | 2 funções | 2 funções | — |
| `exchange_rates` | 2 funções | 2 funções | — |
| `property_ads` | 3 funções | 3 funções | — |
| `viacep` | 1 função | — | — |
| **Total** | **8 funções** | **7 funções** | **82 subtests** |

---

## Endpoints da API

| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| `GET` | `/v1/health` | Não | Health check |
| `POST` | `/v1/auth/register` | Não | Registrar novo usuário |
| `POST` | `/v1/auth/login` | Não | Autenticar usuário |
| `GET` | `/v1/viacep/{cep}` | Sim | Consultar endereço por CEP |
| `GET` | `/v1/property-ads` | Sim | Listar anúncios imobiliários |
| `POST` | `/v1/property-ads` | Sim | Criar anúncio (`multipart/form-data`) |
| `DELETE` | `/v1/property-ads/{id}` | Sim | Deletar anúncio (soft delete) |
| `GET` | `/v1/exchange-rates` | Sim | Listar cotações de câmbio (histórico completo) |
| `POST` | `/v1/exchange-rates` | Sim | Criar nova cotação |

---

## Segurança

- **JWT HS256:** Todas as rotas protegidas exigem `Authorization: Bearer <token>`
- **bcrypt:** Senhas armazenadas com hash bcrypt custo 14
- **Validação de input:** Todos os campos são validados pelo usecase antes de qualquer operação no banco
- **Erros padronizados:** A API retorna `error_code` fixos — nunca mensagens internas, stack traces ou detalhes de implementação
- **Limite de upload:** Imagens limitadas a 5MB; apenas JPEG e PNG são aceitos (validação por magic bytes, não por extensão)
- **Transações:** Todas as operações de banco são executadas dentro de transações com rollback automático em caso de erro

---

## Observações Importantes

### Soft delete

Anúncios deletados via `DELETE /v1/property-ads/{id}` não são removidos fisicamente do banco de dados. O campo `deleted_at` é preenchido com a data/hora da exclusão. Esses registros não aparecem na listagem de anúncios.

### Conversão BRL → USD

O campo `price_usd` nos anúncios é calculado dinamicamente pelo backend usando a cotação ativa registrada em `/v1/exchange-rates`. Se não houver cotação ativa cadastrada, `price_usd` retorna `null`.

### Histórico de cotações

O endpoint `GET /v1/exchange-rates` retorna **todo o histórico**, incluindo cotações inativas (`deleted_at != null`). Ao criar uma nova cotação, todas as anteriores são automaticamente inativadas.

### Upload de imagens

As imagens são armazenadas localmente no servidor. As respostas da API incluem o campo `image_data` com a imagem codificada em base64.