import { Sidebar } from '@/widgets/layout/ui/Sidebar';
import { TopBar } from '@/widgets/layout/ui/TopBar';
import './AdminPage.css';

export const AdminPage = () => {
  // Skeleton user list
  const users = [
    { id: 1, username: 'admin', role: 'admin', email: 'admin@mudro.so' },
    { id: 2, username: 'user', role: 'user', email: 'user@mudro.so' },
  ];

  return (
    <div className="admin-page">
      <Sidebar />
      <div className="admin-page__content">
        <TopBar />
        <main className="admin-page__main">
          <section className="admin-page__card">
            <header className="admin-page__card-header">
              <h2>Пользователи</h2>
              <button className="admin-page__btn-primary">Добавить</button>
            </header>
            <div className="admin-page__table-wrapper">
              <table className="admin-page__table">
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
                  {users.map(user => (
                    <tr key={user.id}>
                      <td>{user.id}</td>
                      <td><strong>{user.username}</strong></td>
                      <td>{user.email}</td>
                      <td><span className={`admin-page__badge admin-page__badge--${user.role}`}>{user.role}</span></td>
                      <td>
                        <button className="admin-page__btn-ghost">Изменить</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </main>
      </div>
    </div>
  );
};
