package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var versionFileTemplate = template.Must(template.New("version.go").Parse(`// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instana

// Version is the version of Instana sensor
const Version = "{{ .Version }}"
`))

var mainModuleVersionTagRegExp = regexp.MustCompile("^v\\d+\\.\\d+\\.\\d+$")

func main() {

	var execute bool
	flag.BoolVar(&execute, "execute", false, "to run all the operation and write changes to git")

	var releaseType string
	flag.StringVar(&releaseType, "releaseType", "", "bugfix or minor")

	var pathToProjectRoot string
	flag.StringVar(&pathToProjectRoot, "projectRoot", "..", "path to project root")

	flag.Parse()
	if releaseType != "bugfix" && releaseType != "minor" {
		log.Fatalln("please provide proper releaseType `bugfix` or `minor`")
	}

	//invert value for easy understanding
	dryRun := !execute

	path, err := filepath.Abs(pathToProjectRoot)
	CheckIfError(err)
	fmt.Println("path:", path)

	if !strings.HasSuffix(path, "go-sensor") {
		log.Fatalln("wrong path:", path)
	}

	runAndCheckError(exec.Command("git", "-C", path, "checkout", "master"), dryRun)
	runAndCheckError(exec.Command("git", "-C", path, "pull", "--rebase"), dryRun)

	tags := getTags(path)
	filteredTags := filterTags(tags)
	lastVersion, err := recentVersion(filteredTags)
	CheckIfError(err)
	fmt.Println(lastVersion)

	versionFromFile, err := findVersionInFile(path)
	CheckIfError(err)

	fmt.Println(lastVersion, versionFromFile)

	lastVersionObj, err := semver.NewVersion(lastVersion)
	CheckIfError(err)

	versionFromFileObj, err := semver.NewVersion(versionFromFile)
	CheckIfError(err)

	if lastVersionObj.Equal(versionFromFileObj) {
		fmt.Println("versions are equal")
	} else {
		log.Fatalln(fmt.Sprintf("versions are not equal, fix manually %s!=%s", lastVersion, versionFromFile))
	}

	var newVersion string
	switch releaseType {
	case "minor":
		newVersion = lastVersionObj.IncMinor().String()
	case "bugfix":
		newVersion = lastVersionObj.IncPatch().String()
	default:
		log.Fatalln("wrong release type:", releaseType)
	}

	fmt.Println("new version will be:", newVersion)

	//todo: be smarter and extract it somewhere
	file := filepath.Join(path, "version.go")

	fd, err := os.OpenFile(file, os.O_RDWR, 0)
	if err != nil {
		log.Fatalln("error while opening a file:", file)
	}
	defer fd.Close()

	err = versionFileTemplate.Execute(fd, map[string]string{
		"Version": newVersion,
	})
	CheckIfError(err)
	fd.Close()

	runAndCheckError(exec.Command("git", "-C", path, "add", "version.go"), dryRun)
	runAndCheckError(exec.Command("git", "-C", path, "commit", "-m", fmt.Sprintf(`"Bump go-sensor version to %s"`, newVersion)), dryRun)
	runAndCheckError(exec.Command("git", "-C", path, "tag", fmt.Sprintf(`v%s`, newVersion)), dryRun)
	runAndCheckError(exec.Command("git", "-C", path, "push", fmt.Sprintf(`v%s`, newVersion)), dryRun)
}

func runAndCheckError(command *exec.Cmd, dryRun bool) {
	fmt.Println(command.String())
	if dryRun {
		return
	}
}

func filterTags(tags []string) []string {
	var result []string
	for _, v := range tags {
		if mainModuleVersionTagRegExp.MatchString(v) {
			result = append(result, v)
		}
	}

	return result
}

func recentVersion(tags []string) (string, error) {
	if len(tags) == 0 {
		return "", errors.New("no versions was found")
	}
	collection := semver.Collection{}

	for _, tag := range tags {
		collection = append(collection, semver.MustParse(tag))
	}

	sort.Sort(&collection)

	return collection[len(collection)-1].String(), nil
}

func getTags(path string) []string {
	var tags []string

	cmd := exec.Command("git", "-C", path, "ls-remote", "--tags")
	fmt.Println(cmd.String())
	cmd.Stderr = cmd.Stdout
	r, err := cmd.StdoutPipe()
	CheckIfError(err)

	err = cmd.Start()
	CheckIfError(err)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, "refs/tags/")
		if len(parts) != 2 {
			log.Fatalln("wrong tag :", line)
		}

		tags = append(tags, strings.TrimSpace(parts[1]))
	}

	err = cmd.Wait()
	CheckIfError(err)

	return tags
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func findVersionInFile(path string) (string, error) {
	file := filepath.Join(path, "version.go")

	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer fd.Close()

	f, err := parser.ParseFile(token.NewFileSet(), "version.go", fd, 0)

	visitor := &VersionSearch{}
	ast.Walk(visitor, f)

	if visitor.CurrentVersion == "" {
		return "", errors.New("version was not found in " + file)
	}

	return "v" + visitor.CurrentVersion, nil
}

type VersionSearch struct {
	CurrentVersion string
}

func (v *VersionSearch) Visit(node ast.Node) ast.Visitor {

	switch elem := node.(type) {
	case *ast.GenDecl:
		if elem.Tok == token.CONST {
			for _, sp := range elem.Specs {
				if valSp, ok := sp.(*ast.ValueSpec); ok {
					for _, name := range valSp.Names {
						if name.Name == "Version" {
							v.CurrentVersion = strings.Trim(valSp.Values[0].(*ast.BasicLit).Value, `"`)
						}
					}
				}
			}
		}
	}

	return v
}
