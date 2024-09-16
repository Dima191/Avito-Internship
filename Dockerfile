FROM golang:latest
WORKDIR /app

COPY . .

ENV POSTGRES_CONN=postgres://{username}:{password}@{host}:{port}/{db_name}
ENV SERVER_ADDRESS={host}:8080

EXPOSE 8080

COPY . .

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
RUN ./migrate -path ./migrations -database "postgres://{username}:{password}@{host}:{port}/{db_name}" up
RUN go build -o app ./cmd/app/main.go
CMD ["./app"]