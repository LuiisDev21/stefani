import { describe, expect, it } from "vitest";
import { hiResArtwork } from "./artwork";

describe("hiResArtwork", () => {
  it("upgrades the size suffix YouTube adds to thumbnails", () => {
    const small =
      "https://yt3.googleusercontent.com/abc=w120-h120-l90-rj";
    expect(hiResArtwork(small)).toBe(
      "https://yt3.googleusercontent.com/abc=w544-h544-l90-rj",
    );
  });

  it("keeps URLs without a size suffix untouched", () => {
    const local = "/media/artwork/42";
    expect(hiResArtwork(local)).toBe(local);
  });

  it("handles an empty value", () => {
    expect(hiResArtwork("")).toBe("");
  });
});
