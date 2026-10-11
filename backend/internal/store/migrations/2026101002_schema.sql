
INSERT INTO products (name, description, price_cents, stock)
SELECT seed.name, seed.description, seed.price_cents, seed.stock
FROM (
    VALUES
        ('Wireless Headphones', 'Premium noise-cancelling wireless headphones with 30-hour battery life', 29999::BIGINT, 50::BIGINT),
        ('Mechanical Keyboard', 'RGB backlit mechanical keyboard with Cherry MX Brown switches', 14999::BIGINT, 30::BIGINT),
        ('USB-C Hub', '7-in-1 USB-C hub with HDMI, USB-A, SD card reader, and PD charging', 4999::BIGINT, 100::BIGINT),
        ('Ergonomic Mouse', 'Vertical ergonomic mouse with adjustable DPI and wireless connectivity', 3999::BIGINT, 75::BIGINT),
        ('Laptop Stand', 'Aluminum adjustable laptop stand with cable management', 2999::BIGINT, 60::BIGINT),
        ('Webcam Cover', 'Magnetic sliding webcam cover for privacy (pack of 3)', 999::BIGINT, 200::BIGINT)
) AS seed(name, description, price_cents, stock)
WHERE NOT EXISTS (
    SELECT 1
    FROM products existing
    WHERE existing.name = seed.name
);
