import { useUiStore, type ToastType } from '@/stores/ui'
import IconBase from './IconBase'

const BORDER: Record<ToastType, string> = {
  success: 'border-success-200 dark:border-success-500/30',
  error: 'border-error-200 dark:border-error-500/30',
  warning: 'border-warning-200 dark:border-warning-500/30',
  info: 'border-blue-light-200 dark:border-blue-light-500/30',
}

const ICON_NAME: Record<ToastType, string> = {
  success: 'check',
  error: 'x',
  warning: 'bell',
  info: 'bell',
}

const ICON_COLOR: Record<ToastType, string> = {
  success: 'text-success-500',
  error: 'text-error-500',
  warning: 'text-warning-500',
  info: 'text-blue-light-500',
}

export default function ToastStack() {
  const toasts = useUiStore((s) => s.toasts)
  const dismiss = useUiStore((s) => s.dismissToast)

  return (
    <div className="pointer-events-none fixed top-4 right-4 z-[100] flex flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex w-80 items-start gap-3 rounded-lg border bg-white p-4 shadow-theme-md dark:bg-gray-dark ${BORDER[t.type]}`}
        >
          <IconBase name={ICON_NAME[t.type]} size={20} className={ICON_COLOR[t.type]} />
          <div className="flex-1">
            <p className="text-sm font-medium text-heading">{t.message}</p>
          </div>
          <button
            type="button"
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
            onClick={() => dismiss(t.id)}
          >
            <IconBase name="x" size={16} />
          </button>
        </div>
      ))}
    </div>
  )
}
