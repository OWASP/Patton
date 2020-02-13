package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
)

type pattonParams struct {
	searchTerm, version string
	distro              string
	searchType          string
}

type pattonOutput struct {
	exitCode int
	stdout   []string
}
type execution struct {
	binaryPath string
	params     *pattonParams
	output     *pattonOutput
}

func (ex *execution) iHaveSearchTerm(searchTerm string) error {
	ex.params.searchTerm = searchTerm

	return nil
}

func (ex *execution) iHaveSearchTermAndVersion(searchTerm, version string) error {
	ex.params.searchTerm = searchTerm
	ex.params.version = version

	return nil
}

func (ex *execution) itIsAWordpressPlugin() error {
	return godog.ErrPending
}

func (ex *execution) iHaveOutputOfPackageManager(distro string, rawPkgOutput *gherkin.DocString) error {
	ex.params.distro = distro
	ex.params.searchTerm = rawPkgOutput.Content

	return nil
}

func (ex *execution) iExecutePattonSearchWithSearchType(searchType string) error {
	ex.params.searchType = searchType

	cmd := exec.Command(ex.binaryPath, ex.params.searchType)
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error starting command: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		ex.output.stdout = append(ex.output.stdout, scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ex.output.exitCode = exitErr.ExitCode()
		} else {
			return fmt.Errorf("Error starting command: %v", err)
		}
	}

	return nil
}

func (ex *execution) iGetAtLeastOneCve(table *gherkin.DataTable) error {
	count := 0
	for _, row := range table.Rows[1:] {
		for _, outLine := range ex.output.stdout {
			if strings.Contains(outLine, row.Cells[0].Value) {
				count++
				break
			}
		}
	}

	if count < (len(table.Rows) - 1) {
		return fmt.Errorf("Only %d matches", count)
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	exec := &execution{"patton", &pattonParams{}, &pattonOutput{stdout: make([]string, 0)}}

	if binaryPath, ok := os.LookupEnv("PATTON_BINARY"); ok {
		exec.binaryPath = binaryPath
	}

	s.Step(`^I have search term "([^"]*)" and version "([^"]*)"$`, exec.iHaveSearchTermAndVersion)
	s.Step(`^I have search term "([^"]*)"$`, exec.iHaveSearchTerm)
	s.Step(`^It is a Wordpress plugin$`, exec.itIsAWordpressPlugin)
	s.Step(`^I have the output of "([^"]*)" package manager$`, exec.iHaveOutputOfPackageManager)
	s.Step(`^I execute Patton search with search type "([^"]*)"$`, exec.iExecutePattonSearchWithSearchType)
	s.Step(`^I get at least one cve$`, exec.iGetAtLeastOneCve)

	s.BeforeScenario(func(interface{}) {
		exec.params = &pattonParams{}
		exec.output = &pattonOutput{stdout: make([]string, 0)}
	})
}