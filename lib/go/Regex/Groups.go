package Regex

import "fmt"

type groups struct {
	data map[string]string
}

func (g *groups) Get(key string) (result string, err error) {
	if value, exists := g.data[key]; exists {
		result = value
	} else {
		err = fmt.Errorf("Group %s doesn't exists", key)
	}

	return result, err
}
