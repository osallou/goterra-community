package goterragit

import (
	"fmt"
	"os"
	"path"

	terraModel "github.com/osallou/goterra-lib/lib/model"
	"github.com/rs/zerolog/log"
)

// Application defined a cloud endpoint
type Application struct {
	Author      string              `yaml:"author"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Template    string              `yaml:"template"`
	Recipes     map[string][]string `yaml:"recipes"`
	Tags        []string            `yaml:"tags"`
	Path        string
	Defaults    map[string][]string `yaml:"defaults"`
}

// checkRecipeImage checks (sub)recipe exists, returns base image of recipe
func checkRecipeImage(recdb terraModel.Recipe, recipes map[string]terraModel.Recipe) ([]string, error) {
	if recdb.ParentRecipe != "" {
		parentFound := false
		var parentRecipe terraModel.Recipe
		if recipe, ok := recipes[recdb.ParentRecipe]; ok {
			parentRecipe = recipe
			parentFound = true
		}
		if !parentFound {
			return nil, fmt.Errorf("parent recipe not found %s", recdb.ParentRecipe)
		}
		parentImage, err := checkRecipeImage(parentRecipe, recipes)
		if err != nil {
			return nil, err
		}
		return parentImage, nil
	}
	if recdb.BaseImages == nil || len(recdb.BaseImages) == 0 {
		return nil, fmt.Errorf("recipe has no base image nor parent recipe")
	}
	return recdb.BaseImages, nil

}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	//Remove dups from slice.
	inter = removeDups(inter)
	return
}

//Remove dups from slice.
func removeDups(elements []string) (nodups []string) {
	encountered := make(map[string]bool)
	for _, element := range elements {
		if !encountered[element] {
			nodups = append(nodups, element)
			encountered[element] = true
		}
	}
	return
}

// GetAppBaseImages finds intersection of possible images between recipes
func (r *Application) GetAppBaseImages(appRecipes []terraModel.Recipe, recipes map[string]terraModel.Recipe) ([]string, error) {
	possibleBaseImagesNew := true
	//possibleBaseImagesSet := make(map[string]bool, 0)
	possibleBaseImages := make([]string, 0)
	for _, recInfo := range appRecipes {
		parentBaseImages, parentErr := checkRecipeImage(recInfo, recipes)
		if parentErr != nil {
			return nil, parentErr
		}

		//gotACommonBaseImage := false
		// populating for first recipe
		if possibleBaseImagesNew {
			possibleBaseImages = parentBaseImages
			possibleBaseImagesNew = false
			log.Debug().Msgf("bases %+v", parentBaseImages)
			/*
				gotACommonBaseImage = true
				possibleBaseImagesNew = false
				for _, availableImage := range parentBaseImages {
					possibleBaseImagesSet[availableImage] = true
				}
			*/
		} else {
			possibleBaseImagesNew = false
			possibleBaseImages = intersection(possibleBaseImages, parentBaseImages)
			log.Debug().Msgf("bases %+v", parentBaseImages)
			log.Debug().Msgf("intersect %+v", possibleBaseImages)

			/*
				for _, availableImage := range parentBaseImages {
					if _, ok := possibleBaseImagesSet[availableImage]; ok {
						gotACommonBaseImage = true
					} else {
						possibleBaseImagesSet[availableImage] = false
					}
				}*/
		}

		//if !gotACommonBaseImage {
		if len(possibleBaseImages) == 0 {
			// No common base image in recipes
			return nil, fmt.Errorf("No common base image in recipes")
		}

	}

	/*
		for k, v := range possibleBaseImagesSet {
			if v {
				possibleBaseImages = append(possibleBaseImages, k)
			}
		}
	*/
	return possibleBaseImages, nil
}

// Check validates a recipe
func (r *Application) Check() ([]string, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("Missing name")
	}
	if r.Template == " " {
		return nil, fmt.Errorf("Missing template")
	}
	expectedRecipes := make([]string, 0)
	if r.Recipes != nil {
		for _, recipes := range r.Recipes {
			expectedRecipes = append(expectedRecipes, recipes...)
		}
	}

	return expectedRecipes, nil
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
	Defaults    map[string][]string `yaml:"defaults"`
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
	Defaults    map[string][]string `yaml:"defaults"`
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
	Recipes     []string            `yaml:"recipes"`
	Defaults    map[string][]string `yaml:"defaults"`
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
