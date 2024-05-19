FROM golang:latest

WORKDIR /app

COPY go.mod go.sum keno.go ./
COPY templates/ templates/

RUN go build -o keno .

EXPOSE 5000

CMD ["./keno"]
