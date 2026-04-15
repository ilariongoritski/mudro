import React, { useState } from 'react'
import { 
  useGetBlackjackStateQuery, 
  useStartBlackjackMutation, 
  useActionBlackjackMutation
} from '../../api/casinoApi'
import type { BlackjackCard, BlackjackHand } from '../../api/casinoApi'
import './Blackjack.css'

const SuitIcons: Record<string, string> = {
  hearts: '♥', diamonds: '♦', clubs: '♣', spades: '♠'
}

const winnerLabel: Record<string, string> = {
  player: 'Выигрыш',
  dealer: 'Проигрыш',
  push: 'Ничья',
}

const Card = ({ card, hidden }: { card: BlackjackCard; hidden?: boolean }) => {
  if (hidden) return <div className="card hidden" />
  return (
    <div className={`card ${card.suit}`}>
      <div className="card-rank">{card.rank}</div>
      <div className="card-suit">{SuitIcons[card.suit] || card.suit}</div>
    </div>
  )
}

const Hand = ({ hand, title, hideSecondDealerCard }: { hand: BlackjackHand, title: string, hideSecondDealerCard?: boolean }) => {
  return (
    <div className="hand-area">
      <div className="hand-header">
        <span className="hand-title">{title}</span>
        <span className="score-badge">
          {hideSecondDealerCard ? '??' : hand.score}
          {hand.is_bust && ' (BUST)'}
        </span>
      </div>
      <div className="hand-container">
        {hand.cards.map((card, idx) => (
          <Card 
            key={`${card.suit}-${card.rank}-${idx}`} 
            card={card} 
            hidden={hideSecondDealerCard && idx === 1} 
          />
        ))}
      </div>
    </div>
  )
}

export const Blackjack: React.FC = () => {
  const { data: state, isLoading } = useGetBlackjackStateQuery()
  const [start] = useStartBlackjackMutation()
  const [act] = useActionBlackjackMutation()
  const [bet, setBet] = useState(100)

  const game = state && 'id' in state ? state : null
  const isResolved = game?.status === 'resolved'

  if (isLoading) return <div>Загружаем стол...</div>

  return (
    <div className="blackjack-container">
      {game ? (
        <div className="blackjack-table">
          <Hand hand={game.dealer_hand} title="Дилер" hideSecondDealerCard={!isResolved} />
          <Hand hand={game.player_hand} title="Вы" />
          
          <div className="blackjack-controls">
            {!isResolved ? (
              <>
                <button className="blackjack-btn btn-primary" onClick={() => act({ action: 'hit' })}>Ещё карту</button>
                <button className="blackjack-btn btn-secondary" onClick={() => act({ action: 'stand' })}>Стоп</button>
              </>
            ) : (
              <div className="result-area">
                <div className="result-msg">{winnerLabel[game.winner ?? ''] ?? 'Раунд завершён'}</div>
                <button className="blackjack-btn btn-primary" onClick={() => start({ bet })}>Новая игра</button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="blackjack-start">
          <h3>Blackjack</h3>
          <div className="bet-input">
            <label>Ставка:</label>
            <input type="number" value={bet} onChange={e => setBet(Math.max(1, parseInt(e.target.value)))} />
          </div>
          <button className="blackjack-btn btn-primary" onClick={() => start({ bet })}>Раздать</button>
        </div>
      )}
    </div>
  )
}
