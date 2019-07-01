package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	terraConfig "github.com/osallou/goterra-lib/lib/config"
	"github.com/rs/cors"
	mongo "go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gopkg.in/src-d/go-git.v4"

	terraModel "github.com/osallou/goterra-lib/lib/model"
)

// Version of server
var Version string

var mongoClient mongo.Client
var nsCollection *mongo.Collection
var recipeCollection *mongo.Collection
var templateCollection *mongo.Collection

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
	pullErr := pull(workTree)
	if pullErr != nil {
		os.Exit(1)
	}

	for true {
		log.Info().Msg("Try to inject new/updated recipes")
		pullErr := pull(workTree)
		if pullErr != nil {
			log.Error().Msgf("Failed to pull files")
		}
		files, err := findFiles(gitDir+"/recipes", "recipe.yaml")
		if err != nil {
			log.Error().Msgf("failed to search recipes: %s", err)
		}
		//TODO manage recipes
		for _, file := range files {

		}

		files, err = findFiles(gitDir+"/templates", "template.yaml")
		if err != nil {
			log.Error().Msgf("failed to search templates: %s", err)
		}
		//TODO manage templates

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

	recipe := terraModel.Recipe{}

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
