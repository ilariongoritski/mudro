import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import type { SerializedError } from '@reduxjs/toolkit';
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { useRegisterMutation } from '@/entities/session/api/authApi';
import { useAppDispatch } from '@/shared/lib/hooks/storeHooks';
import { setCredentials } from '@/entities/session/model/sessionSlice';
import { getErrorMessage } from '@/shared/lib/apiError';
import { MudroLogoMark } from '@/shared/ui/MudroLogoMark';
import { ProfileForm } from '@/shared/ui/ProfileForm/ProfileForm';

import '@/pages/login-page/ui/Auth.css';

export const RegisterPage = () => {
  const [step, setStep] = useState<'basic' | 'profile'>('basic');
  const [basicData, setBasicData] = useState({ login: '', email: '', password: '' });
  const [profileData, setProfileData] = useState<any>({});

  const [registerMutation, { isLoading }] = useRegisterMutation();
  const dispatch = useAppDispatch();
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleBasicSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setStep('profile');
  };

  const handleProfileSubmit = async (data: any) => {
    setError(null);
    try {
      // Combine basic + profile data and register
      const result = await registerMutation({
        username: basicData.login,
        email: basicData.email || undefined,
        password: basicData.password,
        // Extend payload with profile fields if backend supports
        display_name: data.display_name,
        age: data.age,
        bio: data.bio,
      }).unwrap();

      dispatch(setCredentials(result));
      navigate('/', { replace: true });
    } catch (err) {
      setError(getErrorMessage(err as FetchBaseQueryError | SerializedError | undefined, 'Ошибка регистрации'));
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card max-w-md">
        <Link to="/" className="auth-logo">
          <span className="auth-logo-mark"><MudroLogoMark /></span>
          <span className="auth-logo-text"><strong>Mudro</strong><small>Социальная сеть</small></span>
        </Link>

        <h1>Регистрация</h1>
        <p className="auth-subtitle">Шаг {step === 'basic' ? '1' : '2'} из 2</p>

        {step === 'basic' && (
          <form onSubmit={handleBasicSubmit} className="auth-form">
            <input type="text" placeholder="Логин (username)" value={basicData.login} onChange={e => setBasicData({...basicData, login: e.target.value})} required className="auth-input" />
            <input type="email" placeholder="Email" value={basicData.email} onChange={e => setBasicData({...basicData, email: e.target.value})} required className="auth-input" />
            <input type="password" placeholder="Пароль (мин. 6)" value={basicData.password} onChange={e => setBasicData({...basicData, password: e.target.value})} required minLength={6} className="auth-input" />

            <button type="submit" className="auth-button">Продолжить → Заполнить профиль</button>
          </form>
        )}

        {step === 'profile' && (
          <div>
            <p className="text-sm text-center mb-4 text-gray-600">Заполните информацию о себе (обязательно имя и username)</p>
            <ProfileForm
              initialData={{ display_name: '', username: basicData.login }}
              onSubmit={handleProfileSubmit}
              isLoading={isLoading}
            />
            <button onClick={() => setStep('basic')} className="mt-4 text-sm text-gray-500 hover:underline">← Назад</button>
          </div>
        )}

        {error && <div className="auth-error mt-4">{error}</div>}

        <div className="auth-footer mt-6">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </div>
      </div>
    </div>
  );
};
