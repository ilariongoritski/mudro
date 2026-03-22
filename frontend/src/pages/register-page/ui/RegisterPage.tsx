import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'

import { useRegisterMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'

import '@/pages/login-page/ui/Auth.css'

export const RegisterPage = () => {
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [register, { isLoading, error }] = useRegisterMutation()
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const result = await register({ login, email, password }).unwrap()
      dispatch(setCredentials(result))
      navigate('/', { replace: true })
    } catch (err) {
      console.error('Register failed', err)
    }
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>Регистрация</h1>
        <p className="auth-subtitle">Создайте аккаунт и сразу войдите в Mudro</p>
        <form onSubmit={handleSubmit} className="auth-form">
          <input
            type="text"
            placeholder="Логин"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
            className="auth-input"
          />
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="auth-input"
          />
          <input
            type="password"
            placeholder="Пароль (мин. 6 символов)"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="auth-input"
            minLength={6}
          />
          {error && <div className="auth-error">Ошибка регистрации. Возможно, логин или email уже заняты.</div>}
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
