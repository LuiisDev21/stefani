import { useCallback, useState } from "react";

/**
 * Boolean preference persisted in localStorage.
 *
 * These are view preferences (show IPA, show translation), not application
 * settings: they must not round-trip through Go or the database, but they should
 * survive a restart.
 */
export function useLocalFlag(key: string, initial: boolean): [boolean, () => void] {
  const storageKey = `stefani.pref.${key}`;
  const [value, setValue] = useState<boolean>(() => {
    try {
      const stored = window.localStorage.getItem(storageKey);
      return stored === null ? initial : stored === "1";
    } catch {
      return initial;
    }
  });

  const toggle = useCallback(() => {
    setValue((current) => {
      const next = !current;
      try {
        window.localStorage.setItem(storageKey, next ? "1" : "0");
      } catch {
        // Private mode or a locked profile: the preference stays in memory.
      }
      return next;
    });
  }, [storageKey]);

  return [value, toggle];
}
