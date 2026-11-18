/**
 * Artwork helpers.
 *
 * YouTube returns thumbnails at whatever size the listing asked for (often
 * 120px). The same asset accepts a larger size in the URL, which is what the
 * immersive cover and the Windows media widget need.
 */
export function hiResArtwork(url: string, size = 544): string {
  if (!url) return url;
  return url.replace(/=w\d+-h\d+.*$/, `=w${size}-h${size}-l90-rj`);
}
