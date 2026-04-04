import React, { useState } from 'react';
import { type RouletteBetType } from '../api/casinoApi';
import { rouletteNumberGrid, getRouletteColor } from '../lib/roulette';

export interface PlacedBet {
  type: RouletteBetType;
  value?: string;
  amount: number;
}

interface RouletteBettingBoardProps {
  placedBets: PlacedBet[];
  onAddBet: (type: RouletteBetType, value?: string) => void;
  spinning: boolean;
}

export const RouletteBettingBoard: React.FC<RouletteBettingBoardProps> = ({
  placedBets,
  onAddBet,
  spinning,
}) => {
  const [showGrid, setShowGrid] = useState(false);

  const getBetForType = (type: RouletteBetType, value?: string) => {
    return placedBets.find((b) => b.type === type && b.value === value);
  };

  const numColor = (n: number) => {
    const color = getRouletteColor(n);
    if (n === 0) return '#16a34a';
    return color === 'red' ? '#dc2626' : '#1d4ed8';
  };

  const simpleBets = [
    { label: '🔴 Красное', type: 'red' as RouletteBetType, col: '#dc2626' },
    { label: '⚫ Чёрное', type: 'black' as RouletteBetType, col: '#374151' },
    { label: '🟢 Зеро', type: 'green' as RouletteBetType, col: '#16a34a' },
  ];

  const secondaryBets = [
    { label: 'Нечёт', type: 'odd' as RouletteBetType },
    { label: 'Чёт', type: 'even' as RouletteBetType },
    { label: '1–18', type: 'low' as RouletteBetType },
  ];

  const dozenBets = [
    { label: '1–12', type: 'dozen1' as RouletteBetType },
    { label: '13–24', type: 'dozen2' as RouletteBetType },
    { label: '25–36', type: 'dozen3' as RouletteBetType },
  ];

  return (
    <div
      className="rounded-2xl p-4 space-y-3"
      style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
    >
      <div className="text-xs uppercase tracking-widest opacity-50 mb-1">Ставки</div>

      {/* Simple bets row 1 */}
      <div className="grid grid-cols-3 gap-2">
        {simpleBets.map(({ label, type, col }) => {
          const b = getBetForType(type);
          return (
            <button
              key={type}
              onClick={() => onAddBet(type)}
              disabled={spinning}
              className="h-12 rounded-xl text-xs font-bold transition-all active:scale-90 flex flex-col items-center justify-center gap-0.5"
              style={{
                background: b ? `${col}33` : 'rgba(255,255,255,0.05)',
                border: b ? `2px solid ${col}` : '1px solid rgba(255,255,255,0.1)',
                color: b ? '#fff' : 'rgba(255,255,255,0.7)',
                boxShadow: b ? `0 0 12px ${col}55` : undefined,
              }}
            >
              {label}
              {b && <span className="text-[9px] opacity-70">× {b.amount}</span>}
            </button>
          );
        })}
      </div>

      {/* Simple bets row 2 */}
      <div className="grid grid-cols-3 gap-2">
        {secondaryBets.map(({ label, type }) => {
          const b = getBetForType(type);
          return (
            <button
              key={type}
              onClick={() => onAddBet(type)}
              disabled={spinning}
              className="h-10 rounded-xl text-xs font-bold transition-all active:scale-90"
              style={{
                background: b ? 'rgba(245,200,66,0.15)' : 'rgba(255,255,255,0.05)',
                border: b ? '2px solid rgba(245,200,66,0.6)' : '1px solid rgba(255,255,255,0.1)',
                color: b ? '#f5c842' : 'rgba(255,255,255,0.6)',
              }}
            >
              {label}
              {b && ` ×${b.amount}`}
            </button>
          );
        })}
      </div>

      {/* Dozen bets */}
      <div className="grid grid-cols-3 gap-2">
        {dozenBets.map(({ label, type }) => {
          const b = getBetForType(type);
          return (
            <button
              key={type}
              onClick={() => onAddBet(type)}
              disabled={spinning}
              className="h-10 rounded-xl text-xs font-bold transition-all active:scale-90"
              style={{
                background: b ? 'rgba(245,200,66,0.15)' : 'rgba(255,255,255,0.05)',
                border: b ? '2px solid rgba(245,200,66,0.6)' : '1px solid rgba(255,255,255,0.1)',
                color: b ? '#f5c842' : 'rgba(255,255,255,0.6)',
              }}
            >
              {label}
              {b && ` ×${b.amount}`}
            </button>
          );
        })}
      </div>

      {/* Number grid toggle */}
      <button
        onClick={() => setShowGrid((v) => !v)}
        className="w-full h-9 rounded-xl text-xs font-bold transition-all active:scale-90"
        style={{
          background: showGrid ? 'rgba(245,200,66,0.1)' : 'rgba(255,255,255,0.04)',
          border: '1px solid rgba(245,200,66,0.2)',
          color: '#f5c842',
        }}
      >
        {showGrid ? 'Скрыть числа' : 'Ставка на число'}
      </button>

      {showGrid && (
        <div className="space-y-1 animate-slide-up">
          {/* Zero */}
          <button
            onClick={() => onAddBet('straight', '0')}
            disabled={spinning}
            className="w-full h-8 rounded-lg text-xs font-bold transition-all active:scale-90"
            style={{
              background: getBetForType('straight', '0') ? '#15803d' : '#14532d',
              color: '#fff',
              border: '1px solid #16a34a',
            }}
          >
            0
          </button>

          {rouletteNumberGrid.map((row, ri) => (
            <div key={ri} className="grid grid-cols-12 gap-0.5">
              {row.map((n) => {
                const active = !!getBetForType('straight', String(n));
                const color = numColor(n);
                return (
                  <button
                    key={n}
                    onClick={() => onAddBet('straight', String(n))}
                    disabled={spinning}
                    className="h-8 rounded text-[10px] font-bold transition-all active:scale-90"
                    style={{
                      background: active ? color : `${color}55`,
                      color: '#fff',
                      border: active ? `1px solid ${color}` : '1px solid transparent',
                      boxShadow: active ? `0 0 8px ${color}88` : undefined,
                    }}
                  >
                    {n}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
