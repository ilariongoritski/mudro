"use client";

import { useState } from "react";
import { useSlot } from "@/lib/slot/store";
import { blackjackAction, dropPlinko, getBalance, spinRoulette, startBlackjack, type BlackjackState, type PlinkoRisk, type RouletteResult } from "@/lib/casino-api";

const bets = [1, 2, 5, 10, 25, 50];
type Game = "roulette" | "plinko" | "blackjack";

export function GameLobby() {
  const [game, setGame] = useState<Game>("roulette");
  const [bet, setBet] = useState(1);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("Choose a game. Every action is settled by the casino server.");
  const [roulette, setRoulette] = useState<RouletteResult | null>(null);
  const [risk, setRisk] = useState<PlinkoRisk>("medium");
  const [blackjack, setBlackjack] = useState<BlackjackState | null>(null);
  const balance = useSlot((s) => s.balance);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);
  const setServerBalance = useSlot((s) => s.setServerBalance);

  const refreshBalance = async () => setServerBalance(await getBalance());
  const execute = async (action: () => Promise<void>) => {
    if (!isLoggedIn || busy) return;
    setBusy(true);
    try {
      await action();
      await refreshBalance();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Casino request failed");
    } finally {
      setBusy(false);
    }
  };

  const spinRouletteGame = () => void execute(async () => {
    const result = await spinRoulette(bet, "red");
    setRoulette(result);
    setMessage(`Roulette: ${result.winning_number} ${result.winning_color}; payout ${result.payout_amount}.`);
  });
  const drop = () => void execute(async () => {
    const result = await dropPlinko(bet, risk);
    setMessage(`Plinko: x${result.multiplier.toFixed(2)}; payout ${result.payout}.`);
  });
  const start = () => void execute(async () => {
    const result = await startBlackjack(bet);
    setBlackjack(result);
    setMessage(result.status === "resolved" ? `Blackjack resolved; payout ${result.payout}.` : "Blackjack: Hit or Stand.");
  });
  const action = (value: "hit" | "stand") => void execute(async () => {
    const result = await blackjackAction(value);
    setBlackjack(result);
    setMessage(result.status === "resolved" ? `Blackjack ${result.winner ?? "resolved"}; payout ${result.payout}.` : "Your turn.");
  });

  if (!isLoggedIn) return null;
  return <section className="mx-auto w-full max-w-2xl px-3 pb-8">
    <h2 className="mb-3 text-center text-lg font-black tracking-wide">MORE GAMES</h2>
    <div className="mb-3 grid grid-cols-3 gap-2">
      {([ ["roulette", "🎡 Roulette"], ["plinko", "🔺 Plinko"], ["blackjack", "🂡 Blackjack"] ] as const).map(([id,label]) => <button key={id} onClick={() => setGame(id)} className={`rounded-xl border px-2 py-3 text-xs font-bold ${game === id ? "border-emerald-400 bg-emerald-500/20" : "border-white/10 bg-white/5"}`}>{label}</button>)}
    </div>
    <div className="rounded-2xl border border-white/10 bg-zinc-950/60 p-4 text-center">
      <div className="mb-3 flex flex-wrap justify-center gap-1">{bets.map(value => <button key={value} onClick={() => setBet(value)} className={`rounded-lg px-2 py-1 text-xs ${bet === value ? "bg-emerald-500 text-white" : "bg-white/10"}`}>{value}</button>)}</div>
      {game === "roulette" && <><p className="mb-3 text-sm text-slate-300">Instant roulette: red bet</p><button disabled={busy || balance < bet} onClick={spinRouletteGame} className="rounded-xl bg-amber-400 px-5 py-3 font-black text-zinc-950">{busy ? "SPINNING…" : "SPIN ROULETTE"}</button>{roulette && <p className="mt-3 font-bold">Result: {roulette.winning_number} ({roulette.winning_color})</p>}</>}
      {game === "plinko" && <><div className="mb-3 flex justify-center gap-2">{(["low","medium","high"] as PlinkoRisk[]).map(value => <button key={value} onClick={() => setRisk(value)} className={`rounded-lg px-2 py-1 text-xs ${risk === value ? "bg-sky-500" : "bg-white/10"}`}>{value}</button>)}</div><button disabled={busy || balance < bet} onClick={drop} className="rounded-xl bg-sky-500 px-5 py-3 font-black">{busy ? "DROPPING…" : "DROP BALL"}</button></>}
      {game === "blackjack" && <>{blackjack ? <div><p className="mb-3">Dealer {blackjack.dealer_hand.score} · You {blackjack.player_hand.score}</p>{blackjack.status === "player_turn" ? <div className="flex justify-center gap-2"><button disabled={busy} onClick={() => action("hit")} className="rounded-xl bg-emerald-500 px-4 py-2 font-bold">HIT</button><button disabled={busy} onClick={() => action("stand")} className="rounded-xl bg-white/15 px-4 py-2 font-bold">STAND</button></div> : <button disabled={busy || balance < bet} onClick={start} className="rounded-xl bg-violet-500 px-4 py-2 font-bold">NEW HAND</button>}</div> : <button disabled={busy || balance < bet} onClick={start} className="rounded-xl bg-violet-500 px-5 py-3 font-black">DEAL</button>}</>}
      <p role="status" className="mt-4 text-xs text-slate-400">{message}</p>
    </div>
  </section>;
}
