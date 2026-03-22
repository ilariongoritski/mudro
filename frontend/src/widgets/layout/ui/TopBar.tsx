import { Search } from 'lucide-react'
import { useLocation } from 'react-router'
import { Input } from '@/shared/ui/input'

const pageTitles: Record<string, string> = {
  '/': 'Лента',
  '/movies': 'Фильмы',
  '/chat': 'Чат',
  '/profile': 'Профиль',
}

export const TopBar = () => {
  const { pathname } = useLocation()
  const title = pageTitles[pathname] ?? 'Mudro'

  return (
    <header className="sticky top-0 z-10 flex items-center justify-between gap-4 h-14 px-4 md:px-6 bg-white/80 backdrop-blur border-b border-slate-200">
      <h1 className="text-lg font-semibold text-mudro-text truncate">{title}</h1>

      <div className="relative hidden sm:block w-56">
        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" size={15} />
        <Input className="pl-8 h-8 text-xs" placeholder="Поиск..." disabled />
      </div>
    </header>
  )
}
