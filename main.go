package main

// $ curl -sL \
//		-H "Accept: application/vnd.github+json" \
//		-H "Authorization: Bearer $GITHUB_TOKEN" \
//		-H "X-GitHub-Api-Version: 2022-11-28" \
//		https://api.github.com/user/repos \
// 		| jq --raw-output ".[].full_name"

// Note: it may not be worth using the go-git library for just one operation.  The size of
// the binary will be a lot larger b/c of all its dependencies.

import (
	"bufio"
	"flag"
	"fmt"
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

func clone(project Project, fullRepoName string) {
	_, err := git.PlainClone(fmt.Sprintf("%s/%s", project.OutputDir, fullRepoName), false, &git.CloneOptions{
		URL:      fmt.Sprintf("git@%s:%s", project.Platform, fullRepoName),
		Progress: nil,
	})
	if err != nil {
		fmt.Println("err", err)
	}
}

func migrate(project Project) {
	var wg sync.WaitGroup

	// Create build dirs, i.e., "build/aion".
	err := os.MkdirAll(project.OutputDir, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not create build dirs (build/PROJECT_NAME)")
		//		log.Fatal(err)
	}

	// Contents will never be large enough to need to chunk or buffer.
	readfile, err := os.Open(project.Filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open repositories file")
		//		log.Fatal(err)
	}
	defer readfile.Close()

	filescanner := bufio.NewScanner(readfile)
	filescanner.Split(bufio.ScanLines)
	for filescanner.Scan() {
		wg.Add(1)
		go func(fullRepoName string) {
			defer wg.Done()
			clone(project, fullRepoName)
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
	migrate(Project{
		Filename:  *file,
		Platform:  *platform,
		OutputDir: *outputDir,
	})
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Printf("Total time taken: %f seconds\n", diff.Seconds())
}
