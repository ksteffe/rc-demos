FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/resource-demo .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/resource-demo /resource-demo
EXPOSE 8080
ENTRYPOINT ["/resource-demo"]

