package goterragit

import (
	"fmt"
	"os"
	"path"
)

/*
application:
  author: Olivier Sallou <olivier.sallou@irisa.fr>
  name: "k3s-cluster"
  description: "k3s kubernetes cluster"
  template: "cluster/v1.0"
  recipes:
    recipes_master:
      - "disk_block_automount/v1.0"
      - "k3s-master/v1.0"
    recipes_slave:
      - "k3s-slave/v1.0"
  tags:
    - "k3s"
    - "cluster"
*/

// Application defined a cloud endpoint
type Application struct {
	Author      string              `yaml:"author"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Template    string              `yaml:"template"`
	Recipes     map[string][]string `yaml:"recipes"`
	Tags        []string            `yaml:"tags"`
	Path        string
}

// Check validates a recipe
func (r *Application) Check() error {
	if r.Name == "" {
		return fmt.Errorf("Missing name")
	}
	if r.Template == " " {
		return fmt.Errorf("Missing template")
	}

	return nil
}

// ApplicationDefinition containers a recipe definition
type ApplicationDefinition struct {
	Application Application `yaml:"application"`
}

// Endpoint defined a cloud endpoint
type Endpoint struct {
	Author      string            `yaml:"author"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Kind        string            `yaml:"kind"`
	Features    map[string]string `yaml:"features"`
	Inputs      map[string]string `yaml:"inputs"`
	Config      map[string]string `yaml:"config"`
	Images      map[string]string `yaml:"images"`
	Tags        []string          `yaml:"tags"`
	Path        string
}

// Check validates a recipe
func (r *Endpoint) Check() error {
	if r.Name == "" {
		return fmt.Errorf("Missing name")
	}
	if r.Kind == " " {
		return fmt.Errorf("Missing kind")
	}
	if len(r.Config) == 0 {
		return fmt.Errorf("Missing config info")
	}
	if len(r.Images) == 0 {
		return fmt.Errorf("No image mapping defined")

	}

	return nil
}

// EndpointDefinition containers a recipe definition
type EndpointDefinition struct {
	Endpoint Endpoint `yaml:"endpoint"`
}

// Recipe defines the meta info for a recipe
type Recipe struct {
	Author      string            `yaml:"author"`
	License     string            `yaml:"license"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Inputs      map[string]string `yaml:"inputs"`
	Tags        []string          `yaml:"tags"`
	Base        []string          `yaml:"base"`
	Parent      string            `yaml:"parent"`
	Path        string
}

// Check validates a recipe
func (r *Recipe) Check() error {
	if r.Name == "" {
		return fmt.Errorf("Missing name")
	}
	if r.License == "" {
		return fmt.Errorf("Missing license")
	}
	if r.Inputs == nil {
		r.Inputs = make(map[string]string)
	}
	if r.Tags == nil {
		r.Tags = make([]string, 0)
	}
	if (r.Base == nil || len(r.Base) == 0) && r.Parent == "" {
		return fmt.Errorf("Both base and parent are empty")
	}
	if r.Parent != "" {
		parentRecipe := fmt.Sprintf("%s/recipe.yaml", r.Parent)
		if _, err := os.Stat(parentRecipe); err != nil {
			return fmt.Errorf("Parent recipe %s does not exists", parentRecipe)
		}
	}
	return nil
}

// RecipeDefinition containers a recipe definition
type RecipeDefinition struct {
	Recipe Recipe `yaml:"recipe"`
}

// Template defines the meta info for a template
type Template struct {
	Author      string            `yaml:"author"`
	License     string            `yaml:"license"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Inputs      map[string]string `yaml:"inputs"`
	Tags        []string          `yaml:"tags"`
	Files       map[string]string `yaml:"files"`
	Path        string
	Recipes     []string `yaml:"recipes"`
}

// TemplateDefinition containers a template definition
type TemplateDefinition struct {
	Template Template `yaml:"template"`
}

// Check validates a recipe
func (r *Template) Check() error {
	if r.Name == "" {
		return fmt.Errorf("Missing name")
	}
	if r.License == "" {
		return fmt.Errorf("Missing license")
	}
	if r.Inputs == nil {
		r.Inputs = make(map[string]string)
	}
	if r.Tags == nil {
		r.Tags = make([]string, 0)
	}
	if r.Files == nil || len(r.Files) == 0 {
		return fmt.Errorf("no files specified")
	}

	for cloud, file := range r.Files {
		filePath := fmt.Sprintf("%s/%s/%s", path.Dir(r.Path), cloud, file)
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("File %s does not exists", file)
		}

	}

	return nil
}
