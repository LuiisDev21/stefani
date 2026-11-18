---
version: alpha
name: Raycast
website: "https://www.raycast.com"
description: A dark-canvas developer-tools system that treats the marketing page like an extended product screenshot. Pure-near-black canvas at "#07080a", hairline 1px borders at "#242728", Inter typography with the ss03 stylistic set enabled site-wide, and a single white CTA pill at "#ffffff" that anchors every primary action. The signature visual moment is a red diagonal-stripe gradient banner across the very top of the hero, fading from "#ff5757" to "#a1131a". Color in the chrome is reserved for category accents inside extension illustrations — Hacker News yellow ("#ffc533"), Slack red ("#ff6161"), Linear green ("#59d499"), info blue ("#57c1ff") — never on buttons or text. Section rhythm is generous at 96px, the page never breaks tonal continuity, and depth comes entirely from a four-step surface ladder rather than drop shadows.

seo:
  title: "Raycast Design System — Near-Black and Inter ss03"
  metaDescription: "Raycast's near-black #07080a canvas and Inter with the ss03 alternates as a DESIGN.md file, with components for React, Next.js and AI tools like Claude."
  highlights:
    - "Single dark surface mode — a 4-step ladder from #07080a canvas through #0d0d0d, #101111, to #121212 carries every elevation tier"
    - "White CTA pill at #ffffff is the universal primary action — there is no second primary color in the entire system"
    - "Inter with font-feature-settings ss03 enabled site-wide — the alternate single-story g is the brand's signature typographic detail"
    - "Hairline 1px borders at #242728 carry every card edge — the system has zero drop shadows"
    - "Red diagonal-stripe gradient from #ff5757 to #a1131a appears once per page maximum — the only chromatic moment on chrome"
  tags:
    - "Productivity & SaaS"
    - "Developer Tools & IDEs"
  lastUpdated: "2026-05-12"
  author:
    name: "Dov Azencot"
    url: "https://x.com/dovazencot"
  opening: |
    Raycast's marketing site reads like an extended product screenshot. The chrome IS the in-product command palette scaled up: pure-near-black canvas at #07080a, hairline 1px borders at #242728, command-palette-style cards with corners between 6 and 16px, and a single white CTA pill anchoring every primary action. The system runs Inter with font-feature-settings "calt", "kern", "liga", "ss03" enabled site-wide — the ss03 stylistic set swaps in Inter's single-story alternate g, which is the brand's signature typographic detail. Without that flag the chrome loses its voice and reads as generic Inter.

    This page packages the whole spec into a single DESIGN.md file using the Google Labs format. Inside: 24 color tokens grouped into brand, surface, text, semantic, and gradient bands; 16 typography tokens from a 64px display-xl down to 12px caption-sm; 7 corner radii from 0 to 9999px; 8 spacing tokens including the 96px section rhythm; and 30 components covering buttons, command-palette cards, pricing tiers, store extension cards, keycap glyphs, and the hero stripe band. Every value is quoted and machine-readable.

    Feed the file to Claude, Cursor, GitHub Copilot, or any agent that reads structured tokens. The AI will reproduce Raycast's restraint — single white CTA, hairline borders, ss03-enabled Inter, four-step surface ladder — rather than a generic dark-mode shadcn theme. Reference tokens directly in Tailwind config or CSS variables. Raycast is worth studying because it's the rare system where the marketing chrome and the in-product chrome are genuinely the same surface, and the discipline shows.
  related:
    - href: "https://github.com/google-labs-code/design.md"
      title: "The DESIGN.md specification"
      description: "Google Labs' open spec for machine-readable design system files — the format this page is built on."
    - href: "/design"
      title: "Browse all design systems"
      description: "The full directory of DESIGN.md files on shadcn.io, with live mockups for each."
    - href: "/blocks"
      title: "React blocks for shadcn/ui"
      description: "Production-ready hero, pricing, CTA, and dashboard sections built with Tailwind and shadcn primitives."
  questions:
    - id: "primary-color"
      title: "What is Raycast's primary brand color?"
      answer: "Raycast has no chromatic brand color in the traditional sense — the universal primary CTA is pure white at #ffffff, paired with black text at #000000. Saturated reds, yellows, greens, and blues appear only inside extension illustrations and category tiles, never on chrome. The one chromatic moment on chrome is a red diagonal-stripe gradient that fades from #ff5757 to #a1131a across the top of the hero band, used once per page maximum."
    - id: "dark-mode"
      title: "Does Raycast's design system have a light mode?"
      answer: "No — the system is dark-only by design. The base canvas is pure-near-black at #07080a, with a four-step surface ladder through #0d0d0d, #101111, and #121212 carrying every elevation tier. There is no light variant in the captured surfaces. Depth comes from the surface ladder rather than drop shadows, which the system avoids entirely."
    - id: "typography"
      title: "What typography does Raycast use, and what makes it distinctive?"
      answer: "Raycast runs Inter with font-feature-settings 'calt', 'kern', 'liga', 'ss03' enabled site-wide. The ss03 stylistic set swaps in Inter's alternate single-story g, which is the brand's signature typographic detail — without it the chrome reads as plain Inter and loses its voice. The display tier additionally enables ss02 and ss08 and disables standard liga to render the hero wordmark with its distinctive geometric construction. Sizes run from 64px display-xl at weight 600 down to 12px caption-sm at weight 400."
    - id: "shape-language"
      title: "What is Raycast's corner radius vocabulary?"
      answer: "The radius scale clusters tightly between 4 and 16px. Keycaps use 4px (xs), command-palette rows and inline buttons use 6px (sm), standard buttons and inputs use 8px (md), feature and pricing tier cards use 10px (lg), and large hero command-palette mockup containers use 16px (xl). The 9999px full radius is reserved for pill-tabs and avatar circles. Chrome never goes flat at 0px on cards and never exceeds 16px except for fully-rounded pills."
    - id: "use-in-project"
      title: "How do I use this DESIGN.md in a React project?"
      answer: "Feed the file to Claude, Cursor, Copilot, or any AI tool that reads structured design tokens. The agent will reproduce Raycast's restraint rather than a generic dark theme — single white CTA, hairline borders, ss03-enabled Inter, four-step surface ladder. You can also reference the tokens directly: every color, type style, radius, and spacing value is a quoted value you can paste into Tailwind config, CSS variables, or your own component library. Critically, enable font-feature-settings 'calt', 'kern', 'liga', 'ss03' on the body element or the type will lose its signature voice."
    - id: "known-gaps"
      title: "What's missing from this DESIGN.md spec?"
      answer: "A handful of items documented in the Known Gaps section: mobile screenshots were not captured, so responsive behavior is synthesized from desktop evidence and the breakpoint stack; hover states are not documented by system policy, even though Raycast's in-product command-palette rows have rich hover behavior; the in-product Raycast launcher running on macOS is referenced in marketing screenshots but not documented as a separate UI system; form validation states beyond the focused-input border treatment are absent; and authenticated chrome — account dashboard, billing settings, team management — is not in the captured pages."

