FROM golang:1.24-alpine

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o citadel .

EXPOSE 8080

CMD ["./citadel", "serve"]
