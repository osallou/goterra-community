package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	terraConfig "github.com/osallou/goterra-lib/lib/config"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"

	terraGitModel "github.com/osallou/goterra-community/tools/model"
	terraModel "github.com/osallou/goterra-lib/lib/model"
)

// Version of server
var Version string

var mongoClient mongo.Client
var nsCollection *mongo.Collection
var recipeCollection *mongo.Collection
var templateCollection *mongo.Collection
var endpointCollection *mongo.Collection
var appCollection *mongo.Collection

func pull(workTree *git.Worktree) error {
	pullOptions := git.PullOptions{}
	log.Info().Msg("git pull")
	err := workTree.Pull(&pullOptions)
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			log.Error().Msgf("Git pull error: %s", err)
			return err
		}
	}
	return nil
}

func findFiles(targetDir string, pattern string) (files []string, err error) {
	files = make([]string, 0)
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		if info.Name() == pattern {
			files = append(files, path)
		}
		return nil
	})
	return files, nil
}

// getNS returns namespace id, creates it if not present
func getNS() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var nsdb terraModel.NSData
	req := bson.M{
		"name": "goterra",
	}
	err := nsCollection.FindOne(ctx, req).Decode(&nsdb)
	if err != nil {
		// Create it
		ns := bson.M{
			"name":    "goterra",
			"owners":  make([]string, 0),
			"members": make([]string, 0),
		}
		newns, err := nsCollection.InsertOne(ctx, ns)
		if err != nil {
			log.Error().Msg("Failed to create namespace")
			return "", err
		}
		return newns.InsertedID.(primitive.ObjectID).Hex(), nil
	}
	return nsdb.ID.Hex(), nil
}

func getRecipe(ns string, name string, version string) (*terraModel.Recipe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var recipe terraModel.Recipe
	req := bson.M{
		"namespace":     ns,
		"remote":        name,
		"remoteversion": version,
	}
	log.Info().Msgf("Search recipe %s:%s in ns %s", name, version, ns)
	err := recipeCollection.FindOne(ctx, req).Decode(&recipe)
	if err != nil {
		log.Info().Msgf("error => %s", err)
		return nil, err
	}
	return &recipe, nil
}

func updateRecipe(ns string, recipe *terraModel.Recipe) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := bson.M{
		"_id": recipe.ID,
	}
	_, err := recipeCollection.ReplaceOne(ctx, req, recipe)
	if err != nil {
		log.Error().Msgf("Failed to update recipe %s", err)
	}
}

func createRecipe(ns string, recipe *terraModel.Recipe) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	newRecipe, err := recipeCollection.InsertOne(ctx, recipe)
	if err != nil {
		log.Error().Msgf("Failed to create recipe %+v", recipe)
		return "", err
	}
	return newRecipe.InsertedID.(primitive.ObjectID).Hex(), nil
}

func getTemplate(ns string, name string, version string) (*terraModel.Template, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var template terraModel.Template
	req := bson.M{
		"namespace":     ns,
		"remote":        name,
		"remoteversion": version,
	}
	log.Info().Msgf("Search template %s:%s in ns %s", name, version, ns)
	err := templateCollection.FindOne(ctx, req).Decode(&template)
	if err != nil {
		log.Info().Msgf("error => %s", err)
		return nil, err
	}
	return &template, nil
}

func updateTemplate(ns string, template *terraModel.Template) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := bson.M{
		"_id": template.ID,
	}
	_, err := templateCollection.ReplaceOne(ctx, req, template)
	if err != nil {
		log.Error().Msgf("Failed to update template %s", err)
	}
}

func createTemplate(ns string, template *terraModel.Template) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	newTemplate, err := templateCollection.InsertOne(ctx, template)
	if err != nil {
		log.Error().Msgf("Failed to create template %+v", template)
		return "", err
	}
	return newTemplate.InsertedID.(primitive.ObjectID).Hex(), nil
}

func getEndpoint(ns string, name string) (*terraModel.EndPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var endpoint terraModel.EndPoint
	req := bson.M{
		"namespace": ns,
		"remote":    name,
	}
	log.Info().Msgf("Search endpoint %s in ns %s", name, ns)
	err := endpointCollection.FindOne(ctx, req).Decode(&endpoint)
	if err != nil {
		log.Info().Msgf("error => %s", err)
		return nil, err
	}
	return &endpoint, nil
}

func updateEndpoint(ns string, endpoint *terraModel.EndPoint) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := bson.M{
		"_id": endpoint.ID,
	}
	_, err := endpointCollection.ReplaceOne(ctx, req, endpoint)
	if err != nil {
		log.Error().Msgf("Failed to update endpoint %s", err)
	}
}

