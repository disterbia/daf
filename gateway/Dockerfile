# 기본 이미지로 Go 최신 버전을 사용
FROM golang:1.21.5 AS builder

# 작업 디렉토리 설정
WORKDIR /daf

# 나머지 소스 코드 및 go.mod, go.sum 파일 복사
COPY . .

# 의존성 다운로드
RUN go mod download

# 크로스 컴파일링을 위한 환경 설정
ENV GOOS=linux GOARCH=amd64

# 애플리케이션 빌드
RUN go build -o gateway .

# 최종 실행 이미지
FROM ubuntu:latest

# 필요한 패키지 설치
RUN apt-get update && apt-get install -y tzdata ca-certificates && update-ca-certificates

# 작업 디렉토리 설정
WORKDIR /daf

# 빌더 스테이지에서 생성된 실행 파일 복사
COPY --from=builder /daf/gateway .

# 애플리케이션 실행
CMD ["./gateway"]
