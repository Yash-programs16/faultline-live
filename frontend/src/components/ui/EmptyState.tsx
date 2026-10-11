import type { ReactNode } from 'react'
import IconBase from './IconBase'

interface Props {
  icon?: string
  title: string
  description?: string
  action?: ReactNode
}

export default function EmptyState({ icon = 'package', title, description, action }: Props) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-brand-50 text-brand-500 dark:bg-brand-500/15">
        <IconBase name={icon} size={28} />
      </span>
      <h3 className="mt-5 text-lg font-semibold text-heading">{title}</h3>
      {description && <p className="mt-2 max-w-sm text-sm text-muted">{description}</p>}
      {action && <div className="mt-6">{action}</div>}
    </div>
  )
}
