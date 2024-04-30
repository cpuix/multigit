package flagparser

import (
	"strings"
)

type FlagParser struct {
	args  []string          // command arguments
	Flags map[string]string // degerlere direkt erismek icin map
}

func ParseFlags(args []string) FlagParser {
	parser := FlagParser{args: args} // FlagParser nesnesi oluştur

	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		flagName := arg[1:]
		flagValue := ""

		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			flagValue = args[i+1]
		}

		parser.flags[flagName] = flagValue
	}

	return parser
}

func (parser *FlagParser) GetFlag(flagName string) string {
	flagValue, ok := parser.flags[flagName]
	if !ok {
		return ""
	}

	return flagValue
}