func createEndpoint(ns string, endpoint *terraModel.EndPoint) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	newEndpoint, err := endpointCollection.InsertOne(ctx, endpoint)
	if err != nil {
		log.Error().Msgf("Failed to create endpoint %+v", endpoint)
		return "", err
	}
	return newEndpoint.InsertedID.(primitive.ObjectID).Hex(), nil
}

func getApplication(ns string, name string, version string) (*terraModel.Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var application terraModel.Application
	req := bson.M{
		"namespace":     ns,
		"remote":        name,
		"remoteversion": version,
	}
	log.Info().Msgf("Search application %s in ns %s", name, ns)
	err := appCollection.FindOne(ctx, req).Decode(&application)
	if err != nil {
		log.Info().Msgf("error => %s", err)
		return nil, err
	}
	return &application, nil
}

func updateApplication(ns string, application *terraModel.Application) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := bson.M{
		"_id": application.ID,
	}
	_, err := appCollection.ReplaceOne(ctx, req, application)
	if err != nil {
		log.Error().Msgf("Failed to update application %s", err)
	}
}

func createApplication(ns string, application *terraModel.Application) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	newApplication, err := appCollection.InsertOne(ctx, application)
	if err != nil {
		log.Error().Msgf("Failed to create application %+v", application)
		return "", err
	}
	return newApplication.InsertedID.(primitive.ObjectID).Hex(), nil
}

