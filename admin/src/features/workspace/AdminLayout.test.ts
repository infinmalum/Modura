import { describe, expect, it } from "vitest";

import { visibleNavigation } from "./navigation";

describe("permission navigation", () => {
  it("always keeps overview and exposes only granted features", () => {
    const items = visibleNavigation(
      new Set(["settings.dictionaries/read", "audit.events/read"]),
    );

    expect(items.map((item) => item.key)).toEqual([
      "/",
      "/settings/dictionaries",
      "/audit",
    ]);
  });

  it("does not treat frontend navigation as a wildcard permission", () => {
    expect(visibleNavigation(new Set()).map((item) => item.key)).toEqual(["/"]);
  });
});
