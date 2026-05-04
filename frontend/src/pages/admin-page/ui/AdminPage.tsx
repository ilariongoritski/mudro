import { useGetAdminStatsQuery, useGetAdminUsersQuery } from '@/features/admin/api/adminApi'
import { getErrorMessage } from '@/shared/lib/apiError'

import './AdminPage.css'

export const AdminPage = () => {
  const {
    data: usersData,
    isLoading: isUsersLoading,
    isFetching: isUsersFetching,
    error: usersError,
  } = useGetAdminUsersQuery()
  const {
    data: stats,
    isLoading: isStatsLoading,
    error: statsError,
  } = useGetAdminStatsQuery()

  const users = usersData?.users ?? []
  const isLoading = isUsersLoading || isStatsLoading
  const error = usersError ?? statsError
  const totalUsers = (stats as { total_users?: number } | undefined)?.total_users ?? users.length
  const activeSubscriptions = stats?.active_subscriptions ?? 0

  return (
    <main className="admin-page__main">
      <section className="admin-page__stats" aria-label="Админская сводка">
        <article className="admin-page__stat-card">
          <span>Пользователи</span>
          <strong>{isStatsLoading ? '...' : totalUsers}</strong>
        </article>
        <article className="admin-page__stat-card">
          <span>Активные подписки</span>
          <strong>{isStatsLoading ? '...' : activeSubscriptions}</strong>
        </article>
      </section>

      <section className="admin-page__card">
        <header className="admin-page__card-header">
          <h2>Пользователи</h2>
          {isUsersFetching && !isUsersLoading && <span className="admin-page__refresh">Обновление...</span>}
        </header>

        {isLoading && (
          <div className="admin-page__state" role="status">
            Загрузка админских данных...
          </div>
        )}

        {!isLoading && error && (
          <div className="admin-page__state admin-page__state--error" role="alert">
            {getErrorMessage(error, 'Не удалось загрузить админские данные.')}
          </div>
        )}

        {!isLoading && !error && users.length === 0 && (
          <div className="admin-page__state">
            Пользователей пока нет.
          </div>
        )}

        {!isLoading && !error && users.length > 0 && (
          <div className="admin-page__table-wrapper">
            <table className="admin-page__table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Логин</th>
                  <th>Email</th>
                  <th>Роль</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td>
                    <td><strong>{user.username ?? 'Без логина'}</strong></td>
                    <td>{user.email ?? '-'}</td>
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
