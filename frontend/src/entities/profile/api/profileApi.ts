import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { UserProfile, ProfileUpdate, CasinoStats, Activity } from '../model/profile.types';

export const profileApi = createApi({
  reducerPath: 'profileApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Profile'],
  endpoints: (builder) => ({
    getProfile: builder.query<UserProfile, number>({
      query: (id) => `/profile/${id}`,
      providesTags: (_result, _error, id) => [{ type: 'Profile', id }],
    }),
    updateProfile: builder.mutation<UserProfile, { userId: number; data: ProfileUpdate }>({
      query: ({ userId, data }) => ({
        url: '/profile/me',
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: (_result, _error, { userId }) => [{ type: 'Profile', id: userId }],
    }),
    getCasinoStats: builder.query<CasinoStats, number>({
      query: (userId) => `/profile/${userId}/casino`,
    }),
    getActivities: builder.query<Activity[], number>({
      query: (userId) => `/profile/${userId}/activities`,
    }),
    uploadAvatar: builder.mutation<{ avatar_url: string }, FormData>({
      query: (formData) => ({
        url: '/profile/avatar',
        method: 'POST',
        body: formData,
      }),
    }),
    startMessage: builder.mutation<{ chat_id: string }, number>({
      query: (targetId) => ({
        url: `/profile/${targetId}/message`,
        method: 'POST',
      }),
    }),
  }),
});

export const {
  useGetProfileQuery,
  useUpdateProfileMutation,
  useGetCasinoStatsQuery,
  useGetActivitiesQuery,
  useUploadAvatarMutation,
  useStartMessageMutation,
} = profileApi;
