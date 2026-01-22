FROM golang:1.24
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8888
CMD ["./main"]
