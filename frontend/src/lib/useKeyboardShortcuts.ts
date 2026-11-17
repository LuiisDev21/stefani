import { useEffect } from "react";
import { getAudioElement, usePlayer } from "../stores/playerStore";
import { useApp } from "../stores/appStore";

/**
 * Desktop keyboard shortcuts. They are ignored while a text field has focus so
 * typing a search query never pauses the music.
 */
export function useKeyboardShortcuts(): void {
  const toggle = usePlayer((state) => state.toggle);
  const seekMs = usePlayer((state) => state.seekMs);
  const toggleMute = usePlayer((state) => state.toggleMute);
  const setImmersive = useApp((state) => state.setImmersive);
  const notify = useApp((state) => state.notify);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const isField =
        target?.tagName === "INPUT" ||
        target?.tagName === "TEXTAREA" ||
        target?.tagName === "SELECT" ||
        target?.isContentEditable;
      if (isField) return;
      if (event.metaKey || event.ctrlKey || event.altKey) return;

      switch (event.key) {
        case " ": {
          event.preventDefault();
          toggle();
          break;
        }
        case "ArrowRight": {
          event.preventDefault();
          seekMs(Math.round((audioTime() + 5) * 1000));
          break;
        }
        case "ArrowLeft": {
          event.preventDefault();
          seekMs(Math.round(Math.max(0, audioTime() - 5) * 1000));
          break;
        }
        case "m": {
          toggleMute();
          break;
        }
        case "i": {
          setImmersive(true);
          break;
        }
        case "Escape": {
          setImmersive(false);
          break;
        }
        case "?": {
          notify("info", "Espacio: reproducir/pausar · ←/→: 5 s · M: silencio · I: modo inmersivo");
          break;
        }
        default:
          break;
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [toggle, seekMs, toggleMute, setImmersive, notify]);
}

function audioTime(): number {
  return getAudioElement()?.currentTime ?? 0;
}
