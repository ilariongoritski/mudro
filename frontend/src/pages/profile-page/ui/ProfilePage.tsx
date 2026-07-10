import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useGetProfileQuery, useGetCasinoStatsQuery, useGetActivitiesQuery, useUpdateProfileMutation, useStartMessageMutation } from '@/entities/profile/api/profileApi';
import { ProfileForm } from '@/shared/ui/ProfileForm/ProfileForm';
import { AvatarUpload } from '@/shared/ui/AvatarUpload/AvatarUpload';

const ProfilePage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const userId = id ? parseInt(id) : 1; // fallback to self
  const isOwn = true; // TODO: compare with current user

  const { data: profile, isLoading: profileLoading } = useGetProfileQuery(userId);
  const { data: casino } = useGetCasinoStatsQuery(userId);
  const { data: activities } = useGetActivitiesQuery(userId);
  const [updateProfile, { isLoading: updating }] = useUpdateProfileMutation();
  const [startMessage] = useStartMessageMutation();

  const [editMode, setEditMode] = useState(false);

  const handleUpdate = async (data: any) => {
    await updateProfile({ userId, data }).unwrap();
    setEditMode(false);
  };

  const handleAvatar = async (file: File) => {
    const formData = new FormData();
    formData.append('avatar', file);
    // await uploadAvatar(formData).unwrap();
    // TODO: real upload
  };

  const handleMessage = async () => {
    const res = await startMessage(userId).unwrap();
    // navigate to chat with res.chat_id
    alert('Чат открыт: ' + res.chat_id);
  };

  if (profileLoading) return <div className="p-8 text-center">Загрузка профиля...</div>;
  if (!profile) return <div>Профиль не найден</div>;

  return (
    <div className="max-w-4xl mx-auto p-4 md:p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="w-20 h-20 rounded-full overflow-hidden border">
            {profile.avatar_url ? (
              <img src={profile.avatar_url} alt="Avatar" className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full bg-gray-200 flex items-center justify-center text-3xl">👤</div>
            )}
          </div>
          <div>
            <h1 className="text-3xl font-bold">{profile.display_name}</h1>
            <p className="text-gray-500">@{profile.username}</p>
            <div className="flex gap-2 mt-1">
              <span className="px-2 py-0.5 bg-green-100 text-green-700 rounded text-xs">Рейтинг: {profile.rating}</span>
              <span className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded text-xs">Заполнено: {profile.profile_completion}%</span>
            </div>
          </div>
        </div>

        {isOwn && (
          <div className="flex gap-2">
            <button onClick={() => setEditMode(!editMode)} className="px-4 py-2 border rounded hover:bg-gray-50">
              {editMode ? 'Отмена' : 'Редактировать'}
            </button>
            <button onClick={handleMessage} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
              Написать
            </button>
          </div>
        )}
      </div>

      {/* Avatar Upload (own profile) */}
      {isOwn && <AvatarUpload currentAvatar={profile.avatar_url} onUpload={handleAvatar} />}

      {/* Edit Form */}
      {editMode && isOwn && (
        <div className="border rounded p-6 bg-white">
          <h3 className="font-semibold mb-4">Редактирование профиля</h3>
          <ProfileForm
            initialData={profile}
            onSubmit={handleUpdate}
            isLoading={updating}
          />
        </div>
      )}

      {/* Main Info */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="border rounded p-6 bg-white">
          <h3 className="font-semibold mb-4">О себе</h3>
          <p className="text-gray-700 whitespace-pre-line">{profile.bio || 'Информация не заполнена'}</p>
          {profile.age && <p className="mt-2 text-sm text-gray-500">Возраст: {profile.age}</p>}
          {profile.social_links && Object.keys(profile.social_links).length > 0 && (
            <div className="mt-3">
              <div className="text-sm font-medium mb-1">Социальные сети</div>
              {Object.entries(profile.social_links).map(([k, v]) => (
                <a key={k} href={v} target="_blank" className="block text-blue-600 hover:underline text-sm">{k}: {v}</a>
              ))}
            </div>
          )}
        </div>

        {/* Casino Stats */}
        <div className="border rounded p-6 bg-white">
          <h3 className="font-semibold mb-4">🎰 Казино</h3>
          {casino ? (
            <div className="space-y-2 text-sm">
              <div>Баланс: <span className="font-mono font-semibold">{casino.balance} МДР</span></div>
              <div>Игр сыграно: {casino.games_count}</div>
              <div>Макс. выигрыш: <span className="font-mono">{casino.max_win} МДР</span></div>
            </div>
          ) : (
            <p className="text-gray-500">Статистика казино недоступна</p>
          )}
        </div>
      </div>

      {/* Activities */}
      <div className="border rounded p-6 bg-white">
        <h3 className="font-semibold mb-4">Активность</h3>
        {activities && activities.length > 0 ? (
          <ul className="space-y-2 text-sm">
            {activities.map((act) => (
              <li key={act.id} className="flex justify-between border-b pb-2">
                <span>{act.type}</span>
                <span className="text-gray-500">{new Date(act.created_at).toLocaleDateString()}</span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">Пока нет активности</p>
        )}
      </div>

      {/* Telegram link */}
      {profile.telegram_username && (
        <div className="text-sm text-gray-500">
          Telegram: @{profile.telegram_username}
        </div>
      )}
    </div>
  );
};

export default ProfilePage;
