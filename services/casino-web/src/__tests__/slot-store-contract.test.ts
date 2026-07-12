import { describe, expect, test } from "bun:test";
import { useSlot } from "@/lib/slot/store";

describe("slot store contract", () => {
  test("provides the initial board action required by SlotMachine", () => {
    const state = useSlot.getState();

    expect(typeof state.seedBoard).toBe("function");
    expect(state.board).toHaveLength(5);
  });
});
