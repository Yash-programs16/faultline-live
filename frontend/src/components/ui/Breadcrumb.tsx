import { Link } from 'react-router-dom'
import IconBase from './IconBase'

export interface Crumb {
  label: string
  to?: string
}

interface Props {
  items: Crumb[]
}

export default function Breadcrumb({ items }: Props) {
  return (
    <nav aria-label="Breadcrumb">
      <ol className="flex flex-wrap items-center gap-1.5 text-sm">
        {items.map((item, i) => {
          const last = i === items.length - 1
          return (
            <li key={`${item.label}-${i}`} className="flex items-center gap-1.5">
              {item.to && !last ? (
                <Link
                  to={item.to}
                  className="text-muted transition hover:text-brand-600 dark:hover:text-brand-400"
                >
                  {item.label}
                </Link>
              ) : (
                <span
                  aria-current={last ? 'page' : undefined}
                  className={last ? 'font-medium text-heading' : 'text-muted'}
                >
                  {item.label}
                </span>
              )}
              {!last && <IconBase name="chevron-right" size={14} className="text-gray-400" />}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
