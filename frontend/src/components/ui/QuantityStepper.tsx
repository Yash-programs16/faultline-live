import IconBase from './IconBase'

interface Props {
  value: number
  onChange: (value: number) => void
  min?: number
  max?: number
  size?: 'sm' | 'md'
}

export default function QuantityStepper({ value, onChange, min = 1, max = 99, size = 'md' }: Props) {
  const dim = size === 'sm' ? 'h-8 w-8' : 'h-10 w-10'
  const dec = () => onChange(Math.max(min, value - 1))
  const inc = () => onChange(Math.min(max, value + 1))

  return (
    <div className="inline-flex items-center rounded-lg border border-themed">
      <button
        type="button"
        onClick={dec}
        disabled={value <= min}
        aria-label="-"
        className={`${dim} flex items-center justify-center text-muted transition hover:text-heading disabled:opacity-40`}
      >
        <IconBase name="minus" size={16} />
      </button>
      <span className={`min-w-8 text-center text-sm font-medium text-heading ${size === 'sm' ? '' : 'min-w-10'}`}>
        {value}
      </span>
      <button
        type="button"
        onClick={inc}
        disabled={value >= max}
        aria-label="+"
        className={`${dim} flex items-center justify-center text-muted transition hover:text-heading disabled:opacity-40`}
      >
        <IconBase name="plus" size={16} />
      </button>
    </div>
  )
}
