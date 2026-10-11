import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Formik, Form } from 'formik'
import * as Yup from 'yup'
import { useUiStore } from '@/stores/ui'
import Breadcrumb from '@/components/ui/Breadcrumb'
import { TextField, TextAreaField } from '@/components/ui/FormikField'
import IconBase from '@/components/ui/IconBase'

export default function ContactPage() {
  const { t } = useTranslation()
  const notify = useUiStore((s) => s.notifySuccess)
  const [submitting, setSubmitting] = useState(false)

  const req = t('checkout.validation.required')
  const schema = Yup.object({
    name: Yup.string().required(req),
    email: Yup.string().email(t('checkout.validation.email')).required(req),
    subject: Yup.string().required(req),
    message: Yup.string().required(req).min(10, t('checkout.validation.min')),
  })

  const info = [
    { icon: 'mail', label: t('contact.email'), value: 'contato@loja.com' },
    { icon: 'phone', label: t('contact.phone'), value: '+55 11 4000-0000' },
    { icon: 'map-pin', label: t('contact.address'), value: t('contact.addressValue') },
    { icon: 'clock', label: t('contact.hours'), value: t('contact.hoursValue') },
  ]

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('common.contact') }]} />

      <div className="mt-8 text-center">
        <h1 className="text-3xl font-bold text-heading">{t('contact.title')}</h1>
        <p className="mx-auto mt-2 max-w-md text-sm text-muted">{t('contact.subtitle')}</p>
      </div>

      <div className="mt-10 grid gap-8 lg:grid-cols-[1fr_320px]">
        <div className="surface-card rounded-2xl border p-6 sm:p-8">
          <Formik
            initialValues={{ name: '', email: '', subject: '', message: '' }}
            validationSchema={schema}
            onSubmit={(_, { resetForm }) => {
              setSubmitting(true)
              setTimeout(() => {
                setSubmitting(false)
                notify(t('contact.success'))
                resetForm()
              }, 600)
            }}
          >
            <Form className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <TextField name="name" label={t('contact.name')} />
                <TextField name="email" type="email" label={t('contact.email')} />
              </div>
              <TextField name="subject" label={t('contact.subject')} />
              <TextAreaField name="message" label={t('contact.message')} rows={5} />
              <button
                type="submit"
                disabled={submitting}
                className="flex h-11 items-center gap-2 rounded-lg bg-brand-500 px-6 text-sm font-medium text-white transition hover:bg-brand-600 disabled:opacity-60"
              >
                <IconBase name="mail" size={16} />
                {submitting ? t('contact.sending') : t('contact.send')}
              </button>
            </Form>
          </Formik>
        </div>

        <div className="space-y-4">
          <div className="surface-card rounded-2xl border p-6">
            <h2 className="font-semibold text-heading">{t('contact.infoTitle')}</h2>
            <div className="mt-4 space-y-4">
              {info.map((item) => (
                <div key={item.label} className="flex items-start gap-3">
                  <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-500 dark:bg-brand-500/15">
                    <IconBase name={item.icon} size={18} />
                  </span>
                  <div>
                    <p className="text-xs text-muted">{item.label}</p>
                    <p className="text-sm font-medium text-heading">{item.value}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
