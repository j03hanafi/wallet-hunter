package util

import "regexp"

var AddrRe = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var KeyRe = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)
