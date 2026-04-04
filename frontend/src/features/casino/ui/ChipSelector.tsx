import React from 'react';
import { rouletteQuickStakeValues } from '../lib/roulette';

interface ChipSelectorProps {
  chipAmount: number;
  onSelect: (amount: number) => void;
  disabled?: boolean;
}

export const ChipSelector: React.FC<ChipSelectorProps> = ({ chipAmount, onSelect, disabled }) => {
  return (
    <div>
      <div className="text-xs uppercase tracking-widest opacity-50 mb-2">Фишка</div>
      <div className="flex gap-2">
        {rouletteQuickStakeValues.map((v) => (
          <button
            key={v}
            onClick={() => onSelect(v)}
            disabled={disabled}
            className="flex-1 h-10 rounded-xl text-sm font-heading font-bold transition-all active:scale-90"
            style={{
              background: chipAmount === v
                ? 'linear-gradient(135deg, #f5c842, #c8a000)'
                : 'rgba(255,255,255,0.06)',
              color: chipAmount === v ? '#0d0d1a' : 'rgba(255,255,255,0.7)',
              border: chipAmount === v ? '1px solid #f5c842' : '1px solid rgba(255,255,255,0.08)',
              boxShadow: chipAmount === v ? '0 0 12px rgba(245,200,66,0.4)' : undefined,
              opacity: disabled ? 0.5 : 1,
            }}
          >
            {v >= 1000 ? `${v / 1000}K` : v}
          </button>
        ))}
      </div>
    </div>
  );
};
