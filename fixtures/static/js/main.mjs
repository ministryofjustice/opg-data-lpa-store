import { JsonSchemaEditor } from "./json-schema-editor.mjs";
import { UidGenerator } from "./uid-generator.mjs";

const initiators = {
  "json-schema-editor": JsonSchemaEditor,
  "uid-generator": UidGenerator,
};

export function initAll() {
  Object.entries(initiators).forEach(([name, Component]) => {
    const $elements = document.querySelectorAll(`[data-module="${name}"]`);

    $elements.forEach(($element) => new Component($element));
  });
}
