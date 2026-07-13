import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { User, Mail, Shield, Calendar, Crown, ArrowRight } from 'lucide-react'
import { useMeQuery } from '@/entities/session/api/authApi'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { Card, CardContent } from '@/shared/ui/card'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { Skeleton } from '@/shared/ui/Skeleton'

import './ProfilePage.css'

const buildInitials = (name: string) =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

const formatDate = (iso: string | undefined) => {
  if (!iso) return '—'
  try {
    return new Intl.DateTimeFormat('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    }).format(new Date(iso))
  } catch {
    return '—'
  }
}

const roleLabel: Record<string, string> = {
  admin: 'Администратор',
  user: 'Пользователь',
  moderator: 'Модератор',
}

export const ProfilePage = () => {
  const { user: sessionUser } = useAppSelector((state) => state.session)
  const { data: profile, isLoading } = useMeQuery(undefined, {
    skip: !sessionUser,
  })

  const user = profile ?? sessionUser

  const initials = useMemo(() => buildInitials(user?.username ?? 'M'), [user?.username])
  const displayName = user?.username ?? 'Гость'
  const email = user?.email
  const role = user?.role ?? 'user'
  const isPremium = user?.isPremium ?? false
  const createdAt = profile && 'created_at' in profile ? (profile as Record<string, string>).created_at : undefined

  if (isLoading) {
    return (
      <div className="profile-page">
        <div className="profile-page__header">
          <h1 className="profile-page__title">Профиль</h1>
          <p className="profile-page__sub">Загрузка данных...</p>
        </div>
        <Card>
          <CardContent className="profile-page__loading">
            <Skeleton type="circle" width={80} height={80} />
            <Skeleton type="title" width={200} />
            <Skeleton type="text" width={160} />
            <Skeleton type="text" width={120} />
            <Skeleton type="rect" height={80} className="mt-4" />
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="profile-page">
      <div className="profile-page__header">
        <div className="profile-page__header-left">
          <h1 className="profile-page__title">Профиль</h1>
          <p className="profile-page__sub">Учётная запись и настройки</p>
        </div>
        <Badge variant={isPremium ? 'accent' : 'default'}>
          {isPremium ? 'Премиум' : roleLabel[role] ?? role}
        </Badge>
      </div>

      <div className="profile-page__grid">
        {/* Основная карточка */}
        <Card className="profile-page__main-card">
          <CardContent className="profile-page__main-content">
            <div className="profile-page__avatar" aria-hidden="true">
              {initials || '👤'}
            </div>
            <div className="profile-page__info">
              <h2 className="profile-page__display-name">{displayName}</h2>
              {email && (
                <div className="profile-page__detail">
                  <Mail size={16} />
                  <span>{email}</span>
                </div>
              )}
              <div className="profile-page__detail">
                <User size={16} />
                <span>@{displayName}</span>
              </div>
              <div className="profile-page__detail">
                <Shield size={16} />
                <span>{roleLabel[role] ?? role}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Статистика */}
        <Card className="profile-page__stats-card">
          <CardContent className="profile-page__stats-content">
            <div className="profile-page__stat">
              <div className="profile-page__stat-icon">
                <Crown size={20} />
              </div>
              <div className="profile-page__stat-body">
                <span className="profile-page__stat-value">{isPremium ? 'Активен' : 'Стандарт'}</span>
                <span className="profile-page__stat-label">Статус</span>
              </div>
            </div>
            <div className="profile-page__stat">
              <div className="profile-page__stat-icon">
                <Calendar size={20} />
              </div>
              <div className="profile-page__stat-body">
                <span className="profile-page__stat-value">{formatDate(createdAt)}</span>
                <span className="profile-page__stat-label">На сайте с</span>
              </div>
            </div>
            <div className="profile-page__stat">
              <div className="profile-page__stat-icon">
                <User size={20} />
              </div>
              <div className="profile-page__stat-body">
                <span className="profile-page__stat-value">Telegram</span>
                <span className="profile-page__stat-label">Связь</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Действия */}
      <Card className="profile-page__actions-card">
        <CardContent className="profile-page__actions-content">
          <Button variant="outline" className="profile-page__action" disabled>
            Редактировать профиль
            <ArrowRight size={16} />
          </Button>
          <Button variant="ghost" className="profile-page__action" asChild>
            <Link to="/casino">
              Перейти в казино
              <ArrowRight size={16} />
            </Link>
          </Button>
        </CardContent>
      </Card>

      {/* Информация о безопасности */}
      <Card>
        <CardContent className="profile-page__security">
          <h3 className="profile-page__security-title">Безопасность</h3>
          <p className="profile-page__security-text">
            Сессия активна. Для смены пароля или настройки двухфакторной аутентификации
            обратитесь в поддержку.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

export default ProfilePage