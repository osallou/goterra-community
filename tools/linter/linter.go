package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	terraGitModel "github.com/osallou/goterra-community/tools/model"
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

	if hasError {
		os.Exit(1)
	}
}
