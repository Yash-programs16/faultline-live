import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Formik, Form } from 'formik'
import * as Yup from 'yup'
import { useAuthStore, type Address } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { TextField } from '@/components/ui/FormikField'
import IconBase from '@/components/ui/IconBase'
import EmptyState from '@/components/ui/EmptyState'

type FormValues = Omit<Address, 'id' | 'isDefault'>

export default function AccountAddressesPage() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const addAddress = useAuthStore((s) => s.addAddress)
  const updateAddress = useAuthStore((s) => s.updateAddress)
  const removeAddress = useAuthStore((s) => s.removeAddress)
  const notify = useUiStore((s) => s.notify)

  const [editing, setEditing] = useState<Address | null>(null)
  const [showForm, setShowForm] = useState(false)

  const req = t('checkout.validation.required')
  const schema = Yup.object({
    label: Yup.string().required(req),
    recipient: Yup.string().required(req),
    street: Yup.string().required(req),
    number: Yup.string().required(req),
    district: Yup.string().required(req),
    city: Yup.string().required(req),
    state: Yup.string().required(req),
    zip: Yup.string().required(req),
  })

  const emptyValues: FormValues = {
    label: '', recipient: '', street: '', number: '', complement: '', district: '', city: '', state: '', zip: '',
  }

  function startAdd() {
    setEditing(null)
    setShowForm(true)
  }
  function startEdit(a: Address) {
    setEditing(a)
    setShowForm(true)
  }

  async function handleSubmit(values: FormValues) {
    if (editing) {
      await updateAddress(editing.id, values)
    } else {
      await addAddress(values)
    }
    notify(t('account.addressSaved'), 'success')
    setShowForm(false)
    setEditing(null)
  }

  const addresses = user?.addresses ?? []

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-heading">{t('account.addressBook')}</h2>
        {!showForm && (
          <button
            type="button"
            onClick={startAdd}
            className="flex h-10 items-center gap-2 rounded-lg bg-brand-500 px-4 text-sm font-medium text-white transition hover:bg-brand-600"
          >
            <IconBase name="plus" size={16} />
            {t('account.addAddress')}
          </button>
        )}
      </div>

      {showForm && (
        <div className="surface-card rounded-2xl border p-6">
          <h3 className="mb-4 font-semibold text-heading">
            {editing ? t('account.editAddress') : t('account.addAddress')}
          </h3>
          <Formik initialValues={editing ?? emptyValues} validationSchema={schema} onSubmit={handleSubmit} enableReinitialize>
            <Form className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <TextField name="label" label={t('account.addresses')} placeholder="Casa / Trabalho" />
                <TextField name="recipient" label={t('checkout.identification.name')} />
              </div>
              <div className="grid gap-4 sm:grid-cols-3">
                <TextField name="zip" label={t('checkout.shippingStep.zip')} />
                <div className="sm:col-span-2">
                  <TextField name="street" label={t('checkout.shippingStep.street')} />
                </div>
              </div>
              <div className="grid gap-4 sm:grid-cols-3">
                <TextField name="number" label={t('checkout.shippingStep.number')} />
                <div className="sm:col-span-2">
                  <TextField name="complement" label={t('checkout.shippingStep.complement')} />
                </div>
              </div>
              <div className="grid gap-4 sm:grid-cols-3">
                <TextField name="district" label={t('checkout.shippingStep.district')} />
                <TextField name="city" label={t('checkout.shippingStep.city')} />
                <TextField name="state" label={t('checkout.shippingStep.state')} />
              </div>
              <div className="flex gap-3 pt-2">
                <button type="submit" className="h-11 rounded-lg bg-brand-500 px-6 text-sm font-medium text-white hover:bg-brand-600">
                  {t('common.save')}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowForm(false)
                    setEditing(null)
                  }}
                  className="h-11 rounded-lg border border-themed px-6 text-sm font-medium text-body hover:bg-gray-50 dark:hover:bg-gray-800"
                >
                  {t('common.cancel')}
                </button>
              </div>
            </Form>
          </Formik>
        </div>
      )}

      {addresses.length === 0 && !showForm ? (
        <div className="surface-card rounded-2xl border">
          <EmptyState icon="map-pin" title={t('account.noAddresses')} description={t('account.noAddressesDesc')} />
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {addresses.map((a) => (
            <div key={a.id} className="surface-card rounded-2xl border p-5">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-heading">{a.label}</span>
                  {a.isDefault && (
                    <span className="rounded-full bg-brand-50 px-2 py-0.5 text-xs font-medium text-brand-600 dark:bg-brand-500/15">
                      {t('account.defaultAddress')}
                    </span>
                  )}
                </div>
                <div className="flex gap-1">
                  <button type="button" onClick={() => startEdit(a)} className="text-gray-400 hover:text-brand-600" aria-label={t('common.edit')}>
                    <IconBase name="edit" size={16} />
                  </button>
                  <button
                    type="button"
                    onClick={async () => {
                      await removeAddress(a.id)
                      notify(t('account.addressRemoved'), 'info')
                    }}
                    className="text-gray-400 hover:text-error-500"
                    aria-label={t('common.delete')}
                  >
                    <IconBase name="trash" size={16} />
                  </button>
                </div>
              </div>
              <p className="mt-2 text-sm text-body">{a.recipient}</p>
              <p className="text-sm text-muted">
                {a.street}, {a.number}
                {a.complement ? ` — ${a.complement}` : ''}
              </p>
              <p className="text-sm text-muted">
                {a.district}, {a.city}/{a.state} · {a.zip}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
