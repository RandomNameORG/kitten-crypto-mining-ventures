package ui

// Nerd Font glyph constants used across the UI and log strings.
//
// Requires a Nerd-Font-patched terminal font (any of the major patched
// fonts ship the Font Awesome range we use below). In a plain-font
// terminal these render as tofu boxes — that's the tradeoff. The old
// emoji set rendered inconsistently across terminals too (variation
// selectors, double-width surprises), so we picked our poison and ran
// with the Nerd Font family.
//
// Only Font Awesome range codepoints (U+F0xx–U+F2xx) are used here —
// they're the most broadly shipped subset across Nerd Font builds, and
// they're all single-width in a monospace grid.
const (
	// Logistics / hardware.
	IconPackage = "\uf1b2" // cube — shipping arrival / order placed
	IconBomb    = "\uf1e2" // bomb — GPU broken, catastrophic fail
	IconWrench  = "\uf0ad" // wrench — repair / print blueprint
	IconCog     = "\uf013" // cog — upgrade applied
	IconFire    = "\uf06d" // fire — fire / bricked upgrade
	IconPlug    = "\uf1e6" // plug — power outage / blackout
	IconSnow    = "\uf2dc" // snowflake — emergency vent / cooling
	IconBolt    = "\uf0e7" // bolt — electricity / voltage
	IconThermo  = "\uf2c7" // thermometer — heat gauge
	IconClock   = "\uf017" // clock — timer / countdown

	// Economy / progression.
	IconUSD      = "\uf155" // usd — money / bills
	IconBTC      = "\uf15a" // btc
	IconChartUp  = "\uf201" // line-chart — earning / pump
	IconBriefcase = "\uf0b1" // briefcase — wages
	IconTrophy    = "\uf091" // trophy — win / achievement

	// Research / knowledge.
	IconFlask     = "\uf0c3" // flask — research running
	IconCheck     = "\uf00c" // check — complete
	IconLightbulb = "\uf0eb" // lightbulb-o — tech point / idea

	// Social / threat.
	IconPaw       = "\uf1b0" // paw — kitten action / merc
	IconGift      = "\uf06b" // gift — free item / bonus
	IconShield    = "\uf132" // shield — defense
	IconSpy       = "\uf21b" // user-secret — thief
	IconTarget    = "\uf140" // bullseye — targeted theft
	IconMegaphone = "\uf0a1" // bullhorn — reputation hit / news
	IconWarning   = "\uf071" // exclamation-triangle

	// App-level UI.
	IconCat   = "\uf1b0" // paw (fallback for cat)
	IconPause = "\uf04c" // pause
	IconPlay  = "\uf04b" // play
	IconGlobe = "\uf0ac" // globe — language switch
)
