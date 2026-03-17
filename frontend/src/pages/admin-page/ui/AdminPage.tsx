import { useEffect, useState } from 'react'

export const AdminPage = () => {
  const [users, setUsers] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchAdminData = async () => {
      try {
        const usersResp = await fetch('/api/admin/users').then(r => r.json())
        const statsResp = await fetch('/api/admin/stats').then(r => r.json())
        if (usersResp.status === 'ok') setUsers(usersResp.users)
        setStats(statsResp)
      } catch (err) {
        console.error('Failed to fetch admin data', err)
      } finally {
        setLoading(false)
      }
    }
    fetchAdminData()
  }, [])

  if (loading) return <div className="p-8">Загрузка панели управления...</div>

  return (
    <div className="admin-page p-8" style={{ background: '#fff', minHeight: '100vh', color: '#1f2937' }}>
      <header className="mb-8 flex justify-between items-center">
        <h1 className="text-2xl font-bold">Панель администратора Mudro</h1>
        <div className="flex gap-4">
          <div className="stat-card p-4 border rounded shadow-sm bg-gray-50">
            <div className="text-gray-500 text-sm">Всего пользователей</div>
            <div className="text-xl font-bold">{users.length}</div>
          </div>
          <div className="stat-card p-4 border rounded shadow-sm bg-gray-50">
            <div className="text-gray-500 text-sm">Активные подписки</div>
            <div className="text-xl font-bold">{stats?.active_subscriptions || 0}</div>
          </div>
        </div>
      </header>

      <section className="users-table">
        <h2 className="text-xl font-semibold mb-4">Список пользователей</h2>
        <div className="overflow-x-auto border rounded">
          <table className="w-full text-left">
            <thead className="bg-gray-100 border-b">
              <tr>
                <th className="p-3">ID</th>
                <th className="p-3">Email</th>
                <th className="p-3">Роль</th>
                <th className="p-3">Действия</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-b hover:bg-gray-50">
                  <td className="p-3">{u.id}</td>
                  <td className="p-3">{u.email}</td>
                  <td className="p-3">
                    <span className={`px-2 py-1 rounded text-xs ${u.role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-700'}`}>
                      {u.role}
                    </span>
                  </td>
                  <td className="p-3">
                    <button className="text-blue-600 hover:underline mr-2 text-sm">Сменить роль</button>
                    <button className="text-red-600 hover:underline text-sm">Забанить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  )
}
