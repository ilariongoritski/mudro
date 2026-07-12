import { getBalance, spin, getHistory } from "@/lib/casino-api";
import { useSlot } from "@/lib/slot/store";

describe("Casino API adapter", () => {
  it("exports the supported server-backed actions", () => {
    expect(typeof getBalance).toBe("function");
    expect(typeof spin).toBe("function");
    expect(typeof getHistory).toBe("function");
  });

  it("updates the displayed wallet only from an explicit server balance", () => {
    useSlot.getState().setServerBalance(42);
    expect(useSlot.getState().balance).toBe(42);
  });

  it("treats an omitted backend win as a zero win", () => {
    useSlot.getState().applyServerSpin({ balanceAfter: 41, win: Number.NaN, symbols: [] });
    expect(useSlot.getState().displayWin).toBe(0);
    expect(useSlot.getState().winTier).toBe("none");
    expect(useSlot.getState().fairness).toBeNull();
  });
});
