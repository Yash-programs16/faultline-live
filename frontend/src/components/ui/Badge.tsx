import { type ReactNode } from 'react'
import IconBase from './IconBase'

type Color = 'brand' | 'success' | 'warning' | 'error' | 'gray'
type Variant = 'soft' | 'solid'

interface Props {
  children: ReactNode
  color?: Color
  variant?: Variant
  icon?: string
  /** Show a leading status dot */
  dot?: boolean
}

const SOFT: Record<Color, string> = {
  brand: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-400',
  success: 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500',
  warning: 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-500',
  error: 'bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-500',
  gray: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
}

const SOLID: Record<Color, string> = {
  brand: 'bg-brand-500 text-white',
  success: 'bg-success-500 text-white',
  warning: 'bg-warning-500 text-white',
  error: 'bg-error-500 text-white',
  gray: 'bg-gray-500 text-white',
}

const DOT: Record<Color, string> = {
  brand: 'bg-brand-500',
  success: 'bg-success-500',
  warning: 'bg-warning-500',
  error: 'bg-error-500',
  gray: 'bg-gray-400',
}

export default function Badge({ children, color = 'brand', variant = 'soft', icon, dot = false }: Props) {
  const styles = variant === 'solid' ? SOLID[color] : SOFT[color]
  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ${styles}`}>
      {dot && <span className={`h-1.5 w-1.5 rounded-full ${variant === 'solid' ? 'bg-white' : DOT[color]}`} />}
      {icon && <IconBase name={icon} size={12} strokeWidth={2.2} />}
      {children}
    </span>
  )
}
