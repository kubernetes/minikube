package mcnflag

import "fmt"

type Flag interface {
	fmt.Stringer
	Default() interface{}
}

type StringFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  string
}

// TODO: Could this be done more succinctly using embedding?
func (f StringFlag) String() string {
	return f.Name
}

func (f StringFlag) Default() interface{} {
	return f.Value
}

type StringSliceFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  []string
}

// TODO: Could this be done more succinctly using embedding?
func (f StringSliceFlag) String() string {
	return f.Name
}

func (f StringSliceFlag) Default() interface{} {
	return f.Value
}

type IntFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  int
}

// TODO: Could this be done more succinctly using embedding?
func (f IntFlag) String() string {
	return f.Name
}

func (f IntFlag) Default() interface{} {
	return f.Value
}

type BoolFlag struct {
	Name   string
	Usage  string
	EnvVar string
}

// TODO: Could this be done more succinctly using embedding?
func (f BoolFlag) String() string {
	return f.Name
}

func (f BoolFlag) Default() interface{} {
	return nil
}
