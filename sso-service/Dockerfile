FROM golang:1.25.5-alpine AS builder    

WORKDIR /sso-coderun
COPY go.mod go.sum ./
RUN go mod download 

COPY . .
RUN go build -o sso ./cmd/main.go

# ---
FROM alpine:latest

COPY --from=builder ./sso-coderun/sso .
COPY --from=builder ./sso-coderun/.env .
COPY --from=builder ./sso-coderun/configs/ ./configs
COPY --from=builder ./sso-coderun/*.key .
COPY --from=builder ./sso-coderun/*.key.pub .

EXPOSE 50051

CMD [ "./sso" ]