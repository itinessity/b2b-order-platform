# ADR 003: Use of AI Assistance

## Status

Accepted

## Context

The technical exercise explicitly permits the use of artificial intelligence assistants and requires the candidate to document the tools used, their purpose, the parts reviewed personally, and at least one discarded suggestion.

## Tool used

The AI assistant used during the exercise was Codex.

## Scope of use

Codex was used to support:

- source-code implementation;
- technical documentation;
- preparation and execution of tests;
- preparation of the proposed technical solution;
- repository review and review of the final results.

Before and during implementation, the analysis explicitly considered inconsistent states, race conditions, duplicate delivery, stale events, and partial failures. Particular attention was given to separation of responsibilities and architectural layers. Domain-Driven Design principles were used to keep domain rules independent from transport, persistence, and framework concerns.

## Candidate validation and review

The repository was reviewed after implementation. The candidate personally validated the key requirements, the architectural proposal, the main business rules, the test results, and the accompanying implementation conclusions.

Special attention was given to:

- the separation between domain, application, and infrastructure concerns;
- eligibility, tax, discount, rounding, and aggregation rules;
- API contracts and error responses;
- idempotency and concurrency assumptions;
- failure classification, retry boundaries, and traceability;
- the consistency strategy between persistence and event publication;
- the documented limitations and residual technical risks.

AI assistance was also used as an additional review pass over the completed solution and its documentation. Suggestions were treated as review input rather than accepted automatically. The candidate remained responsible for the final decisions and for verifying the resulting code and documentation.

## Discarded suggestion

One discarded direction was to introduce Redis-based distributed locks and a fully reactive Java processing model from the first version. This was rejected because database-enforced uniqueness and version constraints provide a simpler and more verifiable concurrency boundary for the current scope. A linear Java 21 execution model also keeps processing and incident diagnostics easier to understand. The more complex approach should only be reconsidered if production measurements demonstrate a concrete scalability or contention problem.

## Consequences

AI assistance reduced the time required for implementation, documentation, test preparation, and review. It did not replace technical ownership: every part of the submitted solution must remain explainable, defensible, and modifiable by the candidate during the technical review.

