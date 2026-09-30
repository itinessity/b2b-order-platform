# Reliable B2B Order Platform

A compact polyglot reference implementation for processing B2B orders across Latin American markets. The repository favors explicit contracts, hexagonal boundaries, deterministic money calculations, idempotent writes, and operational clarity.

## Architecture

```text
orders.created.v1 -> order-processor (Java 21 / Spring Boot)
                         |-- HTTP -> products-api (Go)
                         |-- HTTP -> clients-api (NestJS)
                         |-- MongoDB (orders + transactional outbox)
                         `-> orders.processed.v1 / orders.processing.dlt
```

The worker owns order orchestration and calculation. Provider services own their resource contracts. MongoDB stores the order result and an outbox record in one transaction; a relay publishes the outbox asynchronously. Unique indexes on `sourceEventId` and `(orderId,eventVersion)` provide the concurrency boundary.

## Requirements

- Docker Desktop with Compose v2 (recommended)
- Or: Java 21 + Maven 3.9, Go 1.22, Node 22 + pnpm

## Run locally

```bash
docker compose up --build
```

Services: Clients `localhost:3001`, Products `localhost:8081`, Kafka `localhost:9092`, MongoDB `localhost:27017`, order-processor health `localhost:8080/actuator/health`.

Publish the sample order:

```bash
docker compose exec -T kafka kafka-console-producer --bootstrap-server kafka:29092 --topic orders.created.v1 --property parse.key=true --property key.separator=| < contracts/sample-order.txt
```

Inspect results:

```bash
docker compose exec mongodb mongosh orders --quiet --eval "db.orders.find().pretty()"
docker compose exec -T kafka kafka-console-consumer --bootstrap-server kafka:29092 --topic orders.processed.v1 --from-beginning --max-messages 1
```

## Tests

```bash
docker compose --profile test run --rm order-processor-tests
docker compose --profile test run --rm products-api-tests
docker compose --profile test run --rm clients-api-tests
```

The Java suite covers market tax matrices, exemptions, wholesale discounts, rounding, aggregation, eligibility, and input validation. Go and NestJS include service and endpoint tests. CI runs all three suites.

## Failure model

- Validation and business rejections are final and are not retried.
- HTTP `404` is a final business rejection; `429`, `500`, `502`, `503`, connection errors and timeouts are retried up to three times with exponential backoff.
- Exhausted technical failures are persisted as `TECHNICAL_FAILURE` and routed to `orders.processing.dlt` with the original payload and safe diagnostic metadata.
- Kafka offsets are committed only after the application use case returns successfully.
- Duplicate event IDs and duplicate order/version pairs are resolved by unique indexes, not `exists` followed by `save`.
- Older versions never overwrite a newer order version.

Provider APIs accept `X-Failure-Mode: transient|permanent|timeout` in non-production environments to exercise failure paths.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `KAFKA_BOOTSTRAP_SERVERS` | `kafka:29092` | Kafka brokers |
| `MONGODB_URI` | `mongodb://mongodb:27017/orders` | Order store |
| `PRODUCTS_BASE_URL` | `http://products-api:8081` | Product provider |
| `CLIENTS_BASE_URL` | `http://clients-api:3001` | Client provider |
| `HTTP_TIMEOUT` | `2s` | Provider deadline |

## Documentation

- [Architecture proposal](docs/architecture-proposal.md)
- [Implementation notes](docs/implementation-notes.md)
- [Technical leadership plan](docs/technical-leadership.md)
- [ADR 001: concurrency](docs/adr/001-concurrency-model.md)
- [ADR 002: consistency](docs/adr/002-transactional-outbox.md)

## Known limitations

The implementation intentionally omits Flutter, Schema Registry, distributed tracing, authentication, and a production outbox change-stream connector. The included polling relay is suitable for this bounded exercise; production rollout should add tracing, contract publication, secrets management, and broker-level alerting.

