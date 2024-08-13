import { get as jsonGet, set as jsonSet } from "./jsonpointer.mjs";
import { Tabs as GovukTabs } from "../govuk-frontend.min.js";

const { Draft07 } = window.jlib;

const toBase64 = (file) =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result.split(";base64,")[1]);
    reader.onerror = reject;
  });

export class JsonSchemaEditor {
  /**
   * @type {HTMLTextAreaElement}
   */
  $module;

  /**
   * @type {HTMLDivElement}
   */
  $formContainer;

  alphabet = "346789QWERTYUPADFGHJKLXCVBNM";

  /**
   * @param {Element} $module
   */
  constructor($module) {
    if (!($module instanceof HTMLTextAreaElement)) {
      console.error(
        "Can only initialise JsonSchemaEditor on textarea elements"
      );
      return;
    }

    this.$module = $module;

    this.init();
  }

  async init() {
    const url = this.$module.getAttribute("data-module-json-schema-editor-url");

    if (!url) {
      console.error("Attribute data-module-json-schema-editor-url is missing");
      return;
    }

    const response = await fetch(url);
    this.schema = await response.json();

    const $container = document.createElement("div");
    this.$formContainer = $container;

    const $parent = this.$module.parentNode;

    const $tabs = this.addTabs([
      { label: "Visual editor", contents: $container },
      { label: "JSON", contents: this.$module },
    ]);

    $parent?.appendChild($tabs);

    this.$module.addEventListener("input", (e) => {
      this.build();
    });

    $container.addEventListener("input", (e) => {
      if (
        (e.target instanceof HTMLInputElement ||
          e.target instanceof HTMLSelectElement) &&
        e.target.name
      ) {
        const value = JSON.parse(this.$module.value);
        jsonSet(value, e.target.name, e.target.value);
        this.$module.value = JSON.stringify(value);

        if (e.target.name === "/channel") this.build();
      }
    });

    this.build();
  }

  build() {
    const jsonSchema = new Draft07(this.schema, {
      templateDefaultOptions: {
        addOptionalProps: true,
        extendDefaults: false,
      },
    });

    const value = JSON.parse(this.$module.value);

    const data = jsonSchema.getTemplate(value);

    this.$formContainer.innerHTML = "";
    this.constructElements(this.$formContainer, data, jsonSchema);

    /** @type {NodeListOf<HTMLInputElement>} */
    const $inputs = this.$formContainer.querySelectorAll("input,select");
    $inputs.forEach(($input) => {
      $input.value = jsonGet(value, $input.getAttribute("name")) ?? "";
    });
  }

  addToArray(pointer) {
    const value = JSON.parse(this.$module.value);
    const arr = jsonGet(value, pointer) ?? [];
    jsonSet(value, pointer, [...arr, {}]);
    this.$module.value = JSON.stringify(value);

    this.build();
  }

  /**
   * @param {HTMLElement} $input
   */
  createGovukFormGroup(label, $input) {
    const $div = document.createElement("div");
    $div.classList.add("govuk-form-group");

    const $label = document.createElement("label");
    $label.classList.add("govuk-label");
    $label.setAttribute("for", $input.id);
    $label.innerHTML = label;
    $div.appendChild($label);

    const $error = document.querySelector(`[href="#${$input.id}"]`);
    if ($error) {
      $div.classList.add("govuk-form-group--error");
      $input.classList.add("govuk-input--error");

      const $innerError = document.createElement("p");
      $innerError.classList.add("govuk-error-message");
      $innerError.innerHTML = `<span class="govuk-visually-hidden">Error:</span> ${$error.innerHTML.replace(
        `${$input.getAttribute("name")}: `,
        ""
      )}`;
      $div.appendChild($innerError);
    }

    $div.appendChild($input);

    return $div;
  }

