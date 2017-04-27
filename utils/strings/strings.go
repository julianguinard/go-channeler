package strings

import "strings"

func Ucfirst(str string) string {
    return strings.ToUpper(string(str[0]))+str[1:]
}