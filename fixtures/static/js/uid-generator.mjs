export class UidGenerator {
  /**
   * @type {HTMLInputElement}
   */
  $module;

  alphabet = "346789QWERTYUPADFGHJKLXCVBNM";

  /**
   * @param {Element} $module
   */
  constructor($module) {
    if (!($module instanceof HTMLInputElement)) {
      console.error("Can only initialise UidGenerator on input elements");
      return;
    }

    this.$module = $module;

    let $parent = $module.parentElement;
    if (
      $parent instanceof Element &&
      !$parent.classList.contains("govuk-input__wrapper")
    ) {
      const $container = document.createElement("div");
      $container.classList.add("govuk-input__wrapper");

      $parent.insertBefore($container, $module);
      $container.appendChild($module);
    }

    const $btn = document.createElement("button");
    $btn.type = "button";
    $btn.classList.add("govuk-button", "govuk-button--secondary");
    $btn.innerHTML = "🔄";
    $btn.addEventListener("click", this.generate.bind(this));

    $module.insertAdjacentElement("afterend", $btn);
  }

  generate() {
    const value = new Array(12)
      .fill(0)
      .map(
        (x) => this.alphabet[Math.floor(Math.random() * this.alphabet.length)]
      )
      .join("");

    this.$module.value =
      `M-` +
      [value.slice(0, 4), value.slice(4, 8), value.slice(8, 12)].join("-");
  }

  /**
   * @param {Element} $module
   */
  static async create($module) {
    return new UidGenerator($module);
  }
}
