import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";
import { LyricLine } from "./LyricLine";
import type { Line } from "../../lib/types";

const wordLine: Line = {
  id: "l1",
  index: 0,
  text: "Take it slowly",
  startMs: 10000,
  endMs: 12000,
  words: [
    { text: "Take", startMs: 10000, endMs: 10600 },
    { text: "it", startMs: 10600, endMs: 11000 },
    { text: "slowly,", startMs: 11000, endMs: 12000 },
  ],
  annotation: {
    lineId: "l1",
    accent: "en-US",
    ipa: "/teɪk ɪt ˈsloʊli/",
    spanishApprox: "téik it slóuli",
    translation: "Tómatelo con calma",
    notes: [{ type: "stress", text: "El acento cae en la primera silaba." }],
    model: "test",
    source: "ai",
  },
};

function renderLine(overrides: Partial<ComponentProps<typeof LyricLine>> = {}) {
  const onSeek = vi.fn();
  const utils = render(
    <ul>
      <LyricLine
        line={wordLine}
        active
        past={false}
        timeMs={11000}
        showIPA
        showTranslation
        onSeek={onSeek}
        {...overrides}
      />
    </ul>,
  );
  return { onSeek, ...utils };
}

describe("LyricLine", () => {
  it("renders each word once, without duplicated karaoke text", () => {
    renderLine();
    expect(screen.getAllByText("Take")).toHaveLength(1);
    expect(screen.getByText("slowly,")).toBeInTheDocument();
  });

  it("shows IPA, approximation and translation for the active line", () => {
    renderLine();
    expect(screen.getByText("/teɪk ɪt ˈsloʊli/")).toBeInTheDocument();
    expect(screen.getByText("téik it slóuli")).toBeInTheDocument();
    expect(screen.getByText("Tómatelo con calma")).toBeInTheDocument();
    expect(screen.getByText("El acento cae en la primera silaba.")).toBeInTheDocument();
  });

  it("hides study data when the toggles are off", () => {
    renderLine({ showIPA: false, showTranslation: false });
    expect(screen.queryByText("/teɪk ɪt ˈsloʊli/")).not.toBeInTheDocument();
    expect(screen.queryByText("Tómatelo con calma")).not.toBeInTheDocument();
  });



  it("seeks to the line start when the line body is clicked", async () => {
    const { onSeek } = renderLine({ active: false, timeMs: 0 });
    await userEvent.click(screen.getByLabelText("Ir a: Take it slowly"));
    expect(onSeek).toHaveBeenCalledWith(10000);
  });

  it("does not show study data on inactive lines", () => {
    renderLine({ active: false, timeMs: 0 });
    expect(screen.queryByText("/teɪk ɪt ˈsloʊli/")).not.toBeInTheDocument();
  });
});