colors:
  primary: "#ffffff"
  primary-pressed: "#e8e8e8"
  on-primary: "#000000"
  ink: "#f4f4f6"
  body: "#cdcdcd"
  charcoal: "#d3d3d4"
  mute: "#9c9c9d"
  ash: "#6a6b6c"
  stone: "#434345"
  on-dark: "#ffffff"
  on-dark-mute: "rgba(255,255,255,0.72)"
  canvas: "#07080a"
  surface: "#0d0d0d"
  surface-elevated: "#101111"
  surface-card: "#121212"
  button-fg: "#18191a"
  hairline: "#242728"
  hairline-soft: "rgba(255,255,255,0.08)"
  hairline-strong: "rgba(255,255,255,0.16)"
  accent-blue: "#57c1ff"
  accent-blue-soft: "rgba(87,193,255,0.15)"
  accent-red: "#ff6161"
  accent-red-soft: "rgba(255,97,97,0.15)"
  accent-green: "#59d499"
  accent-green-soft: "rgba(89,212,153,0.15)"
  accent-yellow: "#ffc533"
  accent-yellow-soft: "rgba(255,197,51,0.15)"
  hero-stripe-start: "#ff5757"
  hero-stripe-end: "#a1131a"
  key-bg-start: "#121212"
  key-bg-end: "#0d0d0d"

