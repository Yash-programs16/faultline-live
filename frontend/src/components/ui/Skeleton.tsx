import { type CSSProperties } from 'react'

interface Props {
  variant?: 'line' | 'circle' | 'rect'
  width?: string | number
  height?: string | number
  className?: string
}

export default function Skeleton({ variant = 'line', width, height, className = '' }: Props) {
  const base = 'animate-pulse bg-gray-200 dark:bg-gray-700'

  const shape =
    variant === 'circle'
      ? 'rounded-full'
      : variant === 'rect'
        ? 'rounded-xl'
        : 'rounded-md'

  const defaults =
    variant === 'circle'
      ? 'h-10 w-10'
      : variant === 'rect'
        ? 'h-24 w-full'
        : 'h-3.5 w-full'

  const style: CSSProperties = {}
  if (width != null) style.width = typeof width === 'number' ? `${width}px` : width
  if (height != null) style.height = typeof height === 'number' ? `${height}px` : height
  const hasDims = width != null || height != null

  return <div className={`${base} ${shape} ${hasDims ? '' : defaults} ${className}`} style={style} />
}