  /**
   * @param {HTMLElement} $container
   * @param {object} data
   * @param {Draft07} schema
   */
  constructElements($container, data, schema) {
    const parents = {
      "": $container,
    };

    schema.each(data, (schema, _, pointer) => {
      pointer = pointer.substring(1);
      if (pointer === "") return true;

      let $parent = $container;
      const pointerParts = pointer.split("/");
      const nub = pointerParts.slice().pop();
      for (let i = 0; i < pointerParts.length; i++) {
        const parentSpace = pointerParts.slice(0, i).join("/");
        if (typeof parents[parentSpace] !== "undefined") {
          $parent = parents[parentSpace];
        }
      }

      if (schema.type === "string" && schema.enum) {
        const $select = document.createElement("select");
        $select.id = `f-${pointer}`;
        $select.name = pointer;
        $select.classList.add("govuk-select");

        ["", ...schema.enum].forEach((option) => {
          const $option = document.createElement("option");
          $select.appendChild($option);
          $option.innerHTML = option;
        });

        $parent.appendChild(this.createGovukFormGroup(nub, $select));

        $select.dispatchEvent(new InputEvent("input"));

        requestAnimationFrame(() => {
          $select.value = "";
        });
      } else if (schema.type === "string") {
        const $input = document.createElement("input");
        $input.id = `f-${pointer}`;
        $input.name = pointer;
        $input.type = schema.format === "date" ? "date" : "text";
        $input.classList.add("govuk-input");

        $parent.appendChild(this.createGovukFormGroup(nub, $input));
      } else if (
        schema.type === "object" &&
        Object.keys(schema.properties).join(",") === "filename,data"
      ) {
        const $filename = document.createElement("input");
        $filename.type = "hidden";
        $filename.name = `${pointer}/filename`;
        const $data = document.createElement("input");
        $data.type = "hidden";
        $data.name = `${pointer}/data`;

        const $input = document.createElement("input");
        $input.id = `f-${pointer}`;
        $input.type = "file";
        $input.classList.add("govuk-file-upload");

        $input.addEventListener("change", async () => {
          const file = $input.files?.[0];
          if (file) {
            $filename.value = file.name;
            $filename.dispatchEvent(new InputEvent("input", { bubbles: true }));

            $data.value = await toBase64(file);
            $data.dispatchEvent(new InputEvent("input", { bubbles: true }));
          }
        });

        $parent.appendChild($filename);
        $parent.appendChild($data);
        $parent.appendChild(this.createGovukFormGroup("Upload file", $input));

        parents[pointer] = document.createElement("div");
      } else if (schema.type === "object" || schema.type === "array") {
        const $details = document.createElement("details");
        $details.classList.add("govuk-details");
        $details.open = true;

        const $summary = document.createElement("summary");
        $summary.classList.add("govuk-details__summary");
        $summary.innerHTML = `<span class="govuk-details__summary-text">${nub}</span>`;
        $details.appendChild($summary);

        $parent.appendChild($details);

        const $detailsContent = document.createElement("div");
        $detailsContent.classList.add("govuk-details__text");
        $details.appendChild($detailsContent);

        if (schema.type === "array") {
          const $button = document.createElement("button");
          $button.classList.add("govuk-button", "govuk-button--secondary");
          $button.type = "button";
          $button.innerText = `Add`;
          $button.addEventListener("click", () => {
            this.addToArray(pointer);
          });
          $detailsContent.appendChild($button);
        }

        parents[pointer] = $detailsContent;
      } else {
        const $div = document.createElement("div");
        $div.innerHTML = "cannot " + schema.type;
        $parent.appendChild($div);
      }
    });
  }

  addTabs(tabs, selected = 0) {
    const $container = document.createElement("div");
    $container.classList.add("govuk-tabs");

    const $list = document.createElement("ul");
    $list.classList.add("govuk-tabs__list");
    $container.appendChild($list);

    tabs.forEach(({ label, contents }, index) => {
      const $tab = document.createElement("li");
      $tab.classList.add("govuk-tabs__list-item");
      $list.appendChild($tab);

      const $a = document.createElement("a");
      $a.classList.add("govuk-tabs__tab");
      $a.setAttribute("href", "#tab-" + index);
      $a.innerText = label;
      $tab.appendChild($a);

      const $panel = document.createElement("div");
      $panel.setAttribute("id", "tab-" + index);
      $panel.classList.add("govuk-tabs__panel");
      $panel.appendChild(contents);

      $container.appendChild($panel);

      if (index === selected) {
        $tab.classList.add("govuk-tabs__list-item--selected");
      } else {
        $panel.classList.add("govuk-tabs__panel--hidden");
      }
    });

    new GovukTabs($container);

    return $container;
  }
}
