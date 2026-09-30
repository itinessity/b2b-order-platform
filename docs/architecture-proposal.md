# Architecture Proposal

## Understanding and assumptions
The platform consumes at-least-once order events, enriches them through independently deployed providers, applies deterministic eligibility and money rules, persists one traceable result, and emits one logical outcome. `eventId` identifies the fact and `(orderId,eventVersion)` identifies a revision. Prices belong to the incoming order; product status and tax category belong to Products API.

The primary product goal is not merely to move a message between systems. It is to produce a commercially correct, explainable, and auditable order decision. Infrastructure choices are therefore evaluated by how well they protect business invariants: one logical effect per order revision, deterministic monetary results, explicit rejection reasons, and a recoverable processing trail.

## Business constraints

### Supported commercial scope

- The first release supports only `MX`, `CO`, and `PE`.
- Currency is determined by market and cannot be selected independently: `MXN` for MX, `COP` for CO, and `PEN` for PE.
- An order must contain at least one item, each product may occur only once, quantity must be a positive integer, and unit price cannot be negative.
- An invalid input event is not stored as a processed order because it never entered the valid business workflow. It is diagnostic evidence and belongs in the invalid-message/DLT path.

### Eligibility

An order is eligible for approval only when the client exists and is `ACTIVE`, the client's market matches the order market, and every referenced product exists and is `ACTIVE`. A failed eligibility rule produces `REJECTED`, not `TECHNICAL_FAILURE`, because the platform reached a valid and final business decision. The rejection reason must be stable and traceable, for example `CLIENT_BLOCKED`, `CLIENT_MARKET_MISMATCH`, `PRODUCT_NOT_FOUND`, or `PRODUCT_INACTIVE`.

### Pricing, discount, and taxation

- The incoming event owns the unit price used for this order. Products API enriches eligibility and tax classification but does not silently replace the commercial price.
- A `WHOLESALE` client receives a 3% line discount only when buying at least 20 units of that product.
- Tax is calculated after discount. The tax matrix is market-specific: MX uses 16%/8%, CO 19%/5%, and PE 18%/10% for `STANDARD`/`REDUCED`; `EXEMPT` is 0%.
- A client with `taxRegime = EXEMPT` has a 0% effective rate regardless of product category.
- Every monetary value is calculated with `BigDecimal`, rounded to two decimals with `HALF_UP` at line level, and order totals are sums of the already rounded line values.

The order of operations is a business invariant:

```text
grossSubtotal = quantity * unitPrice
discount      = grossSubtotal * discountRate
netSubtotal   = grossSubtotal - discount
taxAmount     = netSubtotal * taxRate
lineTotal     = netSubtotal + taxAmount
```

Changing this order, rounding only at order level, or recalculating price from Products API would change customer-facing totals and must therefore be treated as a business-contract change.

### Outcome semantics

- `APPROVED` means all eligibility and calculation rules completed successfully.
- `REJECTED` means processing completed successfully but a business condition was not satisfied.
- `TECHNICAL_FAILURE` means no trustworthy business decision could be produced after the bounded retry policy was exhausted.

Only approved and rejected orders produce the normal `orders.processed.v1` business outcome. Technical failures preserve the original evidence and diagnostic metadata in `orders.processing.dlt` so that support can investigate and replay safely.

## Business-focused implementation

The implementation starts with a framework-independent domain core rather than Kafka or persistence adapters. `OrderValidator` protects input invariants, `EligibilityPolicy` produces explicit business rejection reasons, and `OrderCalculator` implements the calculation sequence, market tax matrix, exemption, wholesale discount, line-level rounding, and aggregation. These rules are unit-tested without Spring, MongoDB, Kafka, or HTTP so a commercial-rule change can be reviewed and released independently from an infrastructure change.

Provider services are intentionally small. Clients API supplies client status, segment, tax regime, and market. Products API supplies product status and tax category. Neither provider owns order calculation. This prevents three services from duplicating the same pricing and eligibility logic and establishes Order Processor as the single owner of the order decision.

External representations are mapped at the application boundary. Provider DTOs, Kafka contracts, persistence documents, and domain models must not be reused as the same object. This keeps provider contract evolution from directly modifying the business model and makes compatibility decisions explicit.

Current implementation evidence covers domain calculations, validation, eligibility, provider contracts, health checks, and endpoint tests. Kafka consumption, MongoDB atomic version enforcement, the executable outbox relay, retry orchestration, and DLT publication remain target architecture and must be completed before the full end-to-end reliability claim is made.

## Decision drivers and resulting decisions

| Input or constraint | Business/technical risk | Decision | Why this decision follows from the input |
|---|---|---|---|
| Kafka delivery is at-least-once | The same order may create multiple business effects | Use stable event identity plus unique `sourceEventId` and `(orderId,eventVersion)` constraints | Correctness must be enforced atomically at the data boundary, not by an unsafe `exists` then `save` sequence |
| Different revisions for the same order may arrive out of order | An old event may overwrite a newer commercial decision | Apply an atomic version guard and classify lower versions as stale | The latest accepted revision is the business truth and must not be replaced by delayed delivery |
| Persistence and Kafka publication cannot share one native transaction | An order may exist without an outcome event, or an event may exist without its order | Use a MongoDB transactional outbox and an at-least-once relay | Order state and publication intent become one durable decision while replay remains safe through event identity |
| Client and product lookups are remote HTTP calls | Slow or unavailable dependencies may block partitions and create ambiguous outcomes | Apply deadlines, bounded retries, backoff, and final/retryable error classification | Business absence (`404`) must not be retried, while temporary infrastructure failure must not become a false rejection |
| Monetary results are externally visible and market-specific | Floating-point or inconsistent rounding may change invoices | Use `BigDecimal`, line-level `HALF_UP`, and sum rounded lines | This directly implements the specified commercial calculation contract and produces deterministic totals |
| Client exemption overrides product taxation | Generic product tax logic can overcharge exempt clients | Resolve effective tax from both client regime and product category inside the domain | Tax is an order-level business decision requiring data from both bounded contexts |
| Products and Clients are owned independently | Provider changes may break the worker | Keep provider ownership, explicit DTO mapping, additive v1 evolution, and new versions for breaking changes | Teams can deploy independently without leaking provider internals into the order domain |
| The worker is I/O-heavy and the exercise is time-bounded | A fully reactive stack adds cognitive and operational cost | Use Java 21 virtual threads with linear orchestration | It preserves readable failure paths while supporting blocking HTTP, Kafka, and Mongo I/O |
| Provider APIs are supporting components, not the core business challenge | Over-design would consume time without improving order correctness | Use in-memory repositories behind interfaces and idiomatic framework boundaries | The contracts remain replaceable by databases later without contaminating controllers or consumers |
| Three services must be developed by a four-person team | Parallel work can diverge or block integration | Agree contracts and fixtures first, then split domain, integration, Products, and Clients ownership | Parallel delivery remains possible while end-to-end responsibility and shared invariants stay explicit |
| Production incidents must be diagnosable by order and event | A customer-reported order can disappear between components | Propagate `orderId` and `eventId`, log state transitions, and measure retries, duplicates, DLT, latency, and outbox age | Support needs a traceable path from input evidence to the final business or technical outcome |

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
