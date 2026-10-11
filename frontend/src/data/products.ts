import type { Product } from './types'

// Curated Unsplash set — produtos variados (eletrônicos, moda, acessórios, casa). IDs verificados.
const IMG_PRODUCTS = [
  '1523275335684-37898b6baf30', '1491553895911-0055eca6402d', '1542291026-7eec264c27ff',
  '1585386959984-a4155224a1ad', '1560769629-975ec94e6a86', '1608231387042-66d1773070a5',
  '1441986300917-64674bd600d8', '1526170375885-4d8ecf77b99f', '1572635196237-14b3f281503f',
  '1503341504253-dff4815485f1', '1504204267155-aaad8e81290d',
]
function imgHash(s: string): number { let h = 0; for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0; return h }
function unsplash(id: string, w: number, h: number): string {
  return `https://images.unsplash.com/photo-${id}?w=${w}&h=${h}&fit=crop&q=70&auto=format`
}
/** Curated product photo gallery (determinístico por slug + índice). */
export function productImages(slug: string, count = 4): string[] {
  return Array.from({ length: count }, (_, i) =>
    unsplash(IMG_PRODUCTS[imgHash(`${slug}-${i}`) % IMG_PRODUCTS.length], 600, 600)
  )
}

const COLORS = ['#101828', '#465fff', '#f04438', '#12b76a', '#fdb022', '#ffffff']
const SIZES_CLOTHES = ['PP', 'P', 'M', 'G', 'GG']
const SIZES_SHOES = ['38', '39', '40', '41', '42', '43']
const SIZES_NONE: string[] = []

const sampleReviews = (rating: number) => [
  { author: 'Marina S.', rating: Math.min(5, Math.round(rating)), date: '2026-04-12', comment: 'Produto excelente, superou minhas expectativas. Recomendo!' },
  { author: 'Carlos M.', rating: Math.max(3, Math.round(rating) - 1), date: '2026-03-28', comment: 'Boa qualidade pelo preço. Entrega foi rápida.' },
  { author: 'Ana P.', rating: 5, date: '2026-03-02', comment: 'Amei! Chegou bem embalado e funciona perfeitamente.' },
]

type Seed = {
  name: string
  slug: string
  price: number
  priceOld?: number
  category: string
  brand: string
  rating: number
  reviewsCount: number
  stock: number
  sizes: string[]
  flags?: Partial<Pick<Product, 'featured' | 'bestseller' | 'isNew'>>
}

