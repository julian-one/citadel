FROM golang:1.26-alpine

RUN apk add --no-cache gcc musl-dev sqlite-dev tesseract-ocr tesseract-ocr-data-eng

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o citadel .

EXPOSE 8080

CMD ["./citadel", "serve"]
