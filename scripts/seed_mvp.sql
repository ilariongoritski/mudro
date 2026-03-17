INSERT INTO posts (source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, updated_at) VALUES
('tg', 'promo-1', NOW() - INTERVAL '1 hour', 'Добро пожаловать в Mudro — платформу нового поколения! 🔥\nМы создали этот MVP, чтобы показать, как мощно может выглядеть ваша личная лента. Наслаждайтесь Glassmorphism-дизайном и плавными анимациями.', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 124, 1500, 12, NOW()),

('tg', 'promo-2', NOW() - INTERVAL '3 hours', 'Темная тема (Hotpink Dark) создает по-настоящему премиальный опыт. Каждая карточка поста имеет эффект матового стекла (backdrop-filter) и мягкую неоновую подсветку при наведении. ✨', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1550684848-fac1c5b4e853?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 89, 840, 5, NOW()),

('vk', 'promo-3', NOW() - INTERVAL '5 hours', 'База данных PostgreSQL успешно развернута, API-шлюз на Nginx работает без 404 ошибок, а фронтенд на Vercel безупречно проксирует все запросы. Это полностью рабочий Fullstack MVP! 🚀', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1550745165-9bc0b252726f?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 256, 3100, 42, NOW())
ON CONFLICT (source, source_post_id) DO NOTHING;

INSERT INTO post_reactions (post_id, emoji, count)
SELECT id, '❤️', likes_count FROM posts WHERE source_post_id IN ('promo-1', 'promo-2', 'promo-3')
ON CONFLICT DO NOTHING;
