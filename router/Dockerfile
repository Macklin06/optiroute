FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=https://proxy.golang.org,direct

RUN go mod download

COPY . .

RUN go build -o optiroute-router ./cmd/server

EXPOSE 8080

CMD ["./optiroute-router"]