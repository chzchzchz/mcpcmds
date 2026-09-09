package main

import "strings"

type Config struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Tools []Tool `json:"tools"`
}

type Argument struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type Tool struct {
	Name     string     `json:"name"`
	Title    string     `json:"title"`
	Desc     string     `json:"desc"`
	Required []Argument `json:"required"`
	Optional []Argument `json:"optional"`
	Command  []string   `json:"command"`
}

func replacePlaceholders(s string, args map[string]string) string {
	var b strings.Builder
	for {
		start := strings.Index(s, "${")
		if start == -1 {
			b.WriteString(s)
			break
		}
		end := strings.Index(s[start+2:], "}")
		if end == -1 {
			b.WriteString(s)
			break
		}
		key := s[start+2 : start+2+end]
		b.WriteString(s[:start])
		if val, ok := args[key]; ok {
			b.WriteString(val)
		}
		s = s[start+2+end+1:]
	}
	return b.String()
}
