// Explicit locale picker for the web client (issue #392, Slice 2).
//
// Built on the i18n substrate in ../utils/i18n. Renders a small labelled
// control listing the supported locales and persisting the explicit choice.

import { html } from '../vendor/preact-htm.js';
import {
  LOCALE_LABELS,
  SUPPORTED_LOCALES,
  useTranslation,
  type Locale,
} from '../utils/i18n.js';

export interface LanguageOption {
  value: Locale;
  label: string;
  active: boolean;
}

/** Pure helper: build the option list for a given active locale. */
export function buildLanguageOptions(activeLocale: Locale): LanguageOption[] {
  return SUPPORTED_LOCALES.map((value) => ({
    value,
    label: LOCALE_LABELS[value],
    active: value === activeLocale,
  }));
}

/**
 * LanguageSwitcher — a labelled <select> bound to the active locale.
 * `variant="menu"` renders a menu-row layout used inside the hamburger menu.
 */
export function LanguageSwitcher({
  variant = 'inline',
  onChange,
}: {
  variant?: 'inline' | 'menu';
  onChange?: (locale: Locale) => void;
} = {}) {
  const { locale, setLocale, t } = useTranslation();
  const options = buildLanguageOptions(locale);
  const handleChange = (event: any) => {
    const next = event?.currentTarget?.value as Locale;
    setLocale(next);
    onChange?.(next);
  };

  return html`
    <div class=${`language-switcher language-switcher-${variant}`} role="none">
      <label class="language-switcher-label" for="language-switcher-select">${t('language.label')}</label>
      <select
        id="language-switcher-select"
        class="language-switcher-select"
        value=${locale}
        aria-label=${t('language.label')}
        onClick=${(event: any) => event.stopPropagation()}
        onChange=${handleChange}
      >
        ${options.map((option) => html`
          <option key=${option.value} value=${option.value}>${option.label}</option>
        `)}
      </select>
    </div>
  `;
}
