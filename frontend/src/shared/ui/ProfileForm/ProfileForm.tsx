import React, { useState } from 'react';
import type { ProfileUpdate } from '@/entities/profile/model/profile.types';

interface ProfileFormProps {
  initialData?: Partial<ProfileUpdate & { display_name: string; username: string; age?: number | null; bio?: string | null }>;
  onSubmit: (data: ProfileUpdate) => void;
  isLoading?: boolean;
}

export const ProfileForm: React.FC<ProfileFormProps> = ({ initialData = {}, onSubmit, isLoading }) => {
  const [form, setForm] = useState({
    display_name: initialData.display_name || '',
    username: initialData.username || '',
    email: initialData.email || '',
    age: initialData.age || '',
    bio: initialData.bio || '',
    social_links: initialData.social_links || {},
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setForm(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload: ProfileUpdate = {
      display_name: form.display_name,
      username: form.username,
      email: form.email || null,
      age: form.age ? Number(form.age) : null,
      bio: form.bio || null,
      social_links: Object.keys(form.social_links).length > 0 ? form.social_links : undefined,
    };
    onSubmit(payload);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">Имя (display_name)</label>
        <input
          type="text"
          name="display_name"
          value={form.display_name}
          onChange={handleChange}
          required
          className="w-full p-2 border rounded"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Username (уникальный)</label>
        <input
          type="text"
          name="username"
          value={form.username}
          onChange={handleChange}
          required
          className="w-full p-2 border rounded"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Email</label>
        <input
          type="email"
          name="email"
          value={form.email}
          onChange={handleChange}
          className="w-full p-2 border rounded"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Возраст</label>
        <input
          type="number"
          name="age"
          value={form.age}
          onChange={handleChange}
          min={13}
          className="w-full p-2 border rounded"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">О себе (bio)</label>
        <textarea
          name="bio"
          value={form.bio}
          onChange={handleChange}
          rows={4}
          className="w-full p-2 border rounded"
          placeholder="Расскажите о себе..."
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:opacity-50"
      >
        {isLoading ? 'Сохраняем...' : 'Сохранить профиль'}
      </button>
    </form>
  );
};
