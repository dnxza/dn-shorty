FROM golang:1.15-alpine AS builder
WORKDIR /code

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

COPY . .
RUN go get -d -v ./...
RUN go build -a -installsuffix cgo -o app .


FROM alpine:latest 
WORKDIR /root/
COPY --from=builder /code/app .
EXPOSE ${port}
CMD ["./app"]
