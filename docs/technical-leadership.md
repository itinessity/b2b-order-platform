# Technical Leadership Plan

Four-person split: domain/orchestration, Kafka/Mongo integration, Products API, and Clients API/contracts. Owners remain responsible for end-to-end integration. The Tech Lead owns invariants, risk register, release evidence and incident command.

Agree input/output/error schemas, examples, timeouts and ownership before implementation. Providers work against fixtures while the worker uses stubs. Use short-lived branches, mandatory pull requests, a domain-aware reviewer, and a second reviewer for contract/data/concurrency changes. CI must compile all services, run unit/endpoint tests, validate fixtures and scan secrets.

Approve a decision when it protects a stated invariant with evidence and proportional operating cost. Reject it when it shifts failure elsewhere, weakens observability, or adds complexity without measurable risk reduction. For disagreement, record invariant, options, evidence, reversibility and deadline; run a small experiment for reversible choices.

Deliver contracts, providers, shadow worker, persistence/outbox and publication in that order. Compare shadow results, enable a small traffic slice and expand. Roll back the deployment without deleting persisted evidence or breaking schemas.

Definition of done: mandatory contracts executable; core/failure behavior tested; duplicate/stale policy demonstrated; dashboards, alerts and runbooks owned; DLT replay documented; no secrets; rollout and rollback rehearsed; on-call accepts the service.
