import {
  useGetAdminStatsQuery,
  useGetAdminUsersQuery,
  useGetRuntimeDashboardQuery,
} from '@/features/admin/api/adminApi'
import { getErrorMessage } from '@/shared/lib/apiError'

import './AdminPage.css'

const statusLabel: Record<string, string> = {
  healthy: 'healthy',
  unavailable: 'unavailable',
  unknown: 'unknown',
}

export const AdminPage = () => {
  const {
    data: usersData,
    isLoading: isUsersLoading,
    isFetching: isUsersFetching,
    error: usersError,
  } = useGetAdminUsersQuery()
  const { data: stats, isLoading: isStatsLoading, error: statsError } = useGetAdminStatsQuery()
  const { data: runtime, isLoading: isRuntimeLoading, error: runtimeError } = useGetRuntimeDashboardQuery()

  const users = usersData?.users ?? []
  const isLoading = isUsersLoading || isStatsLoading
  const error = usersError ?? statsError
  const totalUsers = stats?.total_users ?? users.length
  const activeSubscriptions = stats?.active_subscriptions ?? 0

  return (
    <main className="admin-page__main">
      <header className="admin-page__heading">
        <div>
          <p className="admin-page__eyebrow">Admin only</p>
          <h1>Управление сервером</h1>
          <p>Только безопасные метаданные. Значения ключей не отображаются и не передаются клиенту.</p>
        </div>
      </header>

      <section className="admin-page__stats" aria-label="Админская сводка">
        <article className="admin-page__stat-card">
          <span>Пользователи</span>
          <strong>{isStatsLoading ? '...' : totalUsers}</strong>
        </article>
        <article className="admin-page__stat-card">
          <span>Активные подписки</span>
          <strong>{isStatsLoading ? '...' : activeSubscriptions}</strong>
        </article>
        <article className="admin-page__stat-card">
          <span>Сервисы healthy</span>
          <strong>{isRuntimeLoading ? '...' : runtime?.services.filter((service) => service.status === 'healthy').length ?? 0}</strong>
        </article>
      </section>

      <section className="admin-page__card">
        <header className="admin-page__card-header">
          <div>
            <h2>Провайдеры, модели и лимиты</h2>
            <p>Состояние берётся только из разрешённых metadata-переменных окружения.</p>
          </div>
        </header>

        {isRuntimeLoading && <div className="admin-page__state" role="status">Загрузка runtime-статуса...</div>}
        {!isRuntimeLoading && runtimeError && (
          <div className="admin-page__state admin-page__state--error" role="alert">
            {getErrorMessage(runtimeError, 'Не удалось загрузить runtime-статус.')}
          </div>
        )}
        {!isRuntimeLoading && !runtimeError && runtime && (
          <>
            <div className="admin-page__provider-grid">
              {runtime.providers.map((provider) => (
                <article className="admin-page__provider" key={provider.name}>
                  <div className="admin-page__provider-title">
                    <h3>{provider.name}</h3>
                    <span className={`admin-page__status admin-page__status--${provider.configured ? 'healthy' : 'unavailable'}`}>
                      {provider.configured ? 'configured' : 'not configured'}
                    </span>
                  </div>
                  <dl>
                    <div><dt>Модель</dt><dd>{provider.model || 'не задана'}</dd></div>
                    <div><dt>Лимит</dt><dd>{provider.limit || 'не задан'}</dd></div>
                  </dl>
                </article>
              ))}
            </div>
            <div className="admin-page__limits">
              <span>API RPS: <strong>{runtime.limits.requests_per_second || 'не задан'}</strong></span>
              <span>API burst: <strong>{runtime.limits.burst || 'не задан'}</strong></span>
            </div>
            <div className="admin-page__service-list" aria-label="Health сервисов">
              {runtime.services.map((service) => (
                <div className="admin-page__service" key={service.name}>
                  <span>{service.name}</span>
                  <span className={`admin-page__status admin-page__status--${service.status}`}>{statusLabel[service.status]}</span>
                </div>
              ))}
            </div>
          </>
        )}
      </section>

      <section className="admin-page__card">
        <header className="admin-page__card-header">
          <h2>Пользователи</h2>
          {isUsersFetching && !isUsersLoading && <span className="admin-page__refresh">Обновление...</span>}
        </header>

        {isLoading && <div className="admin-page__state" role="status">Загрузка админских данных...</div>}
        {!isLoading && error && <div className="admin-page__state admin-page__state--error" role="alert">{getErrorMessage(error, 'Не удалось загрузить админские данные.')}</div>}
        {!isLoading && !error && users.length === 0 && <div className="admin-page__state">Пользователей пока нет.</div>}
        {!isLoading && !error && users.length > 0 && (
          <div className="admin-page__table-wrapper">
            <table className="admin-page__table">
              <thead><tr><th>ID</th><th>Логин</th><th>Email</th><th>Роль</th></tr></thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td><td><strong>{user.username ?? 'Без логина'}</strong></td><td>{user.email ?? '-'}</td>
                    <td><span className={`admin-page__badge admin-page__badge--${user.role}`}>{user.role}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  )
}
