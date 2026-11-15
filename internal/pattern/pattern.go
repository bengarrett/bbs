// Package pattern provides shared regular expression patterns and pre-compiled
// matchers for BBS color code formats.
package pattern

import "regexp"

// Regular expression strings to match BBS color codes.
//
// These expressions and case detections are intentional, based on
// real world examples, documentation and program sources.
//
//   - Celerity codes are sourced from the documentation found in v2 and v1.2.
//   - PCBoard codes are referenced from CUSTBBS.DOC found in disk 3 of the PCBoard v15 DEMO.
//   - Renegard pipes are also Telegard pipes are based on the Telegard v3 and its COLOURS.REF.
//   - Telegard grave is also covered in Telegard v3 COLOURS.REF.
//   - Wildcat codes were based on sample .BBS text files found in Wildcat! v4.
//   - WWIV codes are based on the source code of WWIV 4.21.
const (
	Celerity  = `\|(k|b|g|c|r|m|y|w|d|B|G|C|R|M|Y|W|S)`
	PCBoard   = `@X([0-9A-F][0-9A-F])`
	Renegade  = `\|(0[0-9]|1[0-9]|2[0-3])`
	Telegard  = "`([0-9A-F])([0-9A-F])"
	Wildcat   = `@([0-9A-F])([0-9A-F])@`
	WWIVHash  = `\|#(\d)`
	WWIVHeart = `\x03(\d)`
)

// Pre-compiled regular expressions for BBS color codes.
var (
	CelerityRe  = regexp.MustCompile(Celerity)
	PCBoardRe   = regexp.MustCompile(PCBoard)
	RenegadeRe  = regexp.MustCompile(Renegade)
	TelegardRe  = regexp.MustCompile(Telegard)
	WildcatRe   = regexp.MustCompile(Wildcat)
	WWIVHashRe  = regexp.MustCompile(WWIVHash)
	WWIVHeartRe = regexp.MustCompile(WWIVHeart)
)
