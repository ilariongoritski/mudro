import { MovieCatalogPage } from '@/pages/movie-catalog-page/ui/MovieCatalogPage'
import { useTelegramWebApp } from '@/features/telegram-miniapp/hooks/useTelegramWebApp'

const MoviesPage = () => {
  const { isTelegram, themeParams } = useTelegramWebApp()
  
  return (
    <div 
      style={isTelegram ? { 
        backgroundColor: themeParams?.bg_color,
        color: themeParams?.text_color,
        minHeight: '100vh'
      } : undefined}
    >
      <MovieCatalogPage />
    </div>
  )
}

export default MoviesPage
