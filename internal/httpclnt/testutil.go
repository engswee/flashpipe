package httpclnt

import (
	"strconv"
	"strings"
)

func GetHostPort(url string) (string, int) {
	urlParts := strings.Split(strings.TrimPrefix(url, "http://"), ":")
	i, _ := strconv.Atoi(urlParts[1])
	return urlParts[0], i
}
