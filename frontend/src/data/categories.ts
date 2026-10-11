import type { Category } from './types'

/**
 * Categories. The `name`/`description` here are Portuguese fallbacks; the UI
 * prefers i18n keys `categories.<slug>.name` / `categories.<slug>.desc` so the
 * label localizes. A converter to another framework should mirror this: data
 * is canonical, display strings come from the i18n bundle.
 */
export const categories: Category[] = [
  { id: 1, slug: 'eletronicos', name: 'Eletrônicos', icon: 'package', description: 'Gadgets e dispositivos' },
  { id: 2, slug: 'moda', name: 'Moda', icon: 'tag', description: 'Roupas e acessórios' },
  { id: 3, slug: 'casa', name: 'Casa & Decoração', icon: 'home', description: 'Tudo para o seu lar' },
  { id: 4, slug: 'esportes', name: 'Esportes', icon: 'zap', description: 'Fitness e ar livre' },
  { id: 5, slug: 'beleza', name: 'Beleza', icon: 'star', description: 'Cuidados e cosméticos' },
  { id: 6, slug: 'acessorios', name: 'Acessórios', icon: 'gift', description: 'Complementos e mais' },
]

export function findCategory(slug: string): Category | undefined {
  return categories.find((c) => c.slug === slug)
}
