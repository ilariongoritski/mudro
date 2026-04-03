import { type FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'

import { useChatRoom } from '@/features/chat/model/useChatRoom'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { getErrorMessage } from '@/shared/lib/apiError'

import './ChatPage.css'

const roomName = 'main'

export const ChatPage = () => {
  const [draft, setDraft] = useState('')
  const [sendError, setSendError] = useState<string | null>(null)
  const currentUser = useAppSelector((state) => state.session.user)
  const isAuthenticated = useAppSelector((state) => state.session.isAuthenticated)

  const {
    connectionLabel,
    error,
    isLoading,
    isFetching,
    isSending,
    messages,
    hasMore,
    refetch,
    loadMore,
    sendMessage,
  } = useChatRoom({ room: roomName })

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!draft.trim()) {
      return
    }

    setSendError(null)
    try {
      await sendMessage(draft)
      setDraft('')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Не удалось отправить сообщение.'
      setSendError(msg)
    }
  }

  return (
    <main className="chat-page mudro-fade-up">
      <section className="chat-page__hero">
        <div className="chat-page__copy">
          <span className="chat-page__eyebrow">Мессенджер</span>
          <h1 className="chat-page__title">Общий чат в реальном времени.</h1>
          <p className="chat-page__lead">
            Сообщения сохраняются в базе данных и транслируются через WebSocket всем участникам комнаты.
          </p>
          <div className="chat-page__actions">
            <button type="button" className="chat-page__action" onClick={() => refetch()} disabled={isLoading}>
              {isLoading ? 'Загрузка...' : 'Обновить историю'}
            </button>
          </div>
        </div>

        <div className="chat-page__status-card">
          <span className="chat-page__status-label">Состояние</span>
          <strong>{isAuthenticated ? connectionLabel : 'Гость'}</strong>
          <p>
            Комната: <code>{roomName}</code>
          </p>
          <p>
            Участник:{' '}
            <strong>{currentUser?.username ?? 'Не авторизован'}</strong>
          </p>
        </div>
      </section>

      <section className="chat-page__surface">
        <div className="chat-page__messages" aria-live="polite">
          {isLoading && messages.length === 0 ? <p className="chat-page__empty">Загружаем историю сообщений…</p> : null}

          {isAuthenticated && hasMore && messages.length > 0 ? (
            <div className="chat-page__load-more">
              <button type="button" className="chat-page__load-more-btn" onClick={() => loadMore()} disabled={isFetching}>
                {isFetching ? 'Загружаем...' : 'Загрузить более старые'}
              </button>
            </div>
          ) : null}

          {!isLoading && messages.length === 0 && isAuthenticated ? (
            <p className="chat-page__empty">Комната пуста. Отправьте первое сообщение.</p>
          ) : null}

          {!isLoading && !isAuthenticated && messages.length === 0 ? (
            <p className="chat-page__empty">Войдите в аккаунт, чтобы видеть историю и отправлять сообщения.</p>
          ) : null}

          {messages.map((message) => {
            const ownMessage = currentUser?.id === message.user.id

            return (
              <article
                key={message.id}
                className={`chat-page__message${ownMessage ? ' chat-page__message_own' : ''}`}
              >
                <header className="chat-page__message-meta">
                  <strong>{message.user.username}</strong>
                  <span>{new Date(message.created_at).toLocaleString('ru-RU')}</span>
                </header>
                <p>{message.body}</p>
              </article>
            )
          })}
        </div>

        {isAuthenticated ? (
          <form className="chat-page__composer" onSubmit={handleSubmit}>
            <label className="chat-page__composer-label" htmlFor="chat-body">
              Сообщение
            </label>
            <textarea
              id="chat-body"
              className="chat-page__textarea"
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder="Напишите сообщение..."
              rows={4}
            />
            {error ? <p className="chat-page__error">{getErrorMessage(error, 'Не удалось синхронизировать чат.')}</p> : null}
            {sendError ? <p className="chat-page__error">{sendError}</p> : null}
            <div className="chat-page__composer-actions">
              <span className="chat-page__hint">Все участники видят сообщения в реальном времени.</span>
              <button type="submit" className="chat-page__submit" disabled={isSending || !draft.trim()}>
                {isSending ? 'Отправляем…' : 'Отправить'}
              </button>
            </div>
          </form>
        ) : (
          <div className="chat-page__guest-cta">
            <p className="chat-page__guest-text">Войдите в аккаунт, чтобы писать в чат</p>
            <div className="chat-page__guest-actions">
              <Link to="/login" className="chat-page__submit">Войти</Link>
              <Link to="/register" className="chat-page__action chat-page__action_secondary">Зарегистрироваться</Link>
            </div>
          </div>
        )}
      </section>
    </main>
  )
}
