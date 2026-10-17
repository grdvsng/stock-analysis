package Regex

import (
	"regexp"
)

type re struct {
	*regexp.Regexp
}

func Compile(patern string) *re {
	result := re{Regexp: regexp.MustCompile(patern)}

	return &result
}

func (r *re) NamedGroups(str string) groups {
	match := r.FindStringSubmatch(str)
	names := r.SubexpNames()
	data := make(map[string]string, len(names))

	if len(match) >= len(names) {
		for i, name := range names {
			data[name] = match[i]
		}
	}

	return groups{data: data}
}
