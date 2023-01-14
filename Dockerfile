FROM golang:alpine as builder

WORKDIR /go/github.com/fireacademy/golden-gate

COPY go.mod go.sum ./
COPY *.go ./
COPY redis/ ./redis
COPY grpc/ ./grpc

RUN go mod download
RUN go mod verify

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -v -ldflags="-w -s" -o /golden-gate .

FROM scratch

COPY --from=builder /golden-gate /golden-gate 

ENTRYPOINT ["/golden-gate"]
