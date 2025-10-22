import { createSignal, For } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n, type Locale } from '../i18n'

const LanguageSwitcher: Component = () => {
  const { locale, setLocale, t } = useI18n()

  const languages: Array<{ code: Locale, name: string }> = [
    { code: 'zh', name: t('language.chinese') },
    { code: 'en', name: t('language.english') }
  ]

  return (
    <div class="dropdown dropdown-end">
      <div tabindex="0" role="button" class="btn btn-ghost btn-sm">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-4 h-4 stroke-current">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"></path>
        </svg>
        <span class="hidden sm:inline ml-1">{locale() === 'zh' ? '中文' : 'EN'}</span>
      </div>
      <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-40">
        <li class="menu-title">
          <span>{t('language.switch_language')}</span>
        </li>
        <For each={languages}>
          {(language) => (
            <li>
              <a 
                class={locale() === language.code ? 'active' : ''}
                onClick={() => setLocale(language.code)}
              >
                {language.name}
              </a>
            </li>
          )}
        </For>
      </ul>
    </div>
  )
}

export default LanguageSwitcher