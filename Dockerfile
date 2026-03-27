FROM golang:latest

LABEL org.opencontainers.image.source="https://forgejoe.makiyama.dev.br/makiyama-dev/go-event-analyser"

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
#COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -v -race ./...

CMD ["go", "run", "."]
