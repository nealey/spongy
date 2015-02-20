package main

import (
	"io/ioutil"
	"path"
	"strings"
)

type Config struct {
	BaseDir string
}

func readString(filename string) (string, error) {
	octets, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(octets)), nil
}


func ReadConfig(cgiDir string) (*Config, error) {
	cfgfn := path.Join(cgiDir, "spongy.cfg")
	basePath, err := readString(cfgfn)
	if err != nil {
		return nil, err
	}
	
	return &Config{basePath}, nil
}

func (c *Config) Get(name string) (string, error) {
	path := path.Join(c.BaseDir, name)
	val, err := readString(path)
	return val, err
}

