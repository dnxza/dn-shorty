FROM golang:1.15 AS builder
WORKDIR /code
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM alpine:latest 
WORKDIR /root/
COPY --from=builder /code/app .
EXPOSE ${port}
CMD ["./app"]
