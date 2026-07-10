import { loginWithTelegram, logout } from "@/lib/auth";

// Basic smoke test for auth module
describe("Auth module", () => {
  it("should export loginWithTelegram function", () => {
    expect(typeof loginWithTelegram).toBe("function");
  });

  it("should export logout function", () => {
    expect(typeof logout).toBe("function");
  });
});
