package main

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
	Desc     []string   `json:"desc"`
	Required []Argument `json:"required"`
	Optional []Argument `json:"optional"`
	Command  []string   `json:"command"`
}
