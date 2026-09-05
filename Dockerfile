# Framework CLI image (not a demo app server).
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /amarra-cais ./cmd/amarra-cais

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1000 cais
COPY --from=build /amarra-cais /usr/local/bin/amarra-cais
USER cais
ENTRYPOINT ["amarra-cais"]
