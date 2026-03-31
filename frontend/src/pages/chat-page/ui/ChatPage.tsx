import { type FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'

import { useChatRoom } from '@/features/chat/model/useChatRoom'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

import './ChatPage.css'

const roomName = 'main'

export const ChatPage = () => {
  const [draft, setDraft] = useState('')
  const currentUser = useAppSelector((state) => state.session.user)
  const {
    connectionLabel,
    error,
    isLoading,
    isSending,
    messages,
    refetch,
    sendMessage,
  } = useChatRoom({ room: roomName })

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!draft.trim()) {
      return
    }

    try {
      await sendMessage(draft)
      setDraft('')
    } catch (err) {
      console.error('Chat send failed', err)
    }
  }

  return (
    <main className="chat-page mudro-fade-up">
      <section className="chat-page__hero">
        <div className="chat-page__copy">
          <span className="chat-page__eyebrow">Mudro chat</span>
          <h1 className="chat-page__title">Realtime-комната поверх текущего feed-api runtime.</h1>
          <p className="chat-page__lead">
            Первый merge не трогает gateway и BFF: история идёт через Postgres, события приходят по WebSocket, а доступ
            остаётся только для авторизованных пользователей.
          </p>
          <div className="chat-page__actions">
            <Link to="/" className="chat-page__action chat-page__action_secondary">
              Вернуться в ленту
            </Link>
            <button type="button" className="chat-page__action" onClick={() => refetch()}>
              Обновить историю
            </button>
          </div>
        </div>

        <div className="chat-page__status-card">
          <span className="chat-page__status-label">Состояние</span>
          <strong>{connectionLabel}</strong>
          <p>
            Комната: <code>{roomName}</code>
          </p>
          <p>
            Пользователь:{' '}
            <strong>{currentUser?.username ?? 'session token loaded'}</strong>
          </p>
        </div>
      </section>

      <section className="chat-page__surface">
        <div className="chat-page__messages" aria-live="polite">
          {isLoading ? <p className="chat-page__empty">Загружаем историю сообщений…</p> : null}

          {!isLoading && messages.length === 0 ? (
            <p className="chat-page__empty">Комната пуста. Отправь первое сообщение и проверь broadcast в другой сессии.</p>
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

        <form className="chat-page__composer" onSubmit={handleSubmit}>
          <label className="chat-page__composer-label" htmlFor="chat-body">
            Сообщение
          </label>
          <textarea
            id="chat-body"
            className="chat-page__textarea"
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            placeholder="Напиши короткое сообщение в main room"
            rows={4}
          />
          {error ? <p className="chat-page__error">Не удалось синхронизировать чат. Проверь auth и feed-api.</p> : null}
          <div className="chat-page__composer-actions">
            <span className="chat-page__hint">Сообщения пишутся в Postgres и раздаются через WebSocket.</span>
            <button type="submit" className="chat-page__submit" disabled={isSending || !draft.trim()}>
              {isSending ? 'Отправляем…' : 'Отправить'}
            </button>
          </div>
        </form>
      </section>
    </main>
  )
}
