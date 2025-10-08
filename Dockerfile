FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY  ./ ./

RUN go clean --modcache && go build -ldflags "-w" -mod=readonly -o /bin main.go

FROM alpine
RUN apk update && apk add wget

WORKDIR /
COPY --from=builder  /bin /bin
COPY /migrations /migrations
COPY /conf /conf

EXPOSE 8080

CMD ["bin/main"]