-- ===========================================================
-- Seed: threaded comments for MVP promo posts
-- ===========================================================

DO $$ 
DECLARE
    post1_id bigint;
    post3_id bigint;
    root1_id bigint;
BEGIN
    SELECT id INTO post1_id FROM posts WHERE source_post_id = 'promo-1';
    SELECT id INTO post3_id FROM posts WHERE source_post_id = 'promo-3';

    -- Feedbacks for promo-1
    IF post1_id IS NOT NULL THEN
        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text, reactions)
        VALUES (post1_id, 'tg', 'tg-c1', 'Alex Cyber', NOW() - INTERVAL '40 minutes', 'Вау! Выглядит потрясающе, особенно dark theme. Это именно то, что я ждал от MUDRO.', '[{"label":"🔥", "count": 15}]'::jsonb)
        ON CONFLICT (source, source_comment_id) DO NOTHING RETURNING id INTO root1_id;

        IF root1_id IS NULL THEN
            SELECT id INTO root1_id FROM post_comments WHERE source_comment_id = 'tg-c1';
        END IF;

        IF root1_id IS NOT NULL THEN
            INSERT INTO post_comments (post_id, source, source_comment_id, parent_comment_id, author_name, published_at, text, reactions)
            VALUES (post1_id, 'tg', 'tg-c1-r1', root1_id, 'MUDRO Team', NOW() - INTERVAL '30 minutes', 'Спасибо, Alex! Мы старались сделать фокус на UX и плавности анимации.', '[{"label":"❤️", "count": 5}]'::jsonb)
            ON CONFLICT (source, source_comment_id) DO NOTHING;
        END IF;

        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text)
        VALUES (post1_id, 'tg', 'tg-c2', 'Tech Reviewer', NOW() - INTERVAL '15 minutes', 'Блюр и glassmorphism отлично смотрятся на десктопе. Главное чтобы на мобилках не лагало.')
        ON CONFLICT (source, source_comment_id) DO NOTHING;
    END IF;

    -- Feedbacks for promo-3
    IF post3_id IS NOT NULL THEN
        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text, reactions)
        VALUES (post3_id, 'vk', 'vk-c1', 'Dmitry DevOps', NOW() - INTERVAL '2 hours', 'Отлично, что API и база работают прямо из коробки на MVP. Архитектура solid. 👍', '[{"raw":"👍", "count": 22}]'::jsonb)
        ON CONFLICT (source, source_comment_id) DO NOTHING;
    END IF;
END $$;
