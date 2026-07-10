import React, { useState } from 'react';

interface AvatarUploadProps {
  currentAvatar?: string | null;
  onUpload: (file: File) => Promise<void>;
  isLoading?: boolean;
}

export const AvatarUpload: React.FC<AvatarUploadProps> = ({ currentAvatar, onUpload, isLoading }) => {
  const [preview, setPreview] = useState<string | null>(currentAvatar || null);

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Simple preview
    const reader = new FileReader();
    reader.onload = (ev) => setPreview(ev.target?.result as string);
    reader.readAsDataURL(file);

    await onUpload(file);
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="w-24 h-24 rounded-full overflow-hidden border-2 border-gray-300">
        {preview ? (
          <img src={preview} alt="Avatar" className="w-full h-full object-cover" />
        ) : (
          <div className="w-full h-full bg-gray-200 flex items-center justify-center text-gray-500">
            No avatar
          </div>
        )}
      </div>
      <label className="cursor-pointer bg-gray-100 hover:bg-gray-200 px-4 py-2 rounded text-sm">
        Загрузить аватар
        <input type="file" accept="image/*" onChange={handleFileChange} className="hidden" disabled={isLoading} />
      </label>
    </div>
  );
};
