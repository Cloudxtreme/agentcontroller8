package core

import (
	"strings"
)

func IsTimeout(err error) bool {
	return strings.Contains(err.Error(), "timeout")
}
