import { useCallback, useRef, useState } from 'react';
import { type RouletteBetType, useInstantRouletteSpinMutation } from '../../api/casinoApi';
import { rouletteWheelOrder } from '../../lib/roulette';
import { type PlacedBet } from '../../ui/RouletteBettingBoard';

export function useInstantRoulette(balance: number, onUpdateBalance: (newBalance: number) => void) {
  const [rotation, setRotation] = useState(0);
  const [spinning, setSpinning] = useState(false);
  const [result, setResult] = useState<number | null>(null);
  const [placedBets, setPlacedBets] = useState<PlacedBet[]>([]);
  const [chipAmount, setChipAmount] = useState(25);
  const [lastWin, setLastWin] = useState<number | null>(null);
  const baseRotationRef = useRef(0);

  const [instantSpin] = useInstantRouletteSpinMutation();

  const totalBet = placedBets.reduce((s, b) => s + b.amount, 0);
  const canSpin = !spinning && placedBets.length > 0 && totalBet <= balance;

  const addBet = useCallback((type: RouletteBetType, value?: string) => {
    if (spinning) return;
    setPlacedBets((prev) => {
      const existing = prev.find((b) => b.type === type && (value === undefined || b.value === value));
      if (existing) {
        return prev.map((b) =>
          b.type === type && (value === undefined || b.value === value)
            ? { ...b, amount: b.amount + chipAmount }
            : b
        );
      }
      return [...prev, { type, value, amount: chipAmount }];
    });
  }, [spinning, chipAmount]);

  const clearBets = useCallback(() => {
    if (spinning) return;
    setPlacedBets([]);
    setLastWin(null);
  }, [spinning]);

  const doSpin = useCallback(async () => {
    if (!canSpin) return;
    
    setSpinning(true);
    setResult(null);
    setLastWin(null);

    try {
      const apiBets = placedBets.map(b => ({
        bet_type: b.type,
        bet_value: b.value ? parseInt(b.value, 10) : undefined,
        stake: b.amount
      }));

      const res = await instantSpin({ bets: apiBets }).unwrap();
      
      const winNumber = res.winning_number;
      const totalPayout = res.payout_amount;

      // Find index in wheel order to calculate rotation
      const resultIdx = rouletteWheelOrder.indexOf(winNumber);
      if (resultIdx === -1) {
        throw new Error('Invalid winning number from server');
      }

      // Spin animation logic
      const sectorDeg = resultIdx * (360 / rouletteWheelOrder.length);
      const spinAmount = 360 * 8 + (360 - sectorDeg);
      baseRotationRef.current += spinAmount;
      setRotation(baseRotationRef.current);

      // Wait for animation (matched with CSS transition)
      await new Promise((r) => setTimeout(r, 4200));

      setResult(winNumber);
      setSpinning(false);
      setLastWin(totalPayout);
      
      // Update global balance from server response
      onUpdateBalance(res.balance);
      
      return { winNumber, totalWin: totalPayout };
    } catch (err) {
      console.error('Spin failed:', err);
      setSpinning(false);
      // Maybe show an error Toast here
    }
  }, [canSpin, placedBets, instantSpin, onUpdateBalance]);

  return {
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
  };
}
