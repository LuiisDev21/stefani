import { cn } from "../../lib/cn";

interface SwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label: string;
  description?: string;
  id?: string;
}

/**
 * Pill switch, the system's control for a binary setting. The track is a
 * hairline capsule; the thumb is white and slides, the fill turns Action Blue
 * only when the setting is on. Native checkboxes stay for lists and forms.
 */
export function Switch({ checked, onChange, label, description, id }: SwitchProps) {
  const switchId = id ?? label.replace(/\s+/g, "-").toLowerCase();
  return (
    <div className="flex items-start justify-between gap-6">
      <div className="min-w-0">
        <label htmlFor={switchId} className="block cursor-pointer text-[14px] text-ink">
          {label}
        </label>
        {description ? <p className="mt-1 max-w-[52ch] text-[13px] leading-[1.43] text-muted">{description}</p> : null}
      </div>
      <button
        id={switchId}
        type="button"
        role="switch"
        aria-checked={checked}
        aria-label={label}
        onClick={() => onChange(!checked)}
        className={cn(
          "relative mt-0.5 h-[26px] w-[44px] shrink-0 rounded-full transition-colors duration-200",
          checked ? "bg-accent" : "bg-[#d2d2d7]",
        )}
      >
        <span
          className={cn(
            "absolute top-[3px] h-5 w-5 rounded-full bg-white shadow-[0_1px_2px_rgba(0,0,0,0.18)] transition-[left] duration-200",
            checked ? "left-[21px]" : "left-[3px]",
          )}
        />
      </button>
    </div>
  );
}
