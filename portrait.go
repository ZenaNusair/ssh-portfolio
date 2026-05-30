package main

import _ "embed"

//go:embed portrait_ascii.txt
var PortraitASCII string

//go:embed portrait_braille.txt
var PortraitBraille string

// Portrait is what the View renders by default — kept for back-compat with
// any old references. Always the safe ASCII version on first load.
var Portrait = PortraitASCII
