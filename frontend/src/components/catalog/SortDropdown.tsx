import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { SortKey } from '@/data/types'
import IconBase from '@/components/ui/IconBase'

interface Props {
  value: SortKey
  onChange: (value: SortKey) => void
}

const OPTIONS: { key: SortKey; label: string }[] = [
  { key: 'relevance', label: 'catalog.sort.relevance' },
  { key: 'price-asc', label: 'catalog.sort.priceAsc' },
  { key: 'price-desc', label: 'catalog.sort.priceDesc' },
  { key: 'newest', label: 'catalog.sort.newest' },
  { key: 'rating', label: 'catalog.sort.rating' },
]

export default function SortDropdown({ value, onChange }: Props) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('click', onClick)
    return () => document.removeEventListener('click', onClick)
  }, [])

  const current = OPTIONS.find((o) => o.key === value) ?? OPTIONS[0]

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((s) => !s)}
        className="flex h-10 items-center gap-2 rounded-lg border border-themed px-3.5 text-sm font-medium text-body transition hover:text-heading"
      >
        <span className="text-muted">{t('catalog.sortBy')}:</span>
        <span className="text-heading">{t(current.label)}</span>
        <IconBase name="chevron-down" size={16} className="text-muted" />
      </button>
      {open && (
        <div className="surface-card absolute right-0 z-20 mt-2 w-52 rounded-xl border py-1 shadow-theme-lg">
          {OPTIONS.map((o) => (
            <button
              key={o.key}
              type="button"
              onClick={() => {
                onChange(o.key)
                setOpen(false)
              }}
              className={`flex w-full items-center justify-between px-4 py-2 text-sm hover:bg-gray-50 dark:hover:bg-gray-800 ${
                o.key === value ? 'font-medium text-brand-600' : 'text-body'
              }`}
            >
              {t(o.label)}
              {o.key === value && <IconBase name="check" size={14} />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