func injector() {
	config := terraConfig.LoadConfig()
	gitDir := "/tmp/goterra-git"
	var repo *git.Repository
	var err error

	if _, ok := os.Stat(gitDir); ok != nil {
		repo, err = git.PlainClone(gitDir, false, &git.CloneOptions{
			URL:      config.Git,
			Progress: os.Stdout,
		})
		if err != nil {
			log.Error().Msgf("Git clone error: %s", err)
			if err != git.ErrRepositoryAlreadyExists {
				os.Exit(1)
			}
		}
	} else {
		repo, err = git.PlainOpen(gitDir)
		if err != nil {
			os.Exit(1)
		}
	}

	workTree, _ := repo.Worktree()

	ns, nserr := getNS()
	if nserr != nil {
		os.Exit(1)
	}

	for true {
		log.Info().Msg("Try to inject new/updated recipes")

		createdRecipes := make(map[string]string)
		foundRecipes := make(map[string]terraModel.Recipe)
		createdTemplates := make(map[string]string)

		if os.Getenv("GOT_PULL_SKIP") != "1" {
			pullErr := pull(workTree)
			if pullErr != nil {
				log.Error().Msgf("Failed to pull files")
				time.Sleep(10 * time.Minute)
				continue
			}
		}
		files, err := findFiles(gitDir+"/recipes", "recipe.yaml")
		if err != nil {
			log.Error().Msgf("failed to search recipes: %s", err)
			time.Sleep(10 * time.Minute)
			continue
		}
		for _, file := range files {
			yamlRecipe, _ := ioutil.ReadFile(file)
			t := terraGitModel.RecipeDefinition{}
			t.Recipe.Path = file
			err := yaml.Unmarshal(yamlRecipe, &t)
			if err != nil {
				log.Error().Msgf("Failed to read %s", file)
				continue
			}
			errCheck := t.Recipe.Check()
			if errCheck != nil {
				log.Error().Msgf("Recipe did not pass the check!  %s", t.Recipe.Path)
				continue
			}
			elts := strings.Split(t.Recipe.Path, "/")
			name := elts[len(elts)-3]
			version := elts[len(elts)-2]
			recipe, rerr := getRecipe(ns, name, version)
			if rerr != nil {
				// Does not exists
				log.Debug().Msgf("Recipe does not exists %s:%s", name, version)
				recipe = &terraModel.Recipe{}
				recipe.Name = t.Recipe.Name
				recipe.BaseImages = t.Recipe.Base
				recipe.Tags = t.Recipe.Tags
				recipe.Timestamp = time.Now().Unix()
				recipe.Inputs = t.Recipe.Inputs
				recipe.Namespace = ns
				recipe.Description = t.Recipe.Description
				recipe.Public = true
				script, scriptErr := ioutil.ReadFile(fmt.Sprintf("%s/recipes/%s/%s/recipe.sh", gitDir, name, version))
				recipe.Remote = name
				recipe.RemoteVersion = version
				recipe.Version = version
				if scriptErr != nil {
					log.Warn().Msgf("Could not read recipe script %s", t.Recipe.Path)
					continue
				}
				recipe.Script = string(script)
				recipe.ParentRecipe = ""
				if t.Recipe.Parent != "" {
					elts := strings.Split(t.Recipe.Parent, "/")
					if len(elts) != 2 {
						log.Error().Msgf("Invalid parent")
						continue
					}
					parentInDB, parentErr := getRecipe(ns, elts[0], elts[1])
					if parentErr != nil {
						log.Error().Msgf("could not find parent in db for %s, skipping... may be an ordering issue", t.Recipe.Path)
						continue
					}
					recipe.ParentRecipe = parentInDB.ID.Hex()
				}
				id, newErr := createRecipe(ns, recipe)
				if newErr == nil {
					createdRecipes[name+"/"+version] = id
				}
			} else {
				// Exists
				log.Debug().Msgf("Recipe exists %s:%s", name, version)
				recipe.Name = t.Recipe.Name
				recipe.BaseImages = t.Recipe.Base
				recipe.Tags = t.Recipe.Tags
				recipe.Timestamp = time.Now().Unix()
				recipe.Inputs = t.Recipe.Inputs
				recipe.Namespace = ns
				recipe.Description = t.Recipe.Description
				recipe.Public = true
				recipe.Version = version
				script, scriptErr := ioutil.ReadFile(fmt.Sprintf("%s/recipes/%s/%s/recipe.sh", gitDir, name, version))
				if scriptErr != nil {
					log.Error().Msgf("Could not read recipe script %s", t.Recipe.Path)
					continue
				}
				recipe.Script = string(script)
				recipe.ParentRecipe = ""
				if t.Recipe.Parent != "" {
					elts := strings.Split(t.Recipe.Parent, "/")
					if len(elts) != 2 {
						log.Error().Msgf("Invalid parent")
						continue
					}
					parentInDB, parentErr := getRecipe(ns, elts[0], elts[1])
					if parentErr != nil {
						log.Error().Msgf("could not find parent in db for %s, skipping... may be an ordering issue", t.Recipe.Path)
						continue
					}
					recipe.ParentRecipe = parentInDB.ID.Hex()
				}
				updateRecipe(ns, recipe)
				createdRecipes[name+"/"+version] = recipe.ID.Hex()
			}
			tmpRecipe := terraModel.Recipe{
				Remote:        name,
				RemoteVersion: version,
				BaseImages:    t.Recipe.Base,
				ParentRecipe:  t.Recipe.Parent,
			}
			foundRecipes[fmt.Sprintf("%s/%s", name, version)] = tmpRecipe

		}

		files, err = findFiles(gitDir+"/templates", "template.yaml")
		if err != nil {
			log.Error().Msgf("failed to search templates: %s", err)
		}
		for _, file := range files {
			yamlTemplate, _ := ioutil.ReadFile(file)
			t := terraGitModel.TemplateDefinition{}
			t.Template.Path = file
			err := yaml.Unmarshal(yamlTemplate, &t)
			if err != nil {
				log.Error().Msgf("Failed to read %s", file)
				continue
			}
			errCheck := t.Template.Check()
			if errCheck != nil {
				log.Error().Msgf("Template did not pass the check!  %s", t.Template.Path)
				continue
			}

			elts := strings.Split(t.Template.Path, "/")
			name := elts[len(elts)-3]
			version := elts[len(elts)-2]
			template, rerr := getTemplate(ns, name, version)
			if rerr != nil {
				// Does not exists
				log.Debug().Msgf("Template does not exists %s:%s", name, version)
				template = &terraModel.Template{}
				template.Name = t.Template.Name
				template.Tags = t.Template.Tags
				template.Timestamp = time.Now().Unix()
				template.Inputs = t.Template.Inputs
				template.Namespace = ns
				template.Description = t.Template.Description
				template.Public = true
				template.Remote = name
				template.RemoteVersion = version
				template.Version = version
				template.Data = make(map[string]string)
				for cloud, file := range t.Template.Files {
					scriptFile := fmt.Sprintf("%s/templates/%s/%s/%s/%s", gitDir, name, version, cloud, file)
					script, scriptErr := ioutil.ReadFile(scriptFile)
					if scriptErr != nil {
						log.Error().Msgf("Could not read template script %s: %s", t.Template.Path, scriptFile)
						continue
					}
					template.Data[cloud] = string(script)
				}
				if t.Template.Recipes == nil {
					t.Template.Recipes = make([]string, 0)
				}
				template.VarRecipes = t.Template.Recipes
				id, newErr := createTemplate(ns, template)
				if newErr == nil {
					createdRecipes[name+"/"+version] = id
				}
			} else {
				// Exists
				log.Debug().Msgf("Template exists %s:%s", name, version)
				template.Name = t.Template.Name
				template.Tags = t.Template.Tags
				template.Timestamp = time.Now().Unix()
				template.Inputs = t.Template.Inputs
				template.Namespace = ns
				template.Description = t.Template.Description
				template.Public = true
				template.Version = version
				template.Data = make(map[string]string)
				for cloud, file := range t.Template.Files {
					scriptFile := fmt.Sprintf("%s/templates/%s/%s/%s/%s", gitDir, name, version, cloud, file)
					script, scriptErr := ioutil.ReadFile(scriptFile)
					if scriptErr != nil {
						log.Error().Msgf("Could not read template script %s: %s", t.Template.Path, scriptFile)
						continue
					}
					template.Data[cloud] = string(script)
				}
				if t.Template.Recipes == nil {
					t.Template.Recipes = make([]string, 0)
				}
				template.VarRecipes = t.Template.Recipes
				updateTemplate(ns, template)
				createdTemplates[name+"/"+version] = template.ID.Hex()
			}

		}

		files, err = findFiles(gitDir+"/endpoints", "endpoint.yaml")
		if err != nil {
			log.Error().Msgf("failed to search endpoints: %s", err)
		}
		for _, file := range files {
			yamlEndpoint, _ := ioutil.ReadFile(file)
			t := terraGitModel.EndpointDefinition{}
			t.Endpoint.Path = file
			err := yaml.Unmarshal(yamlEndpoint, &t)
			if err != nil {
				log.Error().Msgf("Failed to read %s", file)
				continue
			}
			errCheck := t.Endpoint.Check()
			if errCheck != nil {
				log.Error().Msgf("Endpoint did not pass the check!  %s", t.Endpoint.Path)
				continue
			}

			elts := strings.Split(t.Endpoint.Path, "/")
			name := elts[len(elts)-2]
			endpoint, rerr := getEndpoint(ns, name)
			if rerr != nil {
				// Does not exists
				log.Debug().Msgf("Endpoint does not exists %s", name)
				endpoint = &terraModel.EndPoint{}
				endpoint.Name = t.Endpoint.Name
				endpoint.Remote = name
				endpoint.Timestamp = time.Now().Unix()
				endpoint.Namespace = ns
				endpoint.Public = true
				endpoint.Kind = t.Endpoint.Kind

				endpoint.Features = t.Endpoint.Features
				if endpoint.Features == nil {
					endpoint.Features = make(map[string]string)
				}
				endpoint.Inputs = t.Endpoint.Inputs
				if endpoint.Inputs == nil {
					endpoint.Inputs = make(map[string]string)
				}

				endpoint.Config = t.Endpoint.Config
				if endpoint.Config == nil {
					endpoint.Config = make(map[string]string)
				}

				endpoint.Images = t.Endpoint.Images
				if endpoint.Images == nil {
					endpoint.Images = make(map[string]string)
				}

				createEndpoint(ns, endpoint)
			} else {
				// Exists
				log.Debug().Msgf("Endpoint exists %s", name)
				endpoint.Name = t.Endpoint.Name
				endpoint.Remote = name
				endpoint.Timestamp = time.Now().Unix()
				endpoint.Namespace = ns
				endpoint.Public = true
				endpoint.Kind = t.Endpoint.Kind
				endpoint.Features = t.Endpoint.Features
				if endpoint.Features == nil {
					endpoint.Features = make(map[string]string)
				}
				endpoint.Inputs = t.Endpoint.Inputs
				if endpoint.Inputs == nil {
					endpoint.Inputs = make(map[string]string)
				}

				endpoint.Config = t.Endpoint.Config
				if endpoint.Config == nil {
					endpoint.Config = make(map[string]string)
				}

				endpoint.Images = t.Endpoint.Images
				if endpoint.Images == nil {
					endpoint.Images = make(map[string]string)
				}
				updateEndpoint(ns, endpoint)
			}

		}

		files, err = findFiles(gitDir+"/apps", "app.yaml")
		if err != nil {
			log.Error().Msgf("failed to search apps: %s", err)
		}
		for _, file := range files {
			yamlApp, _ := ioutil.ReadFile(file)
			t := terraGitModel.ApplicationDefinition{}
			t.Application.Path = file
			err := yaml.Unmarshal(yamlApp, &t)
			if err != nil {
				log.Error().Msgf("Failed to read %s", file)
				continue
			}

			expectedRecipes, errCheck := t.Application.Check()
			if errCheck != nil {
				log.Error().Msgf("Application did not pass the check!  %s", t.Application.Path)
				continue
			}
			appRecipes := make([]terraModel.Recipe, 0)
			for _, expectedRecipe := range expectedRecipes {
				appRecipes = append(appRecipes, foundRecipes[expectedRecipe])
			}

			baseImages, baseErr := t.Application.GetAppBaseImages(appRecipes, foundRecipes)
			if baseErr != nil {
				log.Error().Msgf("Application %s could not find a base image between recipes", t.Application.Path)
				continue
			}
			log.Error().Msgf("Base images: %+v", baseImages)

			elts := strings.Split(t.Application.Path, "/")
			name := elts[len(elts)-3]
			version := elts[len(elts)-2]
			application, rerr := getApplication(ns, name, version)
			if rerr != nil {
				// Does not exists
				log.Debug().Msgf("Application does not exists %s", name)
				application = &terraModel.Application{}
				application.Image = baseImages
				application.Name = t.Application.Name
				application.Description = t.Application.Description
				application.Remote = name
				application.RemoteVersion = version
				application.Timestamp = time.Now().Unix()
				application.Namespace = ns
				application.Public = true
				application.TemplateRecipes = make(map[string][]string)
				hasError := false
				for tplVar, recipes := range t.Application.Recipes {
					for _, recipe := range recipes {
						if recipeID, ok := createdRecipes[recipe]; ok {
							if application.TemplateRecipes[tplVar] == nil {
								application.TemplateRecipes[tplVar] = make([]string, 0)
							}
							application.TemplateRecipes[tplVar] = append(application.TemplateRecipes[tplVar], recipeID)
						} else {
							log.Error().Msgf("App %s requests recipe %s, but it does not exists!", t.Application.Name, recipe)
							hasError = true
							break
						}

					}
				}
				if hasError {
					continue
				}

				createApplication(ns, application)
			} else {
				// Exists
				log.Debug().Msgf("Application exists %s", name)
				application.Name = t.Application.Name
				application.Image = baseImages
				application.Description = t.Application.Description
				application.Remote = name
				application.RemoteVersion = version
				application.Timestamp = time.Now().Unix()
				application.Namespace = ns
				application.Public = true
				application.TemplateRecipes = make(map[string][]string)
				hasError := false
				for tplVar, recipes := range t.Application.Recipes {
					for _, recipe := range recipes {
						if recipeID, ok := createdRecipes[recipe]; ok {
							if application.TemplateRecipes[tplVar] == nil {
								application.TemplateRecipes[tplVar] = make([]string, 0)
							}
							application.TemplateRecipes[tplVar] = append(application.TemplateRecipes[tplVar], recipeID)
						} else {
							log.Error().Msgf("App %s requests recipe %s, but it does not exists!  %s", t.Application.Name, recipe)
							hasError = true
							break
						}

					}
				}
				if hasError {
					continue
				}
				updateApplication(ns, application)
			}

		}

		// Sleep for one hour
		time.Sleep(1 * time.Hour)
	}

}

