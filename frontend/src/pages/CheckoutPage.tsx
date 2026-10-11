import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Formik, Form } from 'formik'
import * as Yup from 'yup'
import { useCartStore } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'
import { useOrdersStore } from '@/stores/orders'
import { useUiStore } from '@/stores/ui'
import { formatPrice } from '@/lib/format'
import { TextField } from '@/components/ui/FormikField'
import IconBase from '@/components/ui/IconBase'
import EmptyState from '@/components/ui/EmptyState'

type PaymentMethod = 'card' | 'pix' | 'boleto'

interface Address {
  id: number
  label: string
  recipient: string
  street: string
  number: string
  complement?: string
  district: string
  city: string
  state: string
  zip: string
  isDefault?: boolean
}

interface CheckoutValues {
  addressId: number
  payment: PaymentMethod
  cardNumber: string
  cardName: string
  cardExpiry: string
  cardCvv: string
}

const STEPS = ['address', 'payment', 'review'] as const
type Step = (typeof STEPS)[number]

export default function CheckoutPage() {
  const { t, i18n } = useTranslation()
  const lang = i18n.language
  const navigate = useNavigate()

  const user = useAuthStore((s) => s.user)
  const addresses = user?.addresses ?? []
  const items = useCartStore((s) => s.items)
  const subtotal = useCartStore((s) => s.subtotal())
  const discount = useCartStore((s) => s.discount())
  const shipping = useCartStore((s) => s.shipping())
  const total = useCartStore((s) => s.total())
  const clearCart = useCartStore((s) => s.clear)
  const placeOrder = useOrdersStore((s) => s.placeOrder)
  const notify = useUiStore((s) => s.notify)

  const [stepIndex, setStepIndex] = useState(0)
  const step: Step = STEPS[stepIndex]

  if (items.length === 0) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
        <EmptyState
          icon="cart"
          title={t('cart.empty')}
          description={t('cart.emptyDesc')}
          action={
            <Link to="/produtos" className="rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600">
              {t('cart.continueShopping')}
            </Link>
          }
        />
      </div>
    )
  }

  const req = t('checkout.validation.required')
  const stepSchemas: Record<Step, Yup.AnyObjectSchema> = {
    address: Yup.object({
      addressId: Yup.number().required(req),
    }),
    payment: Yup.object({
      payment: Yup.string().required(req),
      cardNumber: Yup.string().when('payment', {
        is: 'card',
        then: (s) => s.required(req).min(13, t('checkout.validation.cardNumber')),
      }),
      cardName: Yup.string().when('payment', { is: 'card', then: (s) => s.required(req) }),
      cardExpiry: Yup.string().when('payment', { is: 'card', then: (s) => s.required(req) }),
      cardCvv: Yup.string().when('payment', {
        is: 'card',
        then: (s) => s.required(req).min(3, t('checkout.validation.cvv')),
      }),
    }),
    review: Yup.object({}),
  }

  const defaultAddress = addresses.find(a => a.isDefault) ?? addresses[0]

  const initialValues: CheckoutValues = {
    addressId: defaultAddress?.id ?? 0,
    payment: 'card',
    cardNumber: '',
    cardName: '',
    cardExpiry: '',
    cardCvv: '',
  }

  function handleSubmit(values: CheckoutValues) {
    if (stepIndex < STEPS.length - 1) {
      setStepIndex((i) => i + 1)
      return
    }

    // Transform cart items to API format
    const apiItems = items.map(item => ({
      productId: item.productId,
      slug: item.slug,
      name: item.name,
      price: item.price,
      image: item.image,
      color: item.color,
      size: item.size,
      qty: item.qty,
      stock: item.stock,
    }))

    placeOrder({
      items: apiItems,
      addressId: values.addressId,
      paymentMethod: values.payment,
    }).then(orderId => {
      if (orderId) {
        clearCart()
        navigate('/pedido/confirmado')
      } else {
        notify(t('checkout.error'), 'error')
      }
    })
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
      <h1 className="text-2xl font-bold text-heading sm:text-3xl">{t('checkout.title')}</h1>

      {/* Stepper */}
      <div className="mt-6 flex items-center">
        {STEPS.map((s, i) => (
          <div key={s} className="flex flex-1 items-center last:flex-none">
            <div className="flex items-center gap-2">
              <span
                className={`flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold transition ${
                  i < stepIndex
                    ? 'bg-success-500 text-white'
                    : i === stepIndex
                      ? 'bg-brand-500 text-white'
                      : 'border border-themed text-muted'
                }`}
              >
                {i < stepIndex ? <IconBase name="check" size={16} /> : i + 1}
              </span>
              <span className={`hidden text-sm font-medium sm:block ${i === stepIndex ? 'text-heading' : 'text-muted'}`}>
                {t(`checkout.steps.${s}`)}
              </span>
            </div>
            {i < STEPS.length - 1 && (
              <span className={`mx-3 h-px flex-1 ${i < stepIndex ? 'bg-success-500' : 'bg-gray-200 dark:bg-gray-700'}`} />
            )}
          </div>
        ))}
      </div>

      <Formik
        initialValues={initialValues}
        validationSchema={stepSchemas[step]}
        onSubmit={handleSubmit}
        validateOnMount={false}
      >
        {({ values, setFieldValue, errors, touched }) => (
          <Form className="mt-8 grid gap-8 lg:grid-cols-[1fr_340px]">
            <div className="surface-card rounded-2xl border p-6">
              {step === 'address' && (
                <div className="space-y-4">
                  <h2 className="text-lg font-semibold text-heading">{t('checkout.address.title')}</h2>
                  
                  {addresses.length > 0 ? (
                    <div className="space-y-2">
                      <p className="text-sm text-muted">{t('checkout.address.select')}</p>
                      {addresses.map(addr => (
                        <label
                          key={addr.id}
                          className={`flex cursor-pointer items-center gap-3 rounded-lg border px-4 py-3 transition ${
                            values.addressId === addr.id
                              ? 'border-brand-500 bg-brand-50 dark:bg-brand-500/10'
                              : 'border-themed'
                          }`}
                        >
                          <input
                            type="radio"
                            name="addressId"
                            value={addr.id}
                            checked={values.addressId === addr.id}
                            onChange={() => setFieldValue('addressId', addr.id)}
                            className="h-4 w-4 accent-brand-500"
                          />
                          <div className="flex-1">
                            <p className="font-medium text-heading">{addr.label}</p>
                            <p className="text-sm text-muted">
                              {addr.recipient} — {addr.street}, {addr.number}{addr.complement ? `, ${addr.complement}` : ''} — {addr.district}, {addr.city}/{addr.state} — {addr.zip}
                            </p>
                          </div>
                          {addr.isDefault && <span className="px-2 py-0.5 text-xs font-medium bg-brand-100 text-brand-700 rounded">{t('checkout.address.default')}</span>}
                        </label>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-warning-600">{t('checkout.address.none')}</p>
                  )}

                  <Link
                    to="/conta/enderecos"
                    className="mt-4 inline-flex items-center gap-1 text-sm font-medium text-brand-600 hover:underline"
                  >
                    <IconBase name="plus" size={14} />
                    {t('checkout.address.addNew')}
                  </Link>
                </div>
              )}

              {step === 'payment' && (
                <div className="space-y-4">
                  <h2 className="text-lg font-semibold text-heading">{t('checkout.payment.title')}</h2>
                  <div className="grid grid-cols-3 gap-2">
                    {(['card', 'pix', 'boleto'] as PaymentMethod[]).map((m) => (
                      <button
                        key={m}
                        type="button"
                        onClick={() => setFieldValue('payment', m)}
                        className={`flex flex-col items-center gap-2 rounded-lg border px-3 py-4 text-sm font-medium transition ${
                          values.payment === m
                            ? 'border-brand-500 bg-brand-50 text-brand-600 dark:bg-brand-500/10'
                            : 'border-themed text-body'
                        }`}
                      >
                        <IconBase name={m === 'card' ? 'credit-card' : m === 'pix' ? 'zap' : 'list'} size={22} />
                        {t(`checkout.payment.${m}`)}
                      </button>
                    ))}
                  </div>

                  {values.payment === 'card' && (
                    <div className="space-y-4 pt-2 border-t border-themed">
                      <TextField name="cardNumber" label={t('checkout.payment.cardNumber')} placeholder="0000 0000 0000 0000" />
                      <TextField name="cardName" label={t('checkout.payment.cardName')} />
                      <div className="grid grid-cols-2 gap-4">
                        <TextField name="cardExpiry" label={t('checkout.payment.cardExpiry')} placeholder="MM/AA" />
                        <TextField name="cardCvv" label={t('checkout.payment.cardCvv')} placeholder="123" />
                      </div>
                    </div>
                  )}
                  {values.payment === 'pix' && (
                    <p className="rounded-lg bg-blue-light-50 p-4 text-sm text-blue-light-700 dark:bg-blue-light-500/10">
                      {t('checkout.payment.pixInfo')}
                    </p>
                  )}
                  {values.payment === 'boleto' && (
                    <p className="rounded-lg bg-warning-50 p-4 text-sm text-warning-700 dark:bg-warning-500/10">
                      {t('checkout.payment.boletoInfo')}
                    </p>
                  )}
                </div>
              )}

              {step === 'review' && (
                <div className="space-y-5">
                  <h2 className="text-lg font-semibold text-heading">{t('checkout.review.title')}</h2>
                  <div className="rounded-lg border border-themed p-4">
                    <p className="text-sm font-medium text-heading">{t('checkout.review.deliverTo')}</p>
                    <p className="mt-1 text-sm text-muted">
                      {addresses.find(a => a.id === values.addressId)?.recipient ?? ''} — 
                      {addresses.find(a => a.id === values.addressId)?.street ?? ''}, 
                      {addresses.find(a => a.id === values.addressId)?.number ?? ''}, 
                      {addresses.find(a => a.id === values.addressId)?.district ?? ''}, 
                      {addresses.find(a => a.id === values.addressId)?.city ?? ''}/{addresses.find(a => a.id === values.addressId)?.state ?? ''}
                    </p>
                  </div>
                  <div className="rounded-lg border border-themed p-4">
                    <p className="text-sm font-medium text-heading">{t('checkout.review.paymentMethod')}</p>
                    <p className="mt-1 text-sm text-muted">{t(`checkout.payment.${values.payment}`)}</p>
                  </div>
                  <div className="rounded-lg border border-themed p-4">
                    <p className="mb-2 text-sm font-medium text-heading">{t('checkout.review.items')}</p>
                    <div className="space-y-2">
                      {items.map((item) => (
                        <div key={item.id} className="flex items-center justify-between text-sm">
                          <span className="text-body">
                            {item.qty}× {item.name}
                          </span>
                          <span className="text-heading">{formatPrice(item.price * item.qty, lang)}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Nav buttons */}
              <div className="mt-8 flex items-center justify-between gap-3">
                {stepIndex > 0 ? (
                  <button
                    type="button"
                    onClick={() => setStepIndex((i) => i - 1)}
                    className="flex h-11 items-center gap-2 rounded-lg border border-themed px-5 text-sm font-medium text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
                  >
                    <IconBase name="arrow-left" size={16} />
                    {t('common.back')}
                  </button>
                ) : (
                  <Link
                    to="/carrinho"
                    className="flex h-11 items-center gap-2 rounded-lg border border-themed px-5 text-sm font-medium text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
                  >
                    <IconBase name="arrow-left" size={16} />
                    {t('cart.title')}
                  </Link>
                )}
                <button
                  type="submit"
                  className="flex h-11 flex-1 items-center justify-center gap-2 rounded-lg bg-brand-500 px-6 text-sm font-medium text-white transition hover:bg-brand-600 sm:flex-none"
                >
                  {step === 'review' ? t('checkout.review.placeOrder') : t('common.continue')}
                  <IconBase name="arrow-right" size={16} />
                </button>
              </div>
            </div>

            {/* Summary */}
            <div>
              <div className="surface-card sticky top-32 rounded-2xl border p-6">
                <h2 className="text-lg font-semibold text-heading">{t('checkout.summary')}</h2>
                <div className="mt-4 max-h-52 space-y-3 overflow-y-auto pr-1">
                  {items.map((item) => (
                    <div key={item.id} className="flex gap-3">
                      <img src={item.image} alt={item.name} className="h-14 w-14 rounded-lg border border-themed object-cover" />
                      <div className="flex flex-1 flex-col text-sm">
                        <span className="line-clamp-1 text-heading">{item.name}</span>
                        <span className="text-xs text-muted">×{item.qty}</span>
                        <span className="mt-auto font-medium text-heading">{formatPrice(item.price * item.qty, lang)}</span>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="mt-5 space-y-2 border-t border-themed pt-4 text-sm">
                  <div className="flex justify-between text-body">
                    <span>{t('cart.subtotal')}</span>
                    <span>{formatPrice(subtotal, lang)}</span>
                  </div>
                  {discount > 0 && (
                    <div className="flex justify-between text-success-600">
                      <span>{t('cart.discount')}</span>
                      <span>-{formatPrice(discount, lang)}</span>
                    </div>
                  )}
                  <div className="flex justify-between text-body">
                    <span>{t('cart.shipping')}</span>
                    <span>{shipping === 0 ? t('cart.shippingFree') : formatPrice(shipping, lang)}</span>
                  </div>
                  <div className="flex justify-between border-t border-themed pt-2 text-base font-bold text-heading">
                    <span>{t('cart.total')}</span>
                    <span>{formatPrice(total, lang)}</span>
                  </div>
                </div>
              </div>
            </div>
          </Form>
        )}
      </Formik>
    </div>
  )
}