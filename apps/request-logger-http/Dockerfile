FROM golang:1.25-alpine AS build

WORKDIR /src
COPY extensions ./extensions
COPY rc-demos/apps/request-logger-http/go.mod ./rc-demos/apps/request-logger-http/go.mod
COPY rc-demos/apps/request-logger-http/*.go ./rc-demos/apps/request-logger-http/

WORKDIR /src/rc-demos/apps/request-logger-http
RUN go mod download
RUN go build -o /out/request-logger .

FROM gcr.io/distroless/static-debian12

COPY --from=build /out/request-logger /request-logger

ENTRYPOINT ["/request-logger"]
