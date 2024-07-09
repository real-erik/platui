package process

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"github.com/pkg/browser"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Process struct {
	token  string
	client *github.Client
	ctx    context.Context
}

type Result struct {
	ID         int64
	Name       string
	Status     string
	Title      string
	Conclusion string
}

func NewProcess(token string) Process {
	return Process{
		token:  token,
		client: github.NewClient(nil).WithAuthToken(token),
		ctx:    context.Background(),
	}
}

func (p *Process) GetOrganizations() ([]Result, error) {

	githubOrgs, _, err := p.client.Organizations.ListOrgMemberships(p.ctx, &github.ListOrgMembershipsOptions{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var orgs []Result
	for _, org := range githubOrgs {
		orgs = append(orgs, Result{
			ID:   org.Organization.GetID(),
			Name: org.Organization.GetLogin(),
		})
	}

	return orgs, nil
}

func (p *Process) GetRepositories(organization string) ([]Result, error) {
	var githubRepositories []*github.Repository
	page := 1
	for {
		r, _, err := p.client.Repositories.ListByOrg(p.ctx, organization, &github.RepositoryListByOrgOptions{Sort: "full_name", ListOptions: github.ListOptions{Page: page, PerPage: 100}})
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		if len(r) == 0 {
			break
		}

		githubRepositories = append(githubRepositories, r...)

		page++
	}

	var repositories []Result
	for _, repository := range githubRepositories {
		repositories = append(repositories, Result{
			ID:   repository.GetID(),
			Name: repository.GetName(),
		})
	}

	return repositories, nil
}

func (p *Process) GetWorkflowRuns(organization string, repository string) ([]Result, error) {
	githubRuns, _, err := p.client.Actions.ListRepositoryWorkflowRuns(p.ctx, organization, repository, &github.ListWorkflowRunsOptions{})
	if err != nil {
		return nil, err
	}

	var runs []Result
	for _, run := range githubRuns.WorkflowRuns {
		runs = append(runs, Result{
			ID:         run.GetID(),
			Name:       run.GetName(),
			Title:      run.GetDisplayTitle(),
			Conclusion: run.GetConclusion(),
		})
	}

	return runs, nil
}

func (p *Process) GetArtifacts(organization string, repository string, workflowId int64) ([]Result, error) {
	githubArtifacts, _, err := p.client.Actions.ListWorkflowRunArtifacts(p.ctx, organization, repository, workflowId, &github.ListOptions{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var artifacts []Result
	for _, artifact := range githubArtifacts.Artifacts {
		artifacts = append(artifacts, Result{
			ID:   artifact.GetID(),
			Name: artifact.GetName(),
		})
	}

	return artifacts, nil
}

func (p *Process) DownloadArtifact(organization string, repository string, artifactId int64) error {
	url, _, err := p.client.Actions.DownloadArtifact(p.ctx, organization, repository, artifactId, 10)
	if err != nil {
		log.Fatal(err)
		return err
	}

	downloadZip(url.String())

	unzip(artifactId)

	return nil
}

// FIXME: return err instead
func downloadZip(url string) {
	err := os.MkdirAll("output", 0755)
	if err != nil {
		panic(err)
	}

	output, err := os.Create("output/file.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// log.Println("ZIP file downloaded successfully.")
}

// FIXME: return err instead
func unzip(artifactId int64) {
	dst := fmt.Sprintf("output/%d", artifactId)
	archive, err := zip.OpenReader("output/file.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		// fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}

		if f.FileInfo().IsDir() {
			// fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func (p *Process) Run(filepath string) error {

	if strings.Contains(filepath, "webm") {
		browser.OpenFile(filepath)
	}
	if strings.Contains(filepath, "png") {
		browser.OpenFile(filepath)
	}

	if strings.Contains(filepath, "zip") {
		cmd := exec.Command("playwright", "show-trace", filepath)

		err := cmd.Start()
		if err != nil {
			// TODO: try pnpm exec playwright
			log.Fatal(err)
			return err
		}

		err = cmd.Wait()
		if err != nil {
			log.Fatal(err)
			return err
		}

	}

	return nil
}
