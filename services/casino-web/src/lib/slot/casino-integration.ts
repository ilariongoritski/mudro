// services/casino-web/src/lib/slot/casino-integration.ts
// Thin bridge between existing local store and real casino-api
// Keeps UI intact, replaces only the core actions

import { useSlot } from './store';
import { spin as apiSpin, getBalance, getHistory, SpinResult } from '../casino-api';

export async function performRealSpin(bet: number, token: string) {
  const store = useSlot.getState();
  
  try {
    const result: SpinResult = await apiSpin(bet, token);
    
    // Update local store with real result
    store.setBalance(result.balanceAfter);
    store.setLastWin(result.win);
    store.setLastSymbols(result.symbols);
    store.setFairness({
      serverSeedHash: result.serverSeedHash,
      nonce: result.nonce,
    });
    
    // Trigger the visual spin animation with real symbols
    store.triggerSpinWithSymbols(result.symbols, result.win);
    
    return result;
  } catch (e) {
    console.error('Real spin failed:', e);
    // Fallback to local simulation if backend unavailable
    store.spin();
    throw e;
  }
}

export async function hydrateRealBalance(token: string) {
  try {
    const balance = await getBalance(token);
    useSlot.getState().setBalance(balance);
  } catch (e) {
    console.warn('Could not hydrate balance from API, using local');
  }
}

export async function loadRealHistory(token: string, limit = 20) {
  try {
    const history = await getHistory(limit, token);
    useSlot.getState().setHistory(history);
  } catch (e) {
    console.warn('History load failed');
  }
}
