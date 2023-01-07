package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

var (
	dockerfileTemplate = template.Must(
		template.ParseFiles("templates/dockerfile-local.tmpl"),
	)
)

type DockerContext struct {
	Name string
}

func writeDockerfile(name, dir string) error {
	context := DockerContext{
		Name: name,
	}

	file, err := os.Create(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		return fmt.Errorf("cannot open Dockerfile path %s: %w", filepath.Join(dir, "Dockerfile"), err)
	}
	defer file.Close()

	err = dockerfileTemplate.Execute(file, context)
	if err != nil {
		return fmt.Errorf("cannot generate Dockerfile template: %w", err)
	}

	return nil
}
