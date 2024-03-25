import { UidGenerator } from "./uid-generator.mjs";

const initiators = {
  "uid-generator": UidGenerator,
};

export function initAll() {
  Object.entries(initiators).forEach(([name, Component]) => {
    const $elements = document.querySelectorAll(`[data-module="${name}"]`);

    $elements.forEach(($element) => new UidGenerator($element));
  });
}
