# Template de loja e-commerce (React 19 + Tailwind CSS 4)

[Read in English](./README.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
![Free](https://img.shields.io/badge/price-free-brightgreen)

Uma loja online voltada ao cliente final, construída com React 19, TypeScript, Vite 6 e Tailwind CSS v4. Ela cobre o fluxo completo de compra de um front-end de e-commerce: navegação no catálogo com filtros, páginas de produto com variações e avaliações, carrinho com cupom, checkout multi-step e área de conta do cliente para acompanhar pedidos. Todo o conteúdo é dado mockado em arquivos TypeScript, então o projeto roda sem backend.

Preview ao vivo: https://template.dev.br/preview/ecommerce-react/

## Views incluídas

17 páginas de rota, cada uma carregada de forma lazy como um chunk próprio:

- Home (`/`): produtos em destaque, cards de categoria e mais vendidos
- Catálogo (`/produtos`): grid de produtos com filtros (categoria, preço, marca, avaliação), ordenação e paginação
- Categoria (`/categoria/:slug`): catálogo restrito a uma das 6 categorias
- Busca (`/busca`): busca por texto na lista de produtos
- Produto (`/produto/:slug`): galeria de imagens, variações de cor e tamanho, avaliações e relacionados
- Carrinho (`/carrinho`): itens, seletor de quantidade, campo de cupom e frete grátis acima de R$299
- Checkout (`/checkout`): formulário multi-step com Formik e Yup, exige login
- Pedido confirmado (`/pedido/confirmado`): resumo exibido após o checkout
- Sobre (`/sobre`) e Contato (`/contato`): páginas institucionais
- Login (`/login`) e Registro (`/registro`): páginas de auth standalone, sem o chrome da loja
- Conta (`/conta`): visão geral, pedidos com timeline de status, endereços e favoritos (4 subpáginas protegidas por auth)
- Não encontrado: 404 catch-all

Um mini-cart deslizante (`CartDrawer`) fica disponível em qualquer página.

## Stack

- React 19.0 e React DOM 19.0
- TypeScript 5.7
- Vite 6.0 com `@vitejs/plugin-react` 4.3
- Tailwind CSS 4.0 via `@tailwindcss/vite` (configuração CSS-first, sem `tailwind.config.js`)
- React Router 7.1 (`react-router-dom`)
- Zustand 5.0 para estado, em 5 stores: `auth`, `cart`, `orders`, `ui`, `wishlist`
- Formik 2.4 e Yup 1.7 nos formulários de checkout e auth
- i18next 26.3 com `react-i18next` 17.0 e detecção de idioma do navegador (pt-BR, en, es)
- `vite-plugin-pwa` 1.3 para o service worker e o manifest de instalação

## Requisitos

- Node.js 18 ou mais recente
- npm (incluso no Node)

## Como rodar

```bash
npm install
npm run dev
```

O servidor de desenvolvimento roda em http://localhost:5173.

## Build de produção

```bash
npm run build
```

O script de build roda `tsc --noEmit` para checagem de tipos e depois `vite build`. A saída vai para `dist/`. Use `npm run preview` para conferir o bundle de produção localmente.

## Estrutura do projeto

```
src/
  assets/styles/app.css   Tokens do tema Tailwind v4 (paleta brand, fontes, breakpoints)
  components/
    cart/                 CartDrawer (mini-cart)
    catalog/              CatalogView, FilterSidebar, SortDropdown, CategoryCard
    layout/               Navbar, Footer, StoreLayout
    product/              ProductCard
    ui/                   Badge, Breadcrumb, FormikField, Pagination, PriceTag,
                          QuantityStepper, RatingStars, Skeleton, ToastStack, ...
  data/                   products.ts, categories.ts, query.ts, types.ts (dados mockados)
  i18n/                   setup do i18next e locales (pt-BR.json, en.json, es.json)
  lib/format.ts           helpers de formatação de moeda e data
  pages/                  páginas de rota, mais as subpáginas de account/
  stores/                 stores Zustand: auth, cart, orders, ui, wishlist
  App.tsx                 tabela de rotas com lazy loading e guard de auth
  main.tsx                ponto de entrada
```

## Temas e customização

O Tailwind CSS v4 é configurado em CSS, não em arquivo de configuração JavaScript. Todos os tokens de design ficam em `src/assets/styles/app.css`, dentro do bloco `@theme`: a escala de cor da marca (`--color-brand-25` a `--color-brand-950`, cor base `#465fff`), a escala de cinzas, tamanhos de título, a fonte Outfit e breakpoints extras (`2xsm`, `xsm`, `3xl`). Altere um token ali e todas as classes utilitárias que o usam acompanham.

O dark mode é baseado em classe (`@custom-variant dark`). A store `ui` alterna a classe `.dark` no elemento raiz e persiste a escolha.

### Conectando uma API real

Todo o conteúdo vem de `src/data/`. `products.ts` tem 24 produtos nas 6 categorias definidas em `categories.ts`, cada um com preço, estoque, variações, avaliação e reviews geradas. `query.ts` concentra os helpers de filtro e ordenação usados pelo catálogo. Para conectar um backend, troque os imports de `src/data/` por chamadas de API e mantenha as interfaces de `src/data/types.ts` como contrato. As stores Zustand (`cart`, `orders`, `wishlist`, `auth`) são stores simples, então dá para adicionar chamadas de API ação por ação.

### Internacionalização

As strings de interface ficam em `src/i18n/locales/` (pt-BR, en, es). O idioma ativo é detectado do navegador e pode ser trocado em tempo de execução pelo seletor de idioma na navbar. Os caminhos de rota e o conteúdo mockado estão em português.

### PWA

O `vite-plugin-pwa` gera um service worker (auto update) e um web manifest, então o build de produção é instalável. Os ícones ficam em `public/`.

## Outras stacks

Esta loja faz parte de um catálogo de templates em que cada vertical existe em várias stacks (HTML, React, Vue 3, Angular, Laravel). As outras versões de e-commerce e os demais verticais estão em https://template.dev.br.

## Apoie este projeto

Este template é gratuito. Se ele economizou seu tempo, você pode apoiar o projeto com uma doação em https://template.dev.br/doar?template=ecommerce-react.

## Licença

MIT. Veja [LICENSE](./LICENSE).
