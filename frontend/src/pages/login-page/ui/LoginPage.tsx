import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'

import { useLoginMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import { getErrorMessage } from '@/shared/lib/apiError'
import { MudroLogoMark } from '@/shared/ui/MudroLogoMark'

import '@/pages/login-page/ui/Auth.css'

export const LoginPage = () => {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [loginMutation, { isLoading, error }] = useLoginMutation()
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const result = await loginMutation({ login, password }).unwrap()
      dispatch(setCredentials(result))
      navigate('/')
    } catch (err) {
      console.error('Login failed', err)
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
        <h1>Вход</h1>
        <p className="auth-subtitle">Войдите, чтобы получить доступ к мессенджеру и казино</p>
        <form onSubmit={handleSubmit} className="auth-form">
          <input
            type="text"
            placeholder="Логин или email"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
            className="auth-input"
          />
          <input
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="auth-input"
          />
          {error && <div className="auth-error">{getErrorMessage(error, 'Неверный логин или пароль.')}</div>}
          <button type="submit" disabled={isLoading} className="auth-button">
            {isLoading ? 'Входим...' : 'Войти'}
          </button>
        </form>
        <div className="auth-footer">
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </div>
      </div>
    </div>
  )
}
