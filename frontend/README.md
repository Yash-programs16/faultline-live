# E-commerce storefront template (React 19 + Tailwind CSS 4)

[Leia em português](./README.pt-BR.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
![Free](https://img.shields.io/badge/price-free-brightgreen)

A customer-facing online store built with React 19, TypeScript, Vite 6 and Tailwind CSS v4. It covers the full shopping flow of an e-commerce front end: browsing a catalog with filters, reading product pages with variations and reviews, managing a cart with coupons, completing a multi-step checkout and reviewing orders in a customer account area. All content is mock data in TypeScript files, so the project runs without a backend.

Live preview: https://template.dev.br/preview/ecommerce-react/

## Views included

17 route-level pages, each lazy loaded as its own chunk:

- Home (`/`): highlighted products, category cards and best sellers
- Catalog (`/produtos`): product grid with filters (category, price, brand, rating), sorting and pagination
- Category (`/categoria/:slug`): catalog scoped to one of the 6 categories
- Search (`/busca`): text search over the product list
- Product (`/produto/:slug`): image gallery, color and size variations, reviews and related products
- Cart (`/carrinho`): line items, quantity stepper, coupon field and free shipping above R$299
- Checkout (`/checkout`): multi-step form with Formik and Yup, requires login
- Order confirmed (`/pedido/confirmado`): summary shown after checkout
- About (`/sobre`) and Contact (`/contato`): institutional pages
- Login (`/login`) and Register (`/registro`): standalone auth pages without the store chrome
- Account (`/conta`): overview, orders with a status timeline, addresses and wishlist (4 auth-guarded subpages)
- Not found: catch-all 404

A slide-in mini-cart (`CartDrawer`) is available from any page.

## Tech stack

- React 19.0 and React DOM 19.0
- TypeScript 5.7
- Vite 6.0 with `@vitejs/plugin-react` 4.3
- Tailwind CSS 4.0 via `@tailwindcss/vite` (CSS-first config, no `tailwind.config.js`)
- React Router 7.1 (`react-router-dom`)
- Zustand 5.0 for state, in 5 stores: `auth`, `cart`, `orders`, `ui`, `wishlist`
- Formik 2.4 and Yup 1.7 for the checkout and auth forms
- i18next 26.3 with `react-i18next` 17.0 and browser language detection (pt-BR, en, es)
- `vite-plugin-pwa` 1.3 for the service worker and install manifest

## Requirements

- Node.js 18 or newer
- npm (bundled with Node)

## Getting started

```bash
npm install
npm run dev
```

The dev server runs at http://localhost:5173.

## Build for production

```bash
npm run build
```

The build script runs `tsc --noEmit` for type checking, then `vite build`. Output goes to `dist/`. Use `npm run preview` to check the production bundle locally.

## Project structure

```
src/
  assets/styles/app.css   Tailwind v4 theme tokens (brand palette, fonts, breakpoints)
  components/
    cart/                 CartDrawer (mini-cart)
    catalog/              CatalogView, FilterSidebar, SortDropdown, CategoryCard
    layout/               Navbar, Footer, StoreLayout
    product/              ProductCard
    ui/                   Badge, Breadcrumb, FormikField, Pagination, PriceTag,
                          QuantityStepper, RatingStars, Skeleton, ToastStack, ...
  data/                   products.ts, categories.ts, query.ts, types.ts (mock data)
  i18n/                   i18next setup and locales (pt-BR.json, en.json, es.json)
  lib/format.ts           currency and date formatting helpers
  pages/                  route pages, plus account/ subpages
  stores/                 Zustand stores: auth, cart, orders, ui, wishlist
  App.tsx                 route table with lazy loading and auth guard
  main.tsx                entry point
```

## Theming and customization

Tailwind CSS v4 is configured in CSS, not in a JavaScript config file. All design tokens live in `src/assets/styles/app.css` inside the `@theme` block: the brand color scale (`--color-brand-25` to `--color-brand-950`, base color `#465fff`), the gray scale, title sizes, the Outfit font and extra breakpoints (`2xsm`, `xsm`, `3xl`). Change a token there and every utility class that uses it follows.

Dark mode is class-based (`@custom-variant dark`). The `ui` store toggles the `.dark` class on the root element and persists the choice.

### Wiring a real API

All content comes from `src/data/`. `products.ts` has 24 products across the 6 categories defined in `categories.ts`, each with price, stock, variations, rating and generated reviews. `query.ts` holds the filter and sort helpers used by the catalog. To connect a backend, replace the imports from `src/data/` with API calls and keep the interfaces in `src/data/types.ts` as the contract. The Zustand stores (`cart`, `orders`, `wishlist`, `auth`) are plain stores, so API calls can be added action by action.

### Internationalization

Interface strings live in `src/i18n/locales/` (pt-BR, en, es). The active language is detected from the browser and can be changed at runtime with the language switcher in the navbar. Route paths and mock content are in Portuguese.

### PWA

`vite-plugin-pwa` generates a service worker (auto update) and a web manifest, so the production build is installable. Icons are in `public/`.

## Other stacks

This storefront is part of a template catalog where each vertical exists in several stacks (HTML, React, Vue 3, Angular, Laravel). The other e-commerce versions and the other verticals are listed at https://template.dev.br.

## Support this project

This template is free to use. If it saved you some time, you can support the project with a donation at https://template.dev.br/doar?template=ecommerce-react.

## License

MIT. See [LICENSE](./LICENSE).
