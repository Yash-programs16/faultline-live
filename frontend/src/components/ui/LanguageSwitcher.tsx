import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LANGUAGES } from '@/i18n'
import IconBase from './IconBase'

export default function LanguageSwitcher() {
  const { i18n } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement | null>(null)

  const current = LANGUAGES.find((l) => l.code === i18n.language) ?? LANGUAGES[0]

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('click', onClick)
    return () => document.removeEventListener('click', onClick)
  }, [])

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        className="flex h-10 w-10 items-center justify-center rounded-full border border-themed text-base hover:bg-gray-50 dark:hover:bg-gray-800"
        onClick={() => setOpen((s) => !s)}
        aria-label="Language"
        title={current.label}
      >
        <span aria-hidden>{current.flag}</span>
      </button>
      {open && (
        <div className="surface-card absolute right-0 mt-3 w-44 rounded-xl border py-1 shadow-theme-lg">
          {LANGUAGES.map((lang) => (
            <button
              key={lang.code}
              type="button"
              onClick={() => {
                i18n.changeLanguage(lang.code)
                setOpen(false)
              }}
              className={`flex w-full items-center gap-2 px-4 py-2 text-sm hover:bg-gray-50 dark:hover:bg-gray-800 ${
                lang.code === current.code ? 'font-medium text-brand-600' : 'text-body'
              }`}
            >
              <span aria-hidden>{lang.flag}</span>
              {lang.label}
              {lang.code === current.code && <IconBase name="check" size={14} className="ml-auto" />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
