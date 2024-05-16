FROM golang:1.20-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY .  /app

RUN CGO_ENABLED=0  go build -o hasherApp ./cmd/api

RUN chmod +x /app/hasherApp

FROM alpine

RUN mkdir /app

WORKDIR /app

COPY --from=build /app/hasherApp /app 

CMD [ "/app/hasherApp" ]



