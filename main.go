package main

// $ curl -sL \
//		-H "Accept: application/vnd.github+json" \
//		-H "Authorization: Bearer $GITHUB_TOKEN" \
//		-H "X-GitHub-Api-Version: 2022-11-28" \
//		https://api.github.com/user/repos \
// 		| jq --raw-output ".[].full_name"

// Note: it may not be worth using the go-git library for just one operation,
// i.e., the size of the binary will be (much?) larger b/c of all its dependencies.

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
)

type Project struct {
	Filename  string
	Platform  string
	OutputDir string
}

func clone(project Project) {
	var wg sync.WaitGroup

	// Build dirs can be nested, i.e., "build/foo".
	err := os.MkdirAll(project.OutputDir, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not create build dirs (build/PROJECT_NAME)")
		//		log.Fatal(err)
	}

	// Contents "should" never be large enough to need to chunk or buffer.
	readfile, err := os.Open(project.Filename)
	defer readfile.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] Could not open repositories file")
		log.Fatal(err)
	}

	filescanner := bufio.NewScanner(readfile)
	filescanner.Split(bufio.ScanLines)
	for filescanner.Scan() {
		wg.Add(1)
		fmt.Println(filescanner.Text())
		go func(fullRepoName string) {
			defer wg.Done()
			//			clone(project, fullRepoName)
			//		git.PlainClone(fmt.Sprintf("%s/%s", project.OutputDir, fullRepoName), false, &git.CloneOptions{
			_, err := git.PlainClone(fmt.Sprintf("%s/%s", project.OutputDir, fullRepoName), false, &git.CloneOptions{
				URL:      fmt.Sprintf("git@%s:%s", project.Platform, fullRepoName),
				Progress: nil,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %s %s\n", fullRepoName, err)
			}
		}(filescanner.Text())
	}
	wg.Wait()
}

func main() {
	file := flag.String("file", "repos.txt", "The name of the file that contains the repositories to clone")
	platform := flag.String("platform", "github.com", "The domain of the Git platform")
	outputDir := flag.String("outputDir", "./projects", "The name of the directory into which to clone the repositories")
	flag.Parse()

	fmt.Println("Cloning repositories...")
	startTime := time.Now()
	clone(Project{
		Filename:  *file,
		Platform:  *platform,
		OutputDir: *outputDir,
	})
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Printf("Total time taken: %f seconds\n", diff.Seconds())
}
