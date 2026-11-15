// Package pattern provides shared regular expression patterns and pre-compiled
// matchers for BBS color code formats.
package pattern

import "regexp"

// Regular expression strings to match BBS color codes.
//
// These expressions are all intentional, based on
// real world examples, documentation and program source code.
//
//   - Celerity codes are sourced from the documentation found in v2 and v1.2.
//   - PCBoard uses the MS-DOS era all uppercase codes, to help avoid false postives.
//   - Renegard pipes are also Telegard pipes are based on the Telegard v3 and its COLOURS.REF.
//   - Telegard grave is also covered in Telegard v3 COLOURS.REF.
//   - Wildcat codes were based on sample .BBS text files found in Wildcat! v4.
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
