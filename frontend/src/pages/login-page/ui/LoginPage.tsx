import React, { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useDispatch } from 'react-redux'
import { useLoginMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
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
        <h1>Вход в Mudro</h1>
        <p className="auth-subtitle">Premium curation platform</p>
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
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="auth-input"
          />
          {error && <div className="auth-error">Ошибка авторизации. Проверьте логин и пароль.</div>}
          <button type="submit" disabled={isLoading} className="auth-button">
            {isLoading ? 'Вход...' : 'Войти'}
          </button>
        </form>
        <div className="auth-footer">
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </div>
      </div>
    </div>
  )
}