typography:
  display-xl:
    fontFamily: Inter
    fontSize: 64px
    fontWeight: 600
    lineHeight: 1.1
    letterSpacing: 0
    fontFeature: '"calt", "kern", "liga", "ss03"'
  display-lg:
    fontFamily: Inter
    fontSize: 56px
    fontWeight: 500
    lineHeight: 1.17
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  heading-xl:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: 500
    lineHeight: 1.6
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  heading-lg:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: 500
    lineHeight: 1.15
    letterSpacing: 0
    fontFeature: '"calt", "kern", "liga", "ss03"'
  heading-md:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  heading-sm:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: 0
    fontFeature: '"calt", "kern", "liga", "ss03"'
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: 0
    fontFeature: '"calt", "kern", "liga", "ss03"'
  body-strong:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: 0
    fontFeature: '"calt", "kern", "liga", "ss03"'
  body-sm-strong:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 500
    lineHeight: 1.6
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  caption-md:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: 0.1px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  caption-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: 0.4px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  link-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.3px
    fontFeature: '"calt", "kern", "liga", "ss03"'
  button-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 500
    lineHeight: 1.6
    letterSpacing: 0.2px
    fontFeature: '"calt", "kern", "liga", "ss03"'

rounded:
  none: 0px
  xs: 4px
  sm: 6px
  md: 8px
  lg: 10px
  xl: 16px
  full: 9999px

spacing:
  xxs: 2px
  xs: 4px
  sm: 8px
  md: 12px
  lg: 16px
  xl: 24px
  xxl: 32px
  section: 96px

components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button-md}"
    rounded: "{rounded.md}"
    padding: 8px 16px
    height: 36px
  button-primary-pressed:
    backgroundColor: "{colors.primary-pressed}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button-md}"
    rounded: "{rounded.md}"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.on-dark}"
    typography: "{typography.button-md}"
    rounded: "{rounded.md}"
    padding: 8px 16px
    height: 36px
  button-tertiary:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.button-md}"
    rounded: "{rounded.md}"
    padding: 8px 16px
    height: 36px
  button-disabled:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.ash}"
    rounded: "{rounded.md}"
  install-button:
    backgroundColor: "transparent"
    textColor: "{colors.on-dark}"
    typography: "{typography.button-md}"
    rounded: "{rounded.md}"
    padding: 6px 14px
  text-input:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 8px 12px
    height: 36px
  text-input-focused:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    rounded: "{rounded.md}"
  store-search-bar:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 10px 16px
    height: 44px
  command-palette-row:
    backgroundColor: "transparent"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.sm}"
    padding: 6px 10px
  command-palette-row-active:
    backgroundColor: "{colors.surface-card}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.sm}"
  pill-tab:
    backgroundColor: "transparent"
    textColor: "{colors.body}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.full}"
    padding: 4px 10px
  pill-tab-active:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.full}"
  badge-pro:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark-mute}"
    typography: "{typography.caption-sm}"
    rounded: "{rounded.xs}"
    padding: 2px 6px
  badge-info-soft:
    backgroundColor: "{colors.accent-blue-soft}"
    textColor: "{colors.accent-blue}"
    typography: "{typography.caption-sm}"
    rounded: "{rounded.xs}"
    padding: 2px 8px
  keycap:
    backgroundColor: "{colors.surface-card}"
    textColor: "{colors.body}"
    typography: "{typography.caption-md}"
    rounded: "{rounded.xs}"
    padding: 1px 6px
    height: 20px
  command-palette-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 0px
  feature-card-dark:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 24px
  feature-card-elevated:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 24px
  store-extension-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 16px
  pricing-tier-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 24px
  pricing-tier-card-featured:
    backgroundColor: "{colors.surface-elevated}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-md}"
    rounded: "{rounded.lg}"
    padding: 24px
  hero-stripe-band:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.on-dark}"
    typography: "{typography.display-xl}"
    rounded: "{rounded.none}"
    padding: 96px 48px
  app-icon-tile:
    backgroundColor: "{colors.surface-card}"
    rounded: "{rounded.md}"
    size: 48px
  app-icon-tile-large:
    backgroundColor: "{colors.surface-card}"
    rounded: "{rounded.md}"
    size: 64px
  primary-nav:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.on-dark}"
    typography: "{typography.body-sm-strong}"
    rounded: "{rounded.none}"
    height: 56px
  footer-section:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.body}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.none}"
    padding: 64px 48px
  link-inline:
    textColor: "{colors.on-dark}"
    typography: "{typography.link-md}"