// HomeHandler manages base entrypoint
var HomeHandler = func(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{"version": Version, "message": "ok"}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("GOT_DEBUG") != "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	config := terraConfig.LoadConfig()

	consulErr := terraConfig.ConsulDeclare("got-injector", "/injector")
	if consulErr != nil {
		log.Error().Msgf("Failed to register: %s", consulErr.Error())
		panic(consulErr)
	}

	mongoClient, err := mongo.NewClient(mongoOptions.Client().ApplyURI(config.Mongo.URL))
	if err != nil {
		log.Error().Msgf("Failed to connect to mongo server %s", config.Mongo.URL)
		os.Exit(1)
	}
	ctx, cancelMongo := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelMongo()

	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Error().Msgf("Failed to connect to mongo server %s", config.Mongo.URL)
		os.Exit(1)
	}

	nsCollection = mongoClient.Database(config.Mongo.DB).Collection("ns")
	recipeCollection = mongoClient.Database(config.Mongo.DB).Collection("recipe")
	templateCollection = mongoClient.Database(config.Mongo.DB).Collection("template")
	endpointCollection = mongoClient.Database(config.Mongo.DB).Collection("endpoint")
	appCollection = mongoClient.Database(config.Mongo.DB).Collection("application")

	go injector()

	r := mux.NewRouter()
	r.HandleFunc("/injector", HomeHandler).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	})
	handler := c.Handler(r)

	loggedRouter := handlers.LoggingHandler(os.Stdout, handler)

	srv := &http.Server{
		Handler:      loggedRouter,
		Addr:         fmt.Sprintf("%s:%d", config.Web.Listen, config.Web.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	srv.ListenAndServe()

}
