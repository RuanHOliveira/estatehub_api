FROM golang:1.25-alpine

RUN apk update && apk add --no-cache git

RUN go install github.com/air-verse/air@latest

RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]