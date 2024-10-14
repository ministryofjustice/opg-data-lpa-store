# 2. Include schema in stored data

Date: 2023-09-11

## Status

Proposed

## Context

The data in this repository must be long lasting: an LPA registered in 2023 will still need to be valid for many years from now. We will also continue to add data over many years. During this time it is likely that we will change the shape of the data.

For this to be possible, we will either need to migrate data over time or be able to handle it in its original form. Either way, it is essential that each document in the store is also clear about its shape.

## Decision

- Each document stored in the service will include a unique identifier which defines its schema
- The identifier will be a URI indicating a JSONSchema document defining the full schema definition
- The identifier will be stored in a `$schema` property on the root of the document

## Consequences

This will give every data point in the store a clear schema, and leave us with no ambiguous data. This means that we can safely migrate any data which, importantly, means that we can always revisit this decision.
