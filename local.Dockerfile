FROM arisatrio03/golang:1.17-ubuntu

LABEL MAINTAINER="Ari Satrio<arisatrioputra03@gmail.com>"

WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Install library auto-reload
RUN go install github.com/cosmtrek/air@v1.42.0

COPY . ./

CMD [ "air" ]
