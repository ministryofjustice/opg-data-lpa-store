# 3. Cross-service authorisation

Date: 2024-03-12

## Status

Accepted

## Context

When an LPA is created or updated in the LPA Store, we need to keep a record of who the initiator of the action is. This ensures we have an audit trail of all changes, whether they come from a member of the public, a member of staff at the OPG or an internal service.

## Decision

Requests to create or update an LPA will include a URN identifier to point to the source of the action. This will be in the format: `urn:opg:${service}:users:${identifier}`.

For **members of the public using Make and Register**, the `service` will be "poas:mrlpa" and the `identifier` will be the actor's unique ID as stored on the LPA document (see ADR 2).

For **OPG members of staff using Sirius**, the `service` will be "sirius" and the `identifier` will be the user's (numeric) user ID in Sirius.

## Consequences

As more services integrate with the LPA Store (e.g. scanning), we will need to expand this ADR to document what service and identifier(s) they will use.
