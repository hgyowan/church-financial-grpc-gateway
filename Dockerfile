FROM library/golang:1.24.9-alpine AS builder
RUN apk add --no-cache git openssl

# 작업 디렉터리 설정
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

# go.mod와 go.sum 파일 복사
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

# 전체 프로젝트 소스 코드 복사
COPY . .

ENV GO111MODULE=on

ARG TOKEN_FOR_GITHUB

RUN git config --global url."https://${TOKEN_FOR_GITHUB}:@github.com/".insteadOf "https://github.com/"

# Go 빌드 실행 (main.go는 /app/cmd/grpc에 위치)
WORKDIR /app/cmd/router
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags timetzdata -a -ldflags '-w -s' -o /app/church-financial-grpc-gateway-server .

# 최종 이미지를 scratch로 설정
FROM scratch
COPY --from=builder /app/church-financial-grpc-gateway-server .
COPY --from=builder /app/swagger /swagger

EXPOSE 8080

ENTRYPOINT ["/church-financial-grpc-gateway-server"]