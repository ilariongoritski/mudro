import { useGetAdminStatsQuery, useGetAdminUsersQuery } from '@/features/admin/api/adminApi'

import './AdminPage.css'

export const AdminPage = () => {
  const { data: usersResp, isLoading: usersLoading } = useGetAdminUsersQuery()
  const { data: statsResp, isLoading: statsLoading } = useGetAdminStatsQuery()

  const users = usersResp?.status === 'ok' && Array.isArray(usersResp.users) ? usersResp.users : []
  const stats = statsResp ?? null
  const loading = usersLoading || statsLoading

  if (loading) {
    return (
      <div className="admin-page">
        <div className="admin-page__loading">Загружаем панель управления</div>
      </div>
    )
  }

  return (
    <main className="admin-page">
      <header className="admin-page__header">
        <div>
          <h1 className="admin-page__title">Панель администратора</h1>
          <p className="admin-page__subtitle">MUDRO · управление пользователями и системой</p>
        </div>

        <div className="admin-page__stats">
          <div className="admin-stat-card">
            <div className="admin-stat-card__label">Пользователей</div>
            <div className="admin-stat-card__value">{users.length}</div>
          </div>
          <div className="admin-stat-card">
            <div className="admin-stat-card__label">Подписки</div>
            <div className="admin-stat-card__value">{stats?.active_subscriptions ?? 0}</div>
          </div>
          <div className="admin-stat-card">
            <div className="admin-stat-card__label">Админов</div>
            <div className="admin-stat-card__value">
              {users.filter((u) => u.role === 'admin').length}
            </div>
          </div>
        </div>
      </header>

      <section className="admin-page__section">
        <h2 className="admin-page__section-title">Список пользователей</h2>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Логин</th>
                <th>Email</th>
                <th>Роль</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              {users.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', color: 'var(--mudro-muted)', padding: '2rem' }}>
                    Пользователи не найдены
                  </td>
                </tr>
              ) : null}
              {users.map((u) => (
                <tr key={u.id}>
                  <td style={{ opacity: 0.5, fontVariantNumeric: 'tabular-nums' }}>{u.id}</td>
                  <td style={{ fontWeight: 600 }}>{u.username ?? '—'}</td>
                  <td style={{ opacity: 0.8 }}>{u.email}</td>
                  <td>
                    <span className={`admin-badge ${u.role === 'admin' ? 'admin-badge--admin' : 'admin-badge--user'}`}>
                      {u.role}
                    </span>
                  </td>
                  <td>
                    <button type="button" className="admin-btn admin-btn--role">
                      Сменить роль
                    </button>
                    <button type="button" className="admin-btn admin-btn--ban">
                      Забанить
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </main>
  )
}
