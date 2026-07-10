import { getBalance, spin, getHistory } from "@/lib/casino-api";

// Smoke tests for casino API adapter
describe("Casino API adapter", () => {
  it("should export getBalance function", () => {
    expect(typeof getBalance).toBe("function");
  });

  it("should export spin function", () => {
    expect(typeof spin).toBe("function");
  });

  it("should export getHistory function", () => {
    expect(typeof getHistory).toBe("function");
  });
});
