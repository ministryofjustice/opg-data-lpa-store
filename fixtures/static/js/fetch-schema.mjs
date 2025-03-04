export async function fetchSchema(url) {
  const response = await fetch(url);
  const body = await response.text();
  let schema = JSON.parse(
    body.replaceAll(
      "https://data-dictionary.opg.service.justice.gov.uk/schema/lpa/2024-10/",
      "assets/schemas/2024-10/"
    )
  );

  if (schema.allOf) {
    for await (const element of schema.allOf) {
      if (element.$ref) {
        const innerSchema = fetchSchema(element.$ref);

        schema = mergeDeep(innerSchema, schema);
      }
    }
  }

  // Remove these from the schema so that blank values are not sent to the API
  // TODO - Find a better solution
  if (schema.properties.donor) {
    delete schema.properties.donor.properties.identityCheck;
  }

  if (schema.properties.certificateProvider) {
    delete schema.properties.certificateProvider.properties.identityCheck;
  }

  return schema;
}

export function isObject(item) {
  return item && typeof item === "object" && !Array.isArray(item);
}

export function mergeDeep(target, ...sources) {
  if (!sources.length) return target;
  const source = sources.shift();

  if (isObject(target) && isObject(source)) {
    for (const key in source) {
      if (isObject(source[key])) {
        if (!target[key]) Object.assign(target, { [key]: {} });
        mergeDeep(target[key], source[key]);
      } else {
        Object.assign(target, { [key]: source[key] });
      }
    }
  }

  return mergeDeep(target, ...sources);
}
