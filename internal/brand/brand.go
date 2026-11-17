// Package brand keeps the product identity in one place. Renaming STEFANI must
// be a data change, not a refactor: every other package reads these values.
package brand

// Name is the product name shown in the UI and window title.
const Name = "STEFANI"

// Tagline is the one-line description used in metadata and the About panel.
const Tagline = "Escucha, lee y pronuncia: ingles a traves de la musica"

// Identifier is the bundle/app identifier used by the packaging task files.
const Identifier = "app.stefani.desktop"

// Company is the vendor shown in installers.
const Company = "Stefani Project"

// DataFolder is the per-user folder name. Changing it moves the user data, so
// the migration path is documented in docs/DECISIONS.md.
const DataFolder = "stefani"

// Version is stamped at build time by the packaging pipeline.
var Version = "0.1.0"
