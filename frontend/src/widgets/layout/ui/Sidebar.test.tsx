import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'

import { Sidebar } from './Sidebar'

vi.mock('@/shared/lib/hooks/storeHooks', () => ({
  useAppDispatch: () => vi.fn(),
  useAppSelector: (selector: (s: unknown) => unknown) =>
    selector({ session: { token: null, user: null } }),
}))

vi.mock('@/entities/session/model/sessionSlice', () => ({
  logout: vi.fn(() => ({ type: 'session/logout' })),
}))

const renderSidebar = () =>
  render(
    <MemoryRouter>
      <Sidebar />
    </MemoryRouter>,
  )

describe('Sidebar', () => {
  it('рендерит логотип', () => {
    renderSidebar()
    expect(screen.getByText('Mudro')).toBeInTheDocument()
  })

  it('рендерит все навигационные ссылки', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: /Лента/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Мессенджер/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Казино/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Контур/ })).toBeInTheDocument()
  })

  it('ссылка Лента ведёт на /', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: /Лента/ })).toHaveAttribute('href', '/')
  })

  it('ссылка Мессенджер ведёт на /chat', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: /Мессенджер/ })).toHaveAttribute('href', '/chat')
  })

  it('без авторизации: показывает Войти и Регистрация', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: /Войти/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Регистрация/ })).toBeInTheDocument()
  })

  it('без авторизации: нет кнопки выхода', () => {
    renderSidebar()
    expect(screen.queryByRole('button', { name: /Выйти/i })).not.toBeInTheDocument()
  })

  it('nav содержится в aside', () => {
    renderSidebar()
    expect(screen.getByRole('complementary')).toBeInTheDocument()
    expect(screen.getByRole('navigation')).toBeInTheDocument()
  })
})
