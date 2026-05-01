import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import type { SerializedError } from '@reduxjs/toolkit'
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query'
import { useRegisterMutation } from '@/entities/session/api/authApi'
import { useAppDispatch } from '@/shared/lib/hooks/storeHooks'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import { getErrorMessage } from '@/shared/lib/apiError'
import { MudroLogoMark } from '@/shared/ui/MudroLogoMark'

import '@/pages/login-page/ui/Auth.css'

export const RegisterPage = () => {
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [registerMutation, { isLoading }] = useRegisterMutation()
  const dispatch = useAppDispatch()
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    try {
      const result = await registerMutation({
        username: login,
        email: email || undefined,
        password,
      }).unwrap()

      dispatch(setCredentials(result))
      navigate('/', { replace: true })
    } catch (err) {
      console.error('Register failed', err)
      setError(getErrorMessage(err as FetchBaseQueryError | SerializedError | undefined, 'Произошла ошибка при регистрации.'))
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
