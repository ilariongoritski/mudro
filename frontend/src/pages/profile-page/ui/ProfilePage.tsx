import { User } from 'lucide-react'
import { Card, CardContent } from '@/shared/ui/card'

const ProfilePage = () => {
  return (
    <div className="max-w-4xl mx-auto p-4 md:p-6">
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-16 text-center">
          <User className="text-mudro-muted mb-4" size={48} />
          <h2 className="text-xl font-semibold text-mudro-text">Профиль</h2>
          <p className="text-sm text-mudro-muted mt-1">Этот раздел скоро появится</p>
        </CardContent>
      </Card>
    </div>
  )
}

export default ProfilePage
