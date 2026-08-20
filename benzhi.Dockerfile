FROM golang:1.22-bookworm

ENV GOPROXY=https://goproxy.cn,direct
ENV GOTOOLCHAIN=local

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ballast-watch .

EXPOSE 8080

CMD ["/ballast-watch"]