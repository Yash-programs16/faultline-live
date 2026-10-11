import { useField } from 'formik'
import type { InputHTMLAttributes, SelectHTMLAttributes, TextareaHTMLAttributes, ReactNode } from 'react'

const baseControl =
  'focus-ring w-full rounded-lg border bg-transparent px-3.5 py-2.5 text-sm text-body placeholder:text-muted disabled:opacity-50'

function fieldClasses(hasError: boolean) {
  return `${baseControl} ${hasError ? 'border-error-500' : 'border-themed'}`
}

interface WrapProps {
  label?: string
  name: string
  hint?: string
  children: (id: string, hasError: boolean) => ReactNode
}

function FieldWrap({ label, name, hint, children }: WrapProps) {
  const [, meta] = useField(name)
  const hasError = Boolean(meta.touched && meta.error)
  const id = `f-${name}`
  return (
    <div className="space-y-1.5">
      {label && (
        <label htmlFor={id} className="block text-sm font-medium text-heading">
          {label}
        </label>
      )}
      {children(id, hasError)}
      {hasError ? (
        <p className="text-xs text-error-500">{meta.error}</p>
      ) : hint ? (
        <p className="text-xs text-muted">{hint}</p>
      ) : null}
    </div>
  )
}

type TextFieldProps = { label?: string; name: string; hint?: string } & InputHTMLAttributes<HTMLInputElement>

export function TextField({ label, name, hint, ...rest }: TextFieldProps) {
  const [field] = useField(name)
  return (
    <FieldWrap label={label} name={name} hint={hint}>
      {(id, hasError) => <input id={id} {...field} {...rest} className={fieldClasses(hasError)} />}
    </FieldWrap>
  )
}

type SelectFieldProps = {
  label?: string
  name: string
  hint?: string
  children: ReactNode
} & SelectHTMLAttributes<HTMLSelectElement>

export function SelectField({ label, name, hint, children, ...rest }: SelectFieldProps) {
  const [field] = useField(name)
  return (
    <FieldWrap label={label} name={name} hint={hint}>
      {(id, hasError) => (
        <select id={id} {...field} {...rest} className={fieldClasses(hasError)}>
          {children}
        </select>
      )}
    </FieldWrap>
  )
}

type TextAreaFieldProps = { label?: string; name: string; hint?: string } & TextareaHTMLAttributes<HTMLTextAreaElement>

export function TextAreaField({ label, name, hint, ...rest }: TextAreaFieldProps) {
  const [field] = useField(name)
  return (
    <FieldWrap label={label} name={name} hint={hint}>
      {(id, hasError) => <textarea id={id} {...field} {...rest} className={fieldClasses(hasError)} />}
    </FieldWrap>
  )
}

interface CheckboxFieldProps {
  label: string
  name: string
}

export function CheckboxField({ label, name }: CheckboxFieldProps) {
  const [field] = useField({ name, type: 'checkbox' })
  return (
    <label className="flex cursor-pointer items-center gap-2 text-sm text-body">
      <input
        type="checkbox"
        {...field}
        className="h-4 w-4 rounded border-themed accent-brand-500"
      />
      {label}
    </label>
  )
}
