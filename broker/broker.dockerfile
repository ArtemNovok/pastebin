FROM golang:1.20-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY .   /app

RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api 

RUN chmod +x /app/brokerApp

FROM alpine

RUN mkdir /app

WORKDIR  /app

COPY --from=build /app/brokerApp  /app

COPY /cmd/api/templates /app 

CMD ["/app/brokerApp"] 

