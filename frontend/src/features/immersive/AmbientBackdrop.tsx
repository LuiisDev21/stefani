import { useEffect, useRef, useState } from "react";

/**
 * Room light for the immersive view.
 *
 * A 12×12 canvas is enough: scaled to the window the browser interpolates it
 * into a soft gradient, and the GPU only samples a tiny texture. That is far
 * cheaper than a 110 px blur over a full-size cover, which has to be
 * re-rasterised whenever the lyric column repaints.
 */
export function AmbientBackdrop({ url, active }: { url?: string; active: boolean }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [tint, setTint] = useState("rgb(28, 30, 36)");
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!url) {
      setReady(false);
      return;
    }
    let cancelled = false;
    const image = new Image();
    image.crossOrigin = "anonymous";
    image.referrerPolicy = "no-referrer";
    image.onload = () => {
      if (cancelled) return;
      const canvas = canvasRef.current;
      const context = canvas?.getContext("2d", { willReadFrequently: false });
      if (!canvas || !context) return;
      canvas.width = 12;
      canvas.height = 12;
      try {
        context.drawImage(image, 0, 0, 12, 12);
        const { data } = context.getImageData(0, 0, 12, 12);
        let r = 0;
        let g = 0;
        let b = 0;
        for (let i = 0; i < data.length; i += 4) {
          r += data[i];
          g += data[i + 1];
          b += data[i + 2];
        }
        const pixels = data.length / 4;
        setTint(`rgb(${Math.round(r / pixels)}, ${Math.round(g / pixels)}, ${Math.round(b / pixels)})`);
        setReady(true);
      } catch {
        // Tainted canvas (no CORS): keep the neutral tint and the flat panel.
        setReady(false);
      }
    };
    image.onerror = () => setReady(false);
    image.src = url;
    return () => {
      cancelled = true;
    };
  }, [url]);

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 overflow-hidden"
      style={{ backgroundColor: tint }}
    >
      {/* No CSS blur: a 12×12 bitmap scaled to the window is already a soft
          gradient, and blurring the scaled surface re-rasterises megabytes on
          every lyric paint. */}
      <canvas
        ref={canvasRef}
        className="absolute inset-0 h-full w-full scale-110"
        style={{ transform: "translateZ(0)", imageRendering: "auto" }}
      />
      <div
        className="absolute inset-0"
        style={{
          background:
            "radial-gradient(120% 85% at 50% 0%, rgba(255,255,255,0.10), transparent 55%), linear-gradient(to bottom, rgba(11,11,12,0.35), rgba(11,11,12,0.72) 55%, rgba(11,11,12,0.92))",
        }}
      />
      {ready ? null : (
        <div className="absolute inset-0 bg-[#0b0b0c]" />
      )}
    </div>
  );
}
