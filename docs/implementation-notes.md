# Implementation Notes

The repository follows a contract-first, domain-first sequence. The Java calculation and eligibility core is framework-independent; provider services keep transport, query logic and seeded repositories separate. Failure injection makes transient, permanent and timeout behavior reproducible.

The bounded exercise documents the production Mongo transactional-outbox boundary, while the next implementation increment is the full Kafka/Mongo Testcontainers suite and production outbox relay. Other conscious debt: authentication, OpenTelemetry traces, Schema Registry, secrets management and load testing. The highest residual risks are broker retry timing, replica-set health and outbox lag, not the money rules.

The next change should demonstrate identical-event, competing same-version and stale-version messages in parallel against real Kafka and a Mongo replica set.
