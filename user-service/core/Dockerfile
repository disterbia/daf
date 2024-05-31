# 기본 이미지로 Go 최신 버전을 사용
FROM golang:1.21.5 AS builder

# 작업 디렉토리 설정
WORKDIR /msa

# 나머지 소스 코드 및 go.mod, go.sum 파일 복사
COPY . .

# 의존성 다운로드
RUN go mod download

# 크로스 컴파일링을 위한 환경 설정
ENV GOOS=linux GOARCH=amd64

# 애플리케이션 빌드
RUN go build -o email_service .

# 최종 실행 이미지
FROM alpine:latest

# tzdata 패키지 설치
RUN apk --no-cache add tzdata

# 작업 디렉토리 설정
WORKDIR /msa

# 빌더 스테이지에서 생성된 실행 파일 복사
COPY --from=builder /msa/email_service .
# .env 파일 복사 추가
COPY --from=builder /msa/.env .

# 애플리케이션 실행
CMD ["./email_service"]
