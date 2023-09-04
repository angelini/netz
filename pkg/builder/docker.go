package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"go.uber.org/zap"
)

var (
	baseDockerfileTemplate = template.Must(
		template.ParseFiles("templates/base-dockerfile.tmpl"),
	)
	dockerfileTemplate = template.Must(
		template.ParseFiles("templates/dockerfile.tmpl"),
	)
)

type BaseDockerContext struct{}

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
		return fmt.Errorf("cannot render Dockerfile template: %w", err)
	}

	return nil
}

func BuildBase(log *zap.Logger, distDir string) error {
	context := BaseDockerContext{}

	dir := filepath.Join(distDir, "base")
	log.Info("building base files", zap.String("dir", dir))

	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return fmt.Errorf("cannot create output dir: %s", dir)
	}

	file, err := os.Create(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		return fmt.Errorf("cannot open Dockerfile path %s: %w", filepath.Join(dir, "Dockerfile"), err)
	}
	defer file.Close()

	err = baseDockerfileTemplate.Execute(file, context)
	if err != nil {
		return fmt.Errorf("cannot render Dockerfile template: %w", err)
	}

	return nil
}
