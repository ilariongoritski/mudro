import { useSlot } from "./store";
import { spin as apiSpin, getBalance, getHistory } from "../casino-api";

export async function performRealSpin(bet: number) {
  const store = useSlot.getState();
  if (!store.token) {
    throw new Error("Not logged in");
  }

  try {
    const result = await apiSpin(bet);

    // Store doesn't have these setters - using local spin for animation
    store.spin(); // triggers local animation

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
  if (!store.token) return;

  try {
    const balance = await getBalance();
    // Balance is managed locally - server is authoritative
    console.log("Server balance:", balance);
  } catch (e) {
    console.warn("Could not hydrate balance from API");
  }
}

export async function loadRealHistory(limit = 20) {
  const store = useSlot.getState();
  if (!store.token) return;

  await store.loadHistory();
}
