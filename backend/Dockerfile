FROM golang:1.22-alpine

WORKDIR /app

ENV GOTOOLCHAIN=local

COPY go.mod ./
COPY internal/ internal/
COPY pkg/ pkg/
COPY cmd/ cmd/
RUN go mod tidy
RUN go build -o server ./cmd/

EXPOSE 8080
CMD ["./server"]
