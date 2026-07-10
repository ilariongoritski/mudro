import { useSlot } from "./store";
import { spin as apiSpin, getBalance, getHistory } from "../casino-api";

export async function performRealSpin(bet: number) {
  const store = useSlot.getState();
  if (!store.isLoggedIn) {
    throw new Error("Not logged in");
  }

  try {
    const result = await apiSpin(bet);

    store.setBalance(result.balanceAfter);
    store.setLastWin(result.win);
    store.setLastSymbols(result.symbols);
    store.setFairness?.({
      serverSeedHash: result.serverSeedHash,
      nonce: result.nonce,
    });

    // Trigger visual animation
    store.triggerSpinWithSymbols?.(result.symbols, result.win);

    return result;
  } catch (e) {
    console.error("Real spin failed:", e);
    // Fallback to local simulation
    store.spin();
    throw e;
  }
}

export async function hydrateRealBalance() {
  const store = useSlot.getState();
  if (!store.isLoggedIn) return;

  try {
    const balance = await getBalance();
    store.setBalance(balance);
  } catch (e) {
    console.warn("Could not hydrate balance from API");
  }
}

export async function loadRealHistory(limit = 20) {
  const store = useSlot.getState();
  if (!store.isLoggedIn) return;

  try {
    const history = await getHistory(limit);
    store.setHistory(history);
  } catch (e) {
    console.warn("History load failed");
  }
}
