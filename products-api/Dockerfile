FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /api /api
EXPOSE 8081
ENTRYPOINT ["/api"]
