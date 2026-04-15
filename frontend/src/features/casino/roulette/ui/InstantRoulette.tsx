import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useGetCasinoBalanceQuery } from '../../api/casinoApi';
import { RouletteWheelSVG } from '../../ui/RouletteWheelSVG';
import { RouletteBettingBoard } from '../../ui/RouletteBettingBoard';
import { ChipSelector } from '../../ui/ChipSelector';
import { useInstantRoulette } from '../model/useInstantRoulette';
import { getRouletteColor, formatRouletteNumber } from '../../lib/roulette';

interface RouletteMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface InstantRouletteProps {
  isAuthenticated: boolean;
  userName?: string | null;
  onMainActionChange?: (action: RouletteMainAction | null) => void;
}

export const InstantRoulette: React.FC<InstantRouletteProps> = ({
  onMainActionChange,
}) => {
  const { data: balanceData, refetch: refetchBalance } = useGetCasinoBalanceQuery();

  const [localBalance, setLocalBalance] = React.useState<number>(0);

  React.useEffect(() => {
    if (balanceData?.balance !== undefined) {
      setLocalBalance(balanceData.balance);
    }
  }, [balanceData]);

  const {
    rotation,
    spinning,
    result,
    placedBets,
    chipAmount,
    setChipAmount,
    lastWin,
    totalBet,
    canSpin,
    addBet,
    clearBets,
    doSpin,
  } = useInstantRoulette(localBalance, setLocalBalance);

  const handleSpin = async () => {
    const res = await doSpin();
    if (res) {
      setTimeout(() => refetchBalance(), 5000); 
    }
  };

  React.useEffect(() => {
    if (onMainActionChange) {
      onMainActionChange({
        label: spinning ? 'КРУТИМ...' : canSpin ? `СПИН (${totalBet})` : 'Соберите ставку',
        busy: spinning,
        disabled: !canSpin,
        onTrigger: handleSpin,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [spinning, canSpin, totalBet, onMainActionChange]);

  const resColor = result !== null ? getRouletteColor(result) : null;

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-black/20 rounded-3xl p-4 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-heading font-black text-[#f5c842]">
            🎡 Мгновенная Рулетка
          </h2>
          <p className="text-[10px] opacity-40 uppercase tracking-wider">Европейская • 0–36</p>
        </div>
        <div className="px-3 py-1.5 rounded-full text-sm font-heading font-bold bg-[#f5c842]/10 text-[#f5c842] border border-[#f5c842]/20">
          🪙 {localBalance.toLocaleString()}
        </div>
      </div>

      {/* Wheel Area */}
      <div 
        className="rounded-3xl p-6 flex flex-col items-center gap-4 transition-all duration-500"
        style={{
          background: 'linear-gradient(145deg, #1a1a2e 0%, #0d0d1a 100%)',
          border: result !== null
            ? `2px solid ${resColor === 'green' ? '#16a34a' : resColor === 'red' ? '#ef4444' : '#1d4ed8'}`
            : '2px solid rgba(245,200,66,0.15)'
        }}
      >
          <RouletteWheelSVG rotation={rotation} />

        {/* Result & Win Display */}
        <div className="h-20 flex flex-col items-center justify-center">
          <AnimatePresence mode="wait">
            {spinning ? (
              <motion.div 
                key="spinning"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="text-sm opacity-50 tracking-[0.2em] uppercase font-black"
              >
                КРУТИМ...
              </motion.div>
            ) : result !== null ? (
              <motion.div 
                key="result"
                initial={{ opacity: 0, scale: 0.5, y: 10 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                className="text-center"
              >
                <div 
                  className="text-5xl font-heading font-black"
                  style={{
                    color: resColor === 'green' ? '#22c55e' : resColor === 'red' ? '#ef4444' : '#60a5fa',
                    textShadow: `0 0 30px ${resColor === 'green' ? 'rgba(34,197,94,0.6)' : resColor === 'red' ? 'rgba(239,68,68,0.6)' : 'rgba(96,165,250,0.8)'}`
                  }}
                >
                  {formatRouletteNumber(result)}
                </div>
                {lastWin !== null && (
                  <motion.div
                    initial={{ opacity: 0, y: 5 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.2 }}
                  >
                    {lastWin > 0 ? (
                      <div className="text-xl font-heading font-black text-[#00ff88] mt-1" style={{ textShadow: '0 0 16px rgba(0,255,136,0.6)' }}>
                        +{lastWin.toLocaleString()} 🪙
                      </div>
                    ) : (
                      <div className="text-xs opacity-50 text-[#ff4466] mt-1">Не сыграло</div>
                    )}
                  </motion.div>
                )}
              </motion.div>
            ) : null}
          </AnimatePresence>
        </div>
      </div>

      {/* Controls Area */}
      <div className="space-y-4">
        <ChipSelector 
          chipAmount={chipAmount} 
          onSelect={setChipAmount} 
          disabled={spinning} 
        />

        <RouletteBettingBoard 
          placedBets={placedBets} 
          onAddBet={addBet} 
          spinning={spinning} 
        />

        {/* Action Bar */}
        <div className="flex gap-2">
          {placedBets.length > 0 && (
            <button
              onClick={clearBets}
              disabled={spinning}
              className="flex-1 h-14 rounded-2xl text-xs font-black transition-all active:scale-95 bg-white/5 border border-white/10 hover:bg-white/10 uppercase tracking-widest"
            >
              Сброс ({totalBet})
            </button>
          )}
          <button
            onClick={handleSpin}
            disabled={!canSpin}
            className={`flex-[2] h-14 rounded-2xl font-heading font-black text-lg tracking-widest transition-all active:scale-95 flex items-center justify-center gap-2 ${
              canSpin ? 'bg-gradient-to-r from-[#f5c842] to-[#ffcf58] text-[#0d0d1a] shadow-[0_4px_20px_rgba(245,200,66,0.3)]' : 'bg-white/5 text-white/20 border border-white/10'
            }`}
          >
            {spinning ? '🎡 КРУТИМ...' : '🎡 СПИН'}
          </button>
        </div>
      </div>
    </div>
  );
};
