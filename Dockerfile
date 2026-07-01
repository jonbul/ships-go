FROM golang:1.26.2

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /app ./...

EXPOSE 3000

CMD ["/app/app"]
