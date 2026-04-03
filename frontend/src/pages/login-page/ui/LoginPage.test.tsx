import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { Provider } from 'react-redux'
import { configureStore } from '@reduxjs/toolkit'

import { LoginPage } from './LoginPage'

// Мокируем API мутацию
const mockLoginMutation = vi.fn()
const mockUnwrap = vi.fn()

vi.mock('@/entities/session/api/authApi', () => ({
  useLoginMutation: () => [
    mockLoginMutation.mockReturnValue({ unwrap: mockUnwrap }),
    { isLoading: false, error: null },
  ],
}))

vi.mock('@/entities/session/model/sessionSlice', () => ({
  setCredentials: vi.fn((payload) => ({ type: 'session/setCredentials', payload })),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  }
})

const mockStore = configureStore({
  reducer: {
    session: (state = { token: null, user: null }) => state,
  },
})

const renderLoginPage = () =>
  render(
    <Provider store={mockStore}>
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    </Provider>,
  )

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUnwrap.mockResolvedValue({ token: 'test-token', user: { id: 1 } })
  })

  it('рендерит заголовок формы', () => {
    renderLoginPage()
    expect(screen.getByRole('heading', { name: 'Вход' })).toBeInTheDocument()
  })

  it('поле логина имеет связанный label', () => {
    renderLoginPage()
    const input = screen.getByLabelText('Логин или email')
    expect(input).toBeInTheDocument()
    expect(input).toHaveAttribute('type', 'text')
  })

  it('поле пароля имеет связанный label', () => {
    renderLoginPage()
    const input = screen.getByLabelText('Пароль')
    expect(input).toBeInTheDocument()
    expect(input).toHaveAttribute('type', 'password')
  })

  it('оба поля доступны для ввода', () => {
    renderLoginPage()
    const loginInput = screen.getByLabelText('Логин или email')
    const passwordInput = screen.getByLabelText('Пароль')

    fireEvent.change(loginInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })

    expect(loginInput).toHaveValue('testuser')
    expect(passwordInput).toHaveValue('password123')
  })

  it('кнопка отправки присутствует', () => {
    renderLoginPage()
    expect(screen.getByRole('button', { name: 'Войти' })).toBeInTheDocument()
  })

  it('ссылка на регистрацию присутствует', () => {
    renderLoginPage()
    expect(screen.getByRole('link', { name: 'Зарегистрироваться' })).toBeInTheDocument()
  })

  it('поле логина имеет autocomplete=username', () => {
    renderLoginPage()
    expect(screen.getByLabelText('Логин или email')).toHaveAttribute('autocomplete', 'username')
  })

  it('поле пароля имеет autocomplete=current-password', () => {
    renderLoginPage()
    expect(screen.getByLabelText('Пароль')).toHaveAttribute('autocomplete', 'current-password')
  })
})
