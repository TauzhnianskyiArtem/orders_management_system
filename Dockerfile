FROM golang:1.21-alpine3.19 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/orders_management_system ./cmd/orders_management_system


FROM scratch AS final

WORKDIR /

COPY --from=build /bin/orders_management_system /orders_management_system

EXPOSE 8080
EXPOSE 8082

ENTRYPOINT ["/orders_management_system"]