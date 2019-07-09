package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	terraGitModel "github.com/osallou/goterra-community/tools/model"
	terraModel "github.com/osallou/goterra-lib/lib/model"
)

func findFiles(targetDir string, pattern string) (files []string, err error) {
	files = make([]string, 0)
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if info.Name() == pattern {
			files = append(files, path)
		}
		return nil
	})
	return files, nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("USAGE : %s <target_directory>\n", os.Args[0])
		os.Exit(1)
	}
	targetDirectory := os.Args[1]

	files, err := findFiles(targetDirectory, "recipe.yaml")
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	hasError := false

	foundRecipes := make(map[string]terraModel.Recipe)

	for _, f := range files {
		fmt.Printf("found %s\n", f)
		yamlRecipe, _ := ioutil.ReadFile(f)
		t := terraGitModel.RecipeDefinition{}
		t.Recipe.Path = f
		err := yaml.Unmarshal(yamlRecipe, &t)
		if err != nil {
			fmt.Printf("Check:recipe:%s:error:, %s", t.Recipe.Path, err)
			hasError = true
		}
		errCheck := t.Recipe.Check()
		if errCheck != nil {
			fmt.Printf("Check:recipe:%s:ko: %s\n", t.Recipe.Name, errCheck)
			hasError = true
		}
		fmt.Printf("Check:recipe:%s:ok\n", t.Recipe.Name)

		t.Recipe.Path = f
		elts := strings.Split(t.Recipe.Path, "/")
		name := elts[len(elts)-3]
		version := elts[len(elts)-2]
		recipe := terraModel.Recipe{
			Remote:        name,
			RemoteVersion: version,
			BaseImages:    t.Recipe.Base,
			ParentRecipe:  t.Recipe.Parent,
		}
		foundRecipes[fmt.Sprintf("%s/%s", name, version)] = recipe

	}

	files, err = findFiles(targetDirectory, "template.yaml")
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	for _, f := range files {
		fmt.Printf("found %s\n", f)
		yamlTemplate, _ := ioutil.ReadFile(f)
		t := terraGitModel.TemplateDefinition{}
		t.Template.Path = f
		err := yaml.Unmarshal(yamlTemplate, &t)
		if err != nil {
			fmt.Printf("Chec:template:%s:error: %s", t.Template.Path, err)
			hasError = true
		}
		errCheck := t.Template.Check()
		if errCheck != nil {
			fmt.Printf("Check:template:%s:error: %s\n", t.Template.Name, errCheck)
			hasError = true
		}
		fmt.Printf("Check:template:%s:ok\n", t.Template.Name)
	}

	files, err = findFiles(targetDirectory, "endpoint.yaml")
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	for _, f := range files {
		fmt.Printf("found %s\n", f)
		yamlTemplate, _ := ioutil.ReadFile(f)
		t := terraGitModel.EndpointDefinition{}
		t.Endpoint.Path = f
		err := yaml.Unmarshal(yamlTemplate, &t)
		if err != nil {
			fmt.Printf("Chec:endpoint:%s:error: %s", t.Endpoint.Path, err)
			hasError = true
		}
		errCheck := t.Endpoint.Check()
		if errCheck != nil {
			fmt.Printf("Check:endpoint:%s:error: %s\n", t.Endpoint.Name, errCheck)
			hasError = true
		}
		fmt.Printf("Check:endpoint:%s:ok\n", t.Endpoint.Name)
	}

	files, err = findFiles(targetDirectory, "app.yaml")
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	for _, f := range files {
		fmt.Printf("found %s\n", f)
		yamlApp, _ := ioutil.ReadFile(f)
		t := terraGitModel.ApplicationDefinition{}
		t.Application.Path = f
		err := yaml.Unmarshal(yamlApp, &t)
		if err != nil {
			fmt.Printf("Chec:application:%s:error: %s", t.Application.Path, err)
			hasError = true
		}
		expectedRecipes, errCheck := t.Application.Check()
		if errCheck != nil {
			fmt.Printf("Check:application:%s:error: %s\n", t.Application.Name, errCheck)
			hasError = true
		}
		appRecipes := make([]terraModel.Recipe, 0)
		for _, expectedRecipe := range expectedRecipes {
			recipeFile := fmt.Sprintf("%s/recipes/%s/recipe.yaml", targetDirectory, expectedRecipe)
			fmt.Printf("Check:application:%s:check:needs:%s\n", t.Application.Name, recipeFile)
			if _, ok := os.Stat(recipeFile); ok != nil {
				fmt.Printf("Check:application:%s:needs:%s:error:notfound\n", t.Application.Name, recipeFile)
				hasError = true
			}
			appRecipes = append(appRecipes, foundRecipes[expectedRecipe])
		}

		_, baseErr := t.Application.GetAppBaseImages(appRecipes, foundRecipes)
		if baseErr != nil {
			fmt.Printf("Check:application:%s:base_image:no base image found\n", t.Application.Name)
			hasError = true
		}

		templateFile := fmt.Sprintf("%s/templates/%s/template.yaml", targetDirectory, t.Application.Template)
		fmt.Printf("Check:application:%s:check:needs:%s\n", t.Application.Name, t.Application.Template)
		if _, ok := os.Stat(templateFile); ok != nil {
			fmt.Printf("Check:application:%s:needs:%s:error:notfound\n", t.Application.Name, templateFile)
			hasError = true
		}

		if hasError {
			fmt.Printf("Check:application:%s:ko\n", t.Application.Name)

		} else {
			fmt.Printf("Check:application:%s:ok\n", t.Application.Name)
		}
	}

	if hasError {
		os.Exit(1)
	}
}
