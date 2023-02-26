package utility

import (
	"os"
	"strings"
)

func isModeAsync(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == "--async"
}

func splitStr(sl []string) []string {
	var res []string
	for _, val := range sl {
		res = append(res, strings.Fields(val)...)
	}

	return res
}

func ParseArguments() ([]string, bool) {
	var UrlSl []string
	asyncMode := isModeAsync(os.Args[1:])
	if asyncMode {
		UrlSl = os.Args[2:]
	} else {
		UrlSl = os.Args[1:]
	}

	UrlSl = splitStr(UrlSl)

	return UrlSl, asyncMode
}
