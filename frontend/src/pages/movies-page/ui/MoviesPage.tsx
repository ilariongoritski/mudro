import { Film } from 'lucide-react'
import { Card, CardContent } from '@/shared/ui/card'

const MoviesPage = () => {
  return (
    <div className="max-w-4xl mx-auto p-4 md:p-6">
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-16 text-center">
          <Film className="text-slate-300 mb-4" size={48} />
          <h2 className="text-xl font-semibold text-slate-700">Фильмы</h2>
          <p className="text-sm text-slate-500 mt-1">Этот раздел скоро появится</p>
        </CardContent>
      </Card>
    </div>
  )
}

export default MoviesPage
