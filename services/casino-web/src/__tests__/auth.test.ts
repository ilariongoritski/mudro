import { loginWithTelegram, logout, telegramAuthEndpoint } from "@/lib/auth";

// Basic smoke test for auth module
describe("Auth module", () => {
  it("should export loginWithTelegram function", () => {
    expect(typeof loginWithTelegram).toBe("function");
  });

  it("should export logout function", () => {
    expect(typeof logout).toBe("function");
  });

  it("uses the public versioned Telegram auth endpoint", () => {
    expect(telegramAuthEndpoint).toBe("/api/v1/auth/telegram");
  });
});
