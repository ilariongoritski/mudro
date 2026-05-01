import { useState } from 'react'
import { Send } from 'lucide-react'
import { useCreateCommentMutation } from '@/entities/post/model/postsApi'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { Input } from '@/shared/ui/input'

interface CommentFormProps {
  postId: number
}

export const CommentForm = ({ postId }: CommentFormProps) => {
  const [text, setText] = useState('')
  const [createComment, { isLoading }] = useCreateCommentMutation()
  const token = useAppSelector((state) => state.session.token)

  if (!token) {
    return (
      <p className="text-xs text-mudro-muted text-center py-2">
        Войдите, чтобы комментировать
      </p>
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = text.trim()
    if (!trimmed) return

    try {
      await createComment({ postId, text: trimmed }).unwrap()
      setText('')
    } catch {
      // error is handled by RTK Query
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-2 pt-3 border-t border-mudro-line">
      <label htmlFor={`comment-input-${postId}`} className="sr-only">Написать комментарий</label>
      <Input
        id={`comment-input-${postId}`}
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="Написать комментарий..."
        className="flex-1 text-sm"
        disabled={isLoading}
      />
      <button
        type="submit"
        disabled={isLoading || !text.trim()}
        aria-label="Отправить комментарий"
        className="p-2 text-mudro-accent hover:bg-mudro-accent/10 rounded-lg transition-colors disabled:opacity-40"
      >
        <Send className="w-4 h-4" aria-hidden="true" />
      </button>
    </form>
  )
}
