-- 002_seed_data.sql
-- Seed data for Faultline Live demo
-- This migration is idempotent and can be run multiple times

-- Insert categories
INSERT INTO categories (slug, name, icon, description) VALUES
    ('eletronicos', 'Eletrônicos', 'package', 'Gadgets e dispositivos'),
    ('moda', 'Moda', 'tag', 'Roupas e acessórios'),
    ('casa', 'Casa & Decoração', 'home', 'Tudo para o seu lar'),
    ('esportes', 'Esportes', 'zap', 'Fitness e ar livre'),
    ('beleza', 'Beleza', 'star', 'Cuidados e cosméticos'),
    ('acessorios', 'Acessórios', 'gift', 'Complementos e mais')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    description = EXCLUDED.description;

-- Helper function to get category ID
-- We'll use CTEs to insert products with proper category references

-- Insert products (idempotent using ON CONFLICT on slug)
WITH cat AS (
    SELECT id, slug FROM categories
)
INSERT INTO products (slug, name, description, price, price_old, category_id, brand, rating, reviews_count, stock, featured, bestseller, is_new)
SELECT * FROM (VALUES
    -- Eletrônicos
    ('fone-novasound-pro', 'Fone Bluetooth NovaSound Pro', 'Fone Bluetooth NovaSound Pro da marca NovaSound. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 34990, 49990, (SELECT id FROM cat WHERE slug='eletronicos'), 'NovaSound', 4.7, 214, 32, true, true, false),
    ('smartwatch-pulse-x2', 'Smartwatch Pulse X2', 'Smartwatch Pulse X2 da marca Pulse. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 59900, 79900, (SELECT id FROM cat WHERE slug='eletronicos'), 'Pulse', 4.5, 158, 18, true, false, true),
    ('caixa-som-boomgo', 'Caixa de Som Portátil BoomGo', 'Caixa de Som Portátil BoomGo da marca BoomGo. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 22990, NULL, (SELECT id FROM cat WHERE slug='eletronicos'), 'BoomGo', 4.3, 96, 54, false, true, false),
    ('teclado-keyforge-rgb', 'Teclado Mecânico KeyForge RGB', 'Teclado Mecânico KeyForge RGB da marca KeyForge. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 41900, 48900, (SELECT id FROM cat WHERE slug='eletronicos'), 'KeyForge', 4.8, 312, 7, true, false, false),
    ('webcam-streamcam-4k', 'Webcam StreamCam 4K', 'Webcam StreamCam 4K da marca StreamCam. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 38990, NULL, (SELECT id FROM cat WHERE slug='eletronicos'), 'StreamCam', 4.2, 44, 0, false, false, false),

    -- Moda
    ('camiseta-essential', 'Camiseta Essential Algodão', 'Camiseta Essential Algodão da marca Essential. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 7990, 9990, (SELECT id FROM cat WHERE slug='moda'), 'Essential', 4.6, 187, 120, false, true, false),
    ('jaqueta-urbanflow', 'Jaqueta Corta-Vento UrbanFlow', 'Jaqueta Corta-Vento UrbanFlow da marca UrbanFlow. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 25990, 34990, (SELECT id FROM cat WHERE slug='moda'), 'UrbanFlow', 4.4, 72, 23, true, false, false),
    ('calca-jeans-denim', 'Calça Jeans Slim Denim Co.', 'Calça Jeans Slim Denim Co. da marca Denim Co.. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 18990, NULL, (SELECT id FROM cat WHERE slug='moda'), 'Denim Co.', 4.1, 58, 41, false, false, false),
    ('vestido-floral', 'Vestido Floral Primavera', 'Vestido Floral Primavera da marca Primavera. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 21990, 27990, (SELECT id FROM cat WHERE slug='moda'), 'Primavera', 4.7, 134, 16, false, false, true),

    -- Casa
    ('luminaria-aurora', 'Luminária de Mesa Aurora', 'Luminária de Mesa Aurora da marca Aurora. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 15990, NULL, (SELECT id FROM cat WHERE slug='casa'), 'Aurora', 4.5, 89, 38, true, false, false),
    ('panelas-chef-master', 'Jogo de Panelas Chef Master', 'Jogo de Panelas Chef Master da marca Chef Master. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 48900, 64900, (SELECT id FROM cat WHERE slug='casa'), 'Chef Master', 4.8, 203, 12, false, true, false),
    ('manta-aconchego', 'Manta Aconchego Soft', 'Manta Aconchego Soft da marca Aconchego. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 11990, 14990, (SELECT id FROM cat WHERE slug='casa'), 'Aconchego', 4.6, 67, 75, false, false, false),
    ('difusor-zenair', 'Difusor de Aromas ZenAir', 'Difusor de Aromas ZenAir da marca ZenAir. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 9990, NULL, (SELECT id FROM cat WHERE slug='casa'), 'ZenAir', 4.3, 51, 60, false, false, true),

    -- Esportes
    ('tenis-speedrun', 'Tênis de Corrida SpeedRun', 'Tênis de Corrida SpeedRun da marca SpeedRun. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 39990, 52990, (SELECT id FROM cat WHERE slug='esportes'), 'SpeedRun', 4.7, 298, 26, true, true, false),
    ('garrafa-hydrofit', 'Garrafa Térmica HydroFit 1L', 'Garrafa Térmica HydroFit 1L da marca HydroFit. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 8990, NULL, (SELECT id FROM cat WHERE slug='esportes'), 'HydroFit', 4.5, 142, 88, false, false, false),
    ('halteres-powerlift', 'Kit Halteres Ajustáveis PowerLift', 'Kit Halteres Ajustáveis PowerLift da marca PowerLift. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 54900, 69900, (SELECT id FROM cat WHERE slug='esportes'), 'PowerLift', 4.8, 76, 9, true, false, false),
    ('mochila-trail-30l', 'Mochila Trail Adventure 30L', 'Mochila Trail Adventure 30L da marca Adventure. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 27990, NULL, (SELECT id FROM cat WHERE slug='esportes'), 'Adventure', 4.4, 63, 34, false, false, true),

    -- Beleza
    ('serum-glow-vitc', 'Sérum Facial Glow Vitamina C', 'Sérum Facial Glow Vitamina C da marca Glow. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 12990, 16990, (SELECT id FROM cat WHERE slug='beleza'), 'Glow', 4.6, 221, 47, false, true, false),
    ('kit-pure-routine', 'Kit Skincare Pure Routine', 'Kit Skincare Pure Routine da marca Pure. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 29990, 39990, (SELECT id FROM cat WHERE slug='beleza'), 'Pure', 4.7, 95, 21, true, false, false),
    ('perfume-noir', 'Perfume Essência Noir 100ml', 'Perfume Essência Noir 100ml da marca Essência. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 35900, NULL, (SELECT id FROM cat WHERE slug='beleza'), 'Essência', 4.5, 118, 15, false, false, false),
    ('secador-iondry', 'Secador Profissional IonDry', 'Secador Profissional IonDry da marca IonDry. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 26990, 32990, (SELECT id FROM cat WHERE slug='beleza'), 'IonDry', 4.3, 54, 28, false, false, true),

    -- Acessórios
    ('oculos-horizon', 'Óculos de Sol Horizon UV400', 'Óculos de Sol Horizon UV400 da marca Horizon. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 19990, 25990, (SELECT id FROM cat WHERE slug='acessorios'), 'Horizon', 4.4, 87, 43, true, false, false),
    ('carteira-heritage', 'Carteira de Couro Heritage', 'Carteira de Couro Heritage da marca Heritage. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 14990, NULL, (SELECT id FROM cat WHERE slug='acessorios'), 'Heritage', 4.6, 109, 56, false, true, false),
    ('mochila-metro-15', 'Mochila Notebook Metro 15"', 'Mochila Notebook Metro 15" da marca Metro. Produto de alta qualidade, ideal para o dia a dia, com acabamento premium e garantia de satisfação. Este é um texto de demonstração para o template de e-commerce.', 22990, 28990, (SELECT id FROM cat WHERE slug='acessorios'), 'Metro', 4.5, 132, 30, false, false, false)
) AS v(slug, name, description, price, price_old, category_id, brand, rating, reviews_count, stock, featured, bestseller, is_new)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    price_old = EXCLUDED.price_old,
    category_id = EXCLUDED.category_id,
    brand = EXCLUDED.brand,
    rating = EXCLUDED.rating,
    reviews_count = EXCLUDED.reviews_count,
    stock = EXCLUDED.stock,
    featured = EXCLUDED.featured,
    bestseller = EXCLUDED.bestseller,
    is_new = EXCLUDED.is_new,
    updated_at = NOW();

-- Insert product variants
WITH prod AS (SELECT id, slug FROM products)
INSERT INTO product_variants (product_id, colors, sizes)
SELECT p.id,
    CASE
        WHEN p.slug IN ('fone-novasound-pro', 'smartwatch-pulse-x2', 'caixa-som-boomgo', 'teclado-keyforge-rgb', 'webcam-streamcam-4k',
                        'luminaria-aurora', 'panelas-chef-master', 'manta-aconchego', 'difusor-zenair',
                        'garrafa-hydrofit', 'halteres-powerlift', 'mochila-trail-30l',
                        'serum-glow-vitc', 'kit-pure-routine', 'perfume-noir', 'secador-iondry',
                        'oculos-horizon', 'carteira-heritage', 'mochila-metro-15')
            THEN ARRAY['#101828', '#465fff', '#f04438', '#12b76a']
        ELSE ARRAY['#101828', '#465fff', '#f04438', '#12b76a', '#fdb022', '#ffffff']
    END,
    CASE
        WHEN p.slug IN ('camiseta-essential', 'jaqueta-urbanflow', 'calca-jeans-denim', 'vestido-floral')
            THEN ARRAY['PP', 'P', 'M', 'G', 'GG']
        WHEN p.slug IN ('tenis-speedrun')
            THEN ARRAY['38', '39', '40', '41', '42', '43']
        ELSE ARRAY[]::TEXT[]
    END
FROM prod p
ON CONFLICT (product_id) DO UPDATE SET
    colors = EXCLUDED.colors,
    sizes = EXCLUDED.sizes;

-- Insert product specs
WITH prod AS (SELECT id, slug, brand FROM products)
INSERT INTO product_specs (product_id, label, value)
SELECT p.id, 'Marca', p.brand FROM prod p
ON CONFLICT DO NOTHING;

WITH prod AS (SELECT id FROM products)
INSERT INTO product_specs (product_id, label, value)
SELECT p.id, 'Garantia', '12 meses' FROM prod p
ON CONFLICT DO NOTHING;

WITH prod AS (SELECT id FROM products)
INSERT INTO product_specs (product_id, label, value)
SELECT p.id, 'Origem', 'Nacional' FROM prod p
ON CONFLICT DO NOTHING;

WITH prod AS (SELECT id, slug FROM products)
INSERT INTO product_specs (product_id, label, value)
SELECT p.id, 'SKU', 'SKU-' || p.id + 1000 FROM prod p
ON CONFLICT DO NOTHING;

-- Insert sample reviews
WITH prod AS (SELECT id, slug FROM products)
INSERT INTO product_reviews (product_id, author, rating, date, comment)
SELECT p.id, 'Marina S.', 5, '2026-04-12', 'Produto excelente, superou minhas expectativas. Recomendo!'
FROM prod p
ON CONFLICT DO NOTHING;

WITH prod AS (SELECT id FROM products)
INSERT INTO product_reviews (product_id, author, rating, date, comment)
SELECT p.id, 'Carlos M.', 4, '2026-03-28', 'Boa qualidade pelo preço. Entrega foi rápida.'
FROM prod p
ON CONFLICT DO NOTHING;

WITH prod AS (SELECT id FROM products)
INSERT INTO product_reviews (product_id, author, rating, date, comment)
SELECT p.id, 'Ana P.', 5, '2026-03-02', 'Amei! Chegou bem embalado e funciona perfeitamente.'
FROM prod p
ON CONFLICT DO NOTHING;

-- Insert demo coupons
INSERT INTO coupons (code, percent, valid_from, active) VALUES
    ('DEMO10', 0.100, NOW(), TRUE),
    ('LOJA20', 0.200, NOW(), TRUE)
ON CONFLICT (code) DO UPDATE SET
    percent = EXCLUDED.percent,
    valid_from = EXCLUDED.valid_from,
    active = EXCLUDED.active;