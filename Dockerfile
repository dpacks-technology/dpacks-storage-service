FROM golang:1.21-alpine

COPY . /app
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go get

COPY . ./

CMD ["go", "run", "main.go"]