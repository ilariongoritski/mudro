import { getBalance, spin, getHistory } from "@/lib/casino-api";
import { useSlot } from "@/lib/slot/store";

// Adapter must remain browser-safe: all requests use the public reverse proxy.
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
});
