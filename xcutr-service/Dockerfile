FROM golang:1.25.5-alpine AS builder    

WORKDIR /xcutr-coderun
COPY go.mod go.sum ./
RUN go mod download 

COPY . .
RUN go build -o xcutr ./cmd/main.go

# ---
FROM alpine:latest

COPY --from=builder ./xcutr-coderun/xcutr .
COPY --from=builder ./xcutr-coderun/.env .
COPY --from=builder ./xcutr-coderun/configs/ ./configs
COPY --from=builder ./xcutr-coderun/*.key.pub .

EXPOSE 50052

CMD [ "./xcutr" ]