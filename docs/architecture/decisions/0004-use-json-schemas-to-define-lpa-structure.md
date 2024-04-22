# 4. Use JSON schemas to define LPA structure

Date: 2024-04-22

## Status

Accepted

## Context

When a donor signs their LPA online, or the OPG receives it on paper, it is stored in the LPA Store as a golden copy of the record. As well as storing the data that the donor submitted, we also need to store the context in which is was executed: the agreed-upon conditions and the meaning of their choices.

This context needs to live with the LPA for its entire lifetime, which is likely to be decades, even if the context of new LPAs changes in the future (for example, if OPG updates the wording of the confirmation statement).

## Decision

To ensure we capture and preserve the context of LPAs, each LPA will be associated with a JSON schema at point of execution (online) or ingestion (paper). That schema will document the shape of the LPA's data.

The schema file will be mapped to translation files that contain any important text associated with the LPA, such as the terms the donor agreed to. There will be a file for each language a donor can use (currently English or Welsh).

The schema file, and translation files, will live in the opg-data-lpa-store repository.

The schemas will be hosted on [the OPG's data dictionary](https://data-dictionary.opg.service.justice.gov.uk/) website to ensure they have a fully resolvable URI under our control.

As the schema changes, it will be versioned by year and month (YYYY-MM). If multiple releases are made within the same month, an increasing integer will be appended to the end (YYYY-MM-N) to differentiate further.

The JSON schema associated with the LPA will be stored inside the LPA (JSON) document itself, in a `$schema` field. It will point to the resolvable URI of the JSON schema noted above.

## Consequences

As validation is currently done manually, there is a risk that the structure of the LPA will not match the JSON schema. We should introduce schema validation when storing an LPA to ensure this is not the case.

We need to ensure that any changes to the structure of the data are properly updated in the schema. An exception to this is before the start of private beta, since we will only be working with test data and can more easily change the schema.

Services that display an LPA may need to pull content from the translation files to ensure they are presenting accurate information.
