package main

type Config struct {
	Tools []Tool
}

type Tool struct {
	Name   string
	Params []string
	Args   []string
}