---

## Overview

Raycast's marketing site reads like an extended product screenshot. The chrome IS the in-product command palette at marketing scale: pure-near-black canvas (`{colors.canvas}` — "#07080a"), hairline 1px borders (`{colors.hairline}` — "#242728"), command-palette-style cards with rounded corners between 6 and 16px, Inter typography with the **ss03 stylistic set enabled site-wide** (a single character — the alternate `g` — that gives Raycast's typography its signature subtle distinction), a single white CTA pill that anchors every primary action, and small splashes of saturated accent reserved for category illustrations.

The system has effectively one surface mode — dark — with a faint four-step surface ladder (`{colors.canvas}` → `{colors.surface}` → `{colors.surface-elevated}` → `{colors.surface-card}`) carrying cards, in-card panels, and key-cap glyph backgrounds. The signature decorative moment is a **red diagonal-stripe gradient band** across the very top of the home page hero, used as a launch-banner motif behind the headline — the only time saturated red appears on chrome. Beyond that single moment, color in the chrome is reserved for category accents inside extension and feature illustrations: Hacker News yellow ("#ffc533"), Slack red ("#ff6161"), Linear green ("#59d499"), info blue ("#57c1ff").

The design philosophy is "the marketing page is the product." Section rhythm is generous (`{spacing.section}` 96px) but the page never breaks tonal continuity — the whole site sits in one continuous dark mode, full-bleed product UI screenshots show Raycast's actual command palette, store, and AI chat surfaces, and the typography ligature settings (`ss03`) are inherited from the in-product app's text rendering.

**Key Characteristics:**
- Single dark surface mode with a 4-step surface ladder: `{colors.canvas}` ("#07080a") → `{colors.surface}` ("#0d0d0d") → `{colors.surface-elevated}` ("#101111") → `{colors.surface-card}` ("#121212")
- White CTA pill (`{colors.primary}` — "#ffffff") is the universal primary action; everything else is monochrome dark
- Inter typography with `font-feature-settings: "calt", "kern", "liga", "ss03"` enabled site-wide — the ss03 alternate `g` is part of the brand voice
- Hairline 1px borders (`{colors.hairline}` — "#242728") carry every card edge; there are no drop shadows in the system
- Multi-radius card vocabulary: `{rounded.sm}` (6px) for keycaps, `{rounded.md}` (8px) for buttons and small cards, `{rounded.lg}` (10px) for feature cards, `{rounded.xl}` (16px) for hero command-palette mockup containers
- Saturated category accents appear only inside extension tile imagery — never on chrome
- Signature red diagonal-stripe gradient band at the very top of the hero — three angled stripes from `{colors.hero-stripe-start}` ("#ff5757") to `{colors.hero-stripe-end}` ("#a1131a"), used once per page maximum

## Known Gaps

- **Mobile screenshots not captured** — responsive behavior synthesizes Raycast's mobile pattern (hamburger drawer, single-column grid, hero downscale) from desktop evidence and the breakpoint stack.
- **Hover states not documented** by system policy. Raycast's in-product app has rich hover behavior on command-palette rows that this document doesn't capture.
- **In-product app chrome** (the actual Raycast launcher running on macOS) is referenced in marketing screenshots but not documented as a separate UI system here. The marketing site is documented; the in-product app surface is its own design system.
- **Dark mode is the only mode** — no light variant exists in the captured surfaces.
- **Form validation states** beyond the focused-input border treatment are not present in the captured surfaces.
- **Authenticated chrome** (account dashboard, billing settings, team management) not in the captured pages.
