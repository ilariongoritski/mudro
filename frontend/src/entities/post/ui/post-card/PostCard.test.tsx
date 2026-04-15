import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { PostCard } from './PostCard'
import type { Post } from '@/entities/post/model/types'

const mockPost: Post = {
  id: 1,
  source: 'tg',
  source_post_id: 'tg_123',
  published_at: '2024-01-15T10:00:00Z',
  text: 'Тестовый пост с интересным содержимым для проверки рендеринга карточки',
  likes_count: 42,
  views_count: 1500,
  comments_count: 7,
  reactions: { '❤️': 10, '👍': 5 },
  media: [],
  comments: [],
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:00:00Z',
}

describe('PostCard', () => {
  it('рендерит текст поста', () => {
    render(<PostCard post={mockPost} />)
    expect(screen.getByText(/Тестовый пост/)).toBeInTheDocument()
  })

  it('показывает источник поста', () => {
    render(<PostCard post={mockPost} />)
    expect(screen.getByText('Мудро (тг)')).toBeInTheDocument()
  })

  it('показывает метрики с aria-label', () => {
    render(<PostCard post={mockPost} />)
    expect(screen.getByLabelText(/лайков/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/комментариев/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/просмотров/i)).toBeInTheDocument()
  })

  it('без onOpen — не является кнопкой', () => {
    render(<PostCard post={mockPost} />)
    const article = screen.queryByRole('button', { name: /Открыть пост/ })
    expect(article).not.toBeInTheDocument()
  })

  it('с onOpen — карточка получает role=button и tabIndex', () => {
    const onOpen = vi.fn()
    render(<PostCard post={mockPost} onOpen={onOpen} />)
    const btn = screen.getByRole('button', { name: /Открыть пост/ })
    expect(btn).toBeInTheDocument()
    expect(btn).toHaveAttribute('tabindex', '0')
  })

  it('клик по карточке вызывает onOpen', () => {
    const onOpen = vi.fn()
    render(<PostCard post={mockPost} onOpen={onOpen} />)
    fireEvent.click(screen.getByRole('button', { name: /Открыть пост/ }))
    expect(onOpen).toHaveBeenCalledWith(mockPost)
    expect(onOpen).toHaveBeenCalledTimes(1)
  })

  it('Enter открывает пост', () => {
    const onOpen = vi.fn()
    render(<PostCard post={mockPost} onOpen={onOpen} />)
    const btn = screen.getByRole('button', { name: /Открыть пост/ })
    fireEvent.keyDown(btn, { key: 'Enter' })
    expect(onOpen).toHaveBeenCalledWith(mockPost)
  })

  it('Space открывает пост', () => {
    const onOpen = vi.fn()
    render(<PostCard post={mockPost} onOpen={onOpen} />)
    const btn = screen.getByRole('button', { name: /Открыть пост/ })
    fireEvent.keyDown(btn, { key: ' ' })
    expect(onOpen).toHaveBeenCalledWith(mockPost)
  })

  it('показывает VK как источник', () => {
    render(<PostCard post={{ ...mockPost, source: 'vk' }} />)
    expect(screen.getByText('Мудро (вк)')).toBeInTheDocument()
  })

  it('показывает фолбэк текст при отсутствии text', () => {
    render(<PostCard post={{ ...mockPost, text: null }} />)
    expect(
      screen.getByText(/Описание для этого поста пока не подтянулось/),
    ).toBeInTheDocument()
  })

  it('показывает реакции', () => {
    render(<PostCard post={mockPost} />)
    expect(screen.getByTitle('❤️')).toBeInTheDocument()
  })

  it('показывает оверлей +N медиа с aria-label', () => {
    const postWithMedia: Post = {
      ...mockPost,
      media: [
        { kind: 'photo', url: 'https://example.com/1.jpg', is_image: true },
        { kind: 'photo', url: 'https://example.com/2.jpg', is_image: true },
        { kind: 'photo', url: 'https://example.com/3.jpg', is_image: true },
        { kind: 'photo', url: 'https://example.com/4.jpg', is_image: true },
      ],
    }
    render(<PostCard post={postWithMedia} />)
    const overlay = screen.getByLabelText(/Ещё 1 медиа/)
    expect(overlay).toBeInTheDocument()
  })
})
