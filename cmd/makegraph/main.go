package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// LernaJSON package.json
type LernaJSON struct {
	Packages []string
}

// PackageJSON package.json
type PackageJSON struct {
	Name            string
	Dependencies    map[string]string
	DevDependencies map[string]string
	Scripts         map[string]string
}

func readLernaJSON(path string) (*LernaJSON, error) {
	fileContents, error := ioutil.ReadFile(path)

	if error != nil {
		return nil, error
	}

	var lernaJSON LernaJSON

	error = json.Unmarshal(fileContents, &lernaJSON)

	return &lernaJSON, error
}

func readPackageJSON(path string) (*PackageJSON, error) {
	fileContents, error := ioutil.ReadFile(path)

	if error != nil {
		return nil, error
	}

	var packageJSON PackageJSON

	error = json.Unmarshal(fileContents, &packageJSON)

	return &packageJSON, error
}

func main() {
	projectDependencyMap, error := makeProjectDependencyMap(options{
		Context: "./mock-monorepo",
	})
	handleError(error)

	funcMap := template.FuncMap{
		"slugify": slugify,
	}

	digraphTemplate := template.Must(template.New("graph.gotmpl").Funcs(funcMap).ParseFiles("./graph.gotmpl"))
	// handleError(error)

	error = digraphTemplate.Execute(os.Stdout, projectDependencyMap)
	handleError(error)
}

type packageJSONDescriptor struct {
	Group            string
	PackageJSONPaths []string
}

type options struct {
	Context string
}

func getPackages(options options) ([]packageJSONDescriptor, error) {

	lernaJSON, error := readLernaJSON(path.Join(options.Context, "lerna.json"))
	handleError(error)

	var packageJSONPaths = []packageJSONDescriptor{}

	for _, lernaProjectGlob := range lernaJSON.Packages {
		absolutePath, error := filepath.Abs(filepath.Join(options.Context, lernaProjectGlob, "package.json"))

		if error != nil {
			return nil, error
		}

		matches, error := filepath.Glob(absolutePath)

		if error != nil {
			return nil, error
		}

		packageJSONPaths = append(packageJSONPaths, packageJSONDescriptor{
			Group:            lernaProjectGlob,
			PackageJSONPaths: matches,
		})
	}
	return packageJSONPaths, nil
}

func resolvePath(partialPath string, options options) (string, error) {
	return filepath.Abs(filepath.Join(options.Context, partialPath))
}

type packageDescriptor struct {
	Name         string
	Group        string
	Flags        map[string]bool
	Dependencies []string
}

func makeProjectDependencyMap(options options) ([]packageDescriptor, error) {

	allPackageJSONPaths, error := getPackages(options)

	if error != nil {
		return nil, error
	}

	// build a list of packages and deps
	var packageDependencyMap = []packageDescriptor{}

	for _, group := range allPackageJSONPaths {
		for _, packageJSONPath := range group.PackageJSONPaths {
			packageJSON, error := readPackageJSON(packageJSONPath)

			if error != nil {
				return nil, error
			}

			dependencies, _ := getKeysAndValues(packageJSON.Dependencies)
			devDependencies, _ := getKeysAndValues(packageJSON.DevDependencies)

			packageDependencyMap = append(packageDependencyMap, packageDescriptor{
				Name:  packageJSON.Name,
				Group: group.Group,
				Flags: map[string]bool{
					"docker": false,
				},
				Dependencies: append(
					dependencies,
					devDependencies...,
				),
			})
		}
	}

	// get all internal project names
	var internalProjectNames = []string{}
	for _, project := range packageDependencyMap {
		internalProjectNames = append(internalProjectNames, project.Name)
	}

	// filter those deps by internal projects
	// const filterFn = internalProjectNames.includes.bind(internalProjectNames)

	var filterFn = func(projectName string) bool {
		return true
	}

	for _, project := range packageDependencyMap {
		project.Dependencies = filter(project.Dependencies, filterFn)
	}

	return packageDependencyMap, nil
}

func getKeysAndValues(myMap map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(myMap))
	values := make([]string, 0, len(myMap))

	for k, v := range myMap {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func handleError(error error) {
	if error != nil {
		fmt.Println(error)
		os.Exit(1)
	}
}

func slugify(input string) string {
	return strings.Replace(strings.Replace(input, "@", "", -1), "/", "", -1)
}
