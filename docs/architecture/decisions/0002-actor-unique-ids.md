# 2. Actor unique IDs

Date: 2024-03-12

## Status

Accepted

## Context

We need a consistent way to refer to each actor on an LPA: donor, certificate provider, attorneys, replacement attorneys and notified people.

This comes from several needs:

- Creating a clear audit trail when an actor makes changes to an LPA (e.g. when the certificate provider signs)
- Clearly identifying actors across independent services, such as in event-driven messages
- Ensuring changes to an actor are made against the right person, particularly when there are multiple actors of the same type on one LPA

## Decision

At the point of donor execution, all actors listed on an LPA will be assigned a v4 UUID that will not change past that point. If an actor already has a UUID (because an upstream service has assigned one) then it will not be overwritten.

## Consequences

We need to ensure that we use the unique ID across services rather than keeping local IDs that won't make sense to other services.