const seeds: Seed[] = [
  // Eletrônicos
  { name: 'Fone Bluetooth NovaSound Pro', slug: 'fone-novasound-pro', price: 349.9, priceOld: 499.9, category: 'eletronicos', brand: 'NovaSound', rating: 4.7, reviewsCount: 214, stock: 32, sizes: SIZES_NONE, flags: { featured: true, bestseller: true } },
  { name: 'Smartwatch Pulse X2', slug: 'smartwatch-pulse-x2', price: 599.0, priceOld: 799.0, category: 'eletronicos', brand: 'Pulse', rating: 4.5, reviewsCount: 158, stock: 18, sizes: SIZES_NONE, flags: { featured: true, isNew: true } },
  { name: 'Caixa de Som Portátil BoomGo', slug: 'caixa-som-boomgo', price: 229.9, category: 'eletronicos', brand: 'BoomGo', rating: 4.3, reviewsCount: 96, stock: 54, sizes: SIZES_NONE, flags: { bestseller: true } },
  { name: 'Teclado Mecânico KeyForge RGB', slug: 'teclado-keyforge-rgb', price: 419.0, priceOld: 489.0, category: 'eletronicos', brand: 'KeyForge', rating: 4.8, reviewsCount: 312, stock: 7, sizes: SIZES_NONE, flags: { featured: true } },
  { name: 'Webcam StreamCam 4K', slug: 'webcam-streamcam-4k', price: 389.9, category: 'eletronicos', brand: 'StreamCam', rating: 4.2, reviewsCount: 44, stock: 0, sizes: SIZES_NONE },

  // Moda
  { name: 'Camiseta Essential Algodão', slug: 'camiseta-essential', price: 79.9, priceOld: 99.9, category: 'moda', brand: 'Essential', rating: 4.6, reviewsCount: 187, stock: 120, sizes: SIZES_CLOTHES, flags: { bestseller: true } },
  { name: 'Jaqueta Corta-Vento UrbanFlow', slug: 'jaqueta-urbanflow', price: 259.9, priceOld: 349.9, category: 'moda', brand: 'UrbanFlow', rating: 4.4, reviewsCount: 72, stock: 23, sizes: SIZES_CLOTHES, flags: { featured: true } },
  { name: 'Calça Jeans Slim Denim Co.', slug: 'calca-jeans-denim', price: 189.9, category: 'moda', brand: 'Denim Co.', rating: 4.1, reviewsCount: 58, stock: 41, sizes: SIZES_CLOTHES },
  { name: 'Vestido Floral Primavera', slug: 'vestido-floral', price: 219.9, priceOld: 279.9, category: 'moda', brand: 'Primavera', rating: 4.7, reviewsCount: 134, stock: 16, sizes: SIZES_CLOTHES, flags: { isNew: true } },

  // Casa
  { name: 'Luminária de Mesa Aurora', slug: 'luminaria-aurora', price: 159.9, category: 'casa', brand: 'Aurora', rating: 4.5, reviewsCount: 89, stock: 38, sizes: SIZES_NONE, flags: { featured: true } },
  { name: 'Jogo de Panelas Chef Master', slug: 'panelas-chef-master', price: 489.0, priceOld: 649.0, category: 'casa', brand: 'Chef Master', rating: 4.8, reviewsCount: 203, stock: 12, sizes: SIZES_NONE, flags: { bestseller: true } },
  { name: 'Manta Aconchego Soft', slug: 'manta-aconchego', price: 119.9, priceOld: 149.9, category: 'casa', brand: 'Aconchego', rating: 4.6, reviewsCount: 67, stock: 75, sizes: SIZES_NONE },
  { name: 'Difusor de Aromas ZenAir', slug: 'difusor-zenair', price: 99.9, category: 'casa', brand: 'ZenAir', rating: 4.3, reviewsCount: 51, stock: 60, sizes: SIZES_NONE, flags: { isNew: true } },

  // Esportes
  { name: 'Tênis de Corrida SpeedRun', slug: 'tenis-speedrun', price: 399.9, priceOld: 529.9, category: 'esportes', brand: 'SpeedRun', rating: 4.7, reviewsCount: 298, stock: 26, sizes: SIZES_SHOES, flags: { featured: true, bestseller: true } },
  { name: 'Garrafa Térmica HydroFit 1L', slug: 'garrafa-hydrofit', price: 89.9, category: 'esportes', brand: 'HydroFit', rating: 4.5, reviewsCount: 142, stock: 88, sizes: SIZES_NONE },
  { name: 'Kit Halteres Ajustáveis PowerLift', slug: 'halteres-powerlift', price: 549.0, priceOld: 699.0, category: 'esportes', brand: 'PowerLift', rating: 4.8, reviewsCount: 76, stock: 9, sizes: SIZES_NONE, flags: { featured: true } },
  { name: 'Mochila Trail Adventure 30L', slug: 'mochila-trail-30l', price: 279.9, category: 'esportes', brand: 'Adventure', rating: 4.4, reviewsCount: 63, stock: 34, sizes: SIZES_NONE, flags: { isNew: true } },

  // Beleza
  { name: 'Sérum Facial Glow Vitamina C', slug: 'serum-glow-vitc', price: 129.9, priceOld: 169.9, category: 'beleza', brand: 'Glow', rating: 4.6, reviewsCount: 221, stock: 47, sizes: SIZES_NONE, flags: { bestseller: true } },
  { name: 'Kit Skincare Pure Routine', slug: 'kit-pure-routine', price: 299.9, priceOld: 399.9, category: 'beleza', brand: 'Pure', rating: 4.7, reviewsCount: 95, stock: 21, sizes: SIZES_NONE, flags: { featured: true } },
  { name: 'Perfume Essência Noir 100ml', slug: 'perfume-noir', price: 359.0, category: 'beleza', brand: 'Essência', rating: 4.5, reviewsCount: 118, stock: 15, sizes: SIZES_NONE },
  { name: 'Secador Profissional IonDry', slug: 'secador-iondry', price: 269.9, priceOld: 329.9, category: 'beleza', brand: 'IonDry', rating: 4.3, reviewsCount: 54, stock: 28, sizes: SIZES_NONE, flags: { isNew: true } },

  // Acessórios
  { name: 'Óculos de Sol Horizon UV400', slug: 'oculos-horizon', price: 199.9, priceOld: 259.9, category: 'acessorios', brand: 'Horizon', rating: 4.4, reviewsCount: 87, stock: 43, sizes: SIZES_NONE, flags: { featured: true } },
  { name: 'Carteira de Couro Heritage', slug: 'carteira-heritage', price: 149.9, category: 'acessorios', brand: 'Heritage', rating: 4.6, reviewsCount: 109, stock: 56, sizes: SIZES_NONE, flags: { bestseller: true } },
  { name: 'Mochila Notebook Metro 15"', slug: 'mochila-metro-15', price: 229.9, priceOld: 289.9, category: 'acessorios', brand: 'Metro', rating: 4.5, reviewsCount: 132, stock: 30, sizes: SIZES_NONE },
]

export const products: Product[] = seeds.map((s, i) => ({
  id: i + 1,
  name: s.name,
  slug: s.slug,
  price: s.price,
  priceOld: s.priceOld,
  category: s.category,
  brand: s.brand,
  rating: s.rating,
  reviewsCount: s.reviewsCount,
  stock: s.stock,
  description:
    `${s.name} da marca ${s.brand}. Produto de alta qualidade, ideal para o dia a dia, ` +
    `com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.`,
  variants: {
    colors: COLORS.slice(0, 4),
    sizes: s.sizes,
  },
  specs: [
    { label: 'Marca', value: s.brand },
    { label: 'Garantia', value: '12 meses' },
    { label: 'Origem', value: 'Nacional' },
    { label: 'SKU', value: `SKU-${1000 + i}` },
  ],
  reviews: sampleReviews(s.rating),
  ...s.flags,
  createdAt: new Date(2026, 0, 1 + i * 3).toISOString(),
}))

export const brands = Array.from(new Set(products.map((p) => p.brand))).sort()

export const priceBounds = {
  min: Math.floor(Math.min(...products.map((p) => p.price))),
  max: Math.ceil(Math.max(...products.map((p) => p.price))),
}

export function findProduct(slug: string): Product | undefined {
  return products.find((p) => p.slug === slug)
}

export function relatedProducts(product: Product, limit = 4): Product[] {
  return products
    .filter((p) => p.category === product.category && p.id !== product.id)
    .slice(0, limit)
}
