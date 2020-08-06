//+build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

//var Default = Docker.Build

// Build runs a full build.
func Build() error {
	mg.Deps(Test)
	if err := sh.RunV("go", "build", "-v"); err != nil {
		return err
	}
	return nil
}

// Test runs the complete test suite.
func Test() error {
	return sh.RunV("go", "test", "-v", "./...")
}

type Docker mg.Namespace

// Build creates a docker container.
func (Docker) Build() error {
	mg.Deps(Test)
	return sh.Run("docker", "build", "-t", "mwmahlberg/mongen:latest", ".")
}

func (Docker) Push() error {
	return sh.Run("docker", "push", "mwmahlberg/mongen:latest")
}
