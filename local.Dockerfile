FROM golang:1.22.3

LABEL MAINTAINER="Ari Satrio<arisatrioputra03@gmail.com>"

WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN go install github.com/cosmtrek/air@v1.52.0

COPY . ./

RUN go build -o main .

CMD ["air"]
