import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { supabase } from '@/shared/api/supabase'
import { getErrorMessage } from '@/shared/lib/apiError'
import { MudroLogoMark } from '@/shared/ui/MudroLogoMark'

import '@/pages/login-page/ui/Auth.css'

export const RegisterPage = () => {
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      const { error: signUpError } = await supabase.auth.signUp({
        email,
        password,
        options: {
          data: {
            username: login,
          },
        },
      })

      if (signUpError) {
        setError(signUpError.message)
      } else {
        // Typically, without email confirmation, they are logged in automatically.
        // If email confirmation is required, you'd show a success message instead.
        navigate('/', { replace: true })
      }
    } catch (err) {
      console.error('Register failed', err)
      setError('Произошла непредвиденная ошибка.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <Link to="/" className="auth-logo">
          <span className="auth-logo-mark"><MudroLogoMark /></span>
          <span className="auth-logo-text">
            <strong>Mudro</strong>
            <small>Социальная сеть</small>
          </span>
        </Link>
        <h1>Регистрация</h1>
        <p className="auth-subtitle">Создайте аккаунт и сразу войдите в Mudro</p>
        <form onSubmit={handleSubmit} className="auth-form">
          <label htmlFor="reg-login" className="sr-only">Логин</label>
          <input
            id="reg-login"
            type="text"
            placeholder="Логин"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
            autoComplete="username"
            className="auth-input"
          />
          <label htmlFor="reg-email" className="sr-only">Email</label>
          <input
            id="reg-email"
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
            className="auth-input"
          />
          <label htmlFor="reg-password" className="sr-only">Пароль (минимум 6 символов)</label>
          <input
            id="reg-password"
            type="password"
            placeholder="Пароль (мин. 6 символов)"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="new-password"
            className="auth-input"
            minLength={6}
          />
          {error && (
            <div className="auth-error" role="alert" aria-live="assertive">
              {error}
            </div>
          )}
          <button type="submit" disabled={isLoading} className="auth-button">
            {isLoading ? 'Создаём аккаунт...' : 'Зарегистрироваться'}
          </button>
        </form>
        <div className="auth-footer">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </div>
      </div>
    </div>
  )
}
