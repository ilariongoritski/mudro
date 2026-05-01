import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRegisterMutation, useLoginMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import { useAppDispatch } from '@/shared/lib/hooks/storeHooks'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader } from '@/shared/ui/card'
import { Input } from '@/shared/ui/input'

interface ApiErrorShape {
  data?: {
    error?: string
  }
}

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true)
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const [register, { isLoading: isRegistering }] = useRegisterMutation()
  const [login, { isLoading: isLoggingIn }] = useLoginMutation()

  const isLoading = isRegistering || isLoggingIn

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    try {
      const result = isLogin
        ? await login({ email, password }).unwrap()
        : await register({ username, email, password }).unwrap()

      dispatch(setCredentials(result))
      navigate('/')
    } catch (err: unknown) {
      const msg = (err as ApiErrorShape)?.data?.error || 'Something went wrong'
      setError(msg)
    }
  }

  return (
    <div className="max-w-md mx-auto mt-12 p-4">
      <Card>
        <CardHeader>
          <h2 className="text-xl font-bold text-center">
            {isLogin ? 'Вход' : 'Регистрация'}
          </h2>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {!isLogin && (
              <Input
                placeholder="Имя пользователя"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            )}
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            <Input
              type="password"
              placeholder="Пароль"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
            {error && <p className="text-sm text-red-500">{error}</p>}
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Загрузка...' : isLogin ? 'Войти' : 'Зарегистрироваться'}
            </Button>
          </form>
          <button
            type="button"
            onClick={() => { setIsLogin(!isLogin); setError('') }}
            className="mt-4 text-sm text-mudro-muted hover:text-mudro-text w-full text-center transition-colors"
          >
            {isLogin ? 'Нет аккаунта? Зарегистрироваться' : 'Уже есть аккаунт? Войти'}
          </button>
        </CardContent>
      </Card>
    </div>
  )
}
