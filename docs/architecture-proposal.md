# Architecture Proposal

## Understanding and assumptions
The platform consumes at-least-once order events, enriches them through independently deployed providers, applies deterministic eligibility and money rules, persists one traceable result, and emits one logical outcome. `eventId` identifies the fact and `(orderId,eventVersion)` identifies a revision. Prices belong to the incoming order; product status and tax category belong to Products API.

## Boundaries and flow
Order Processor owns orchestration, calculation, order state, retry policy and output events. Products owns availability/tax classification; Clients owns eligibility/tax regime. The Java domain has no Spring, Kafka, HTTP or Mongo dependency. Application ports separate provider lookup, persistence, publication and time. Adapters map external DTOs explicitly.

The worker validates, loads the client and products with deadlines, evaluates eligibility, calculates rounded lines, and persists the result. Approved and rejected outcomes are published; exhausted technical failures go to DLT. `404` is final; `429`, `500`, `502`, `503`, connection failure and timeout are retryable up to three attempts with exponential backoff.

## Idempotency and consistency
Unique indexes on `sourceEventId` and `(orderId,eventVersion)` are the concurrency boundary. A duplicate acknowledges the existing outcome. An atomic version predicate prevents an older version from replacing a newer one. Kafka is keyed by `orderId` for partition ordering, but correctness does not rely on ordering alone.

Production uses a Mongo transaction for the order and an outbox record. The relay publishes with a stable event ID, then marks it delivered. A crash may re-publish, so consumers deduplicate. Offsets are committed only after the use case records a durable decision.

## Contracts and evolution
Each provider publishes its schema. Additive fields remain in v1; breaking semantic or structural changes create a new topic/URL version. CI validates schemas, examples and consumer fixtures before rollout. Teams agree error categories and timeout budgets before coding.

## Risks and tests
Primary risks are contract drift, stale-event races, incorrect retry classification, rounding changes, slow fan-out and outbox lag. Pure unit tests protect domain matrices and invariants; endpoint tests protect provider contracts; Testcontainers should demonstrate Kafka/Mongo concurrency and partial failures; Compose is the final smoke test.

Rejected: distributed locks (database constraints are simpler), Kafka exactly-once (HTTP and Mongo are outside its transaction), and reactive Java (virtual threads keep the bounded blocking flow easier to diagnose).

Incremental delivery: freeze contracts; domain tests; providers; happy path; atomic idempotency/outbox; failure/DLT scenarios; observability; shadow deployment; limited partition rollout; full rollout with rollback to the prior consumer group.
