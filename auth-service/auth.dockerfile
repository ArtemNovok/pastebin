FROM golang:1.20-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY .   /app

RUN CGO_ENABLED=0 go build -o authApp ./cmd/api

RUN chmod +x /app/authApp

FROM alpine

RUN mkdir /app

WORKDIR /app

COPY --from=build /app/authApp  /app

CMD [ "/app/authApp" ]


