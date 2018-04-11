// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates listers.go It can be invoked by running
// go generate
package main

import (
	"log"
	"os"
	"text/template"
	"time"
)

func main() {

	f, err := os.Create("listers.go")
	die(err)
	defer f.Close()

	clientListTemplate.Execute(f, struct {
		Timestamp     time.Time
		PageableLists []repositoryClientListItem
		OtherLists    []repositoryClientListItem
		Collectors    []collector
	}{
		Timestamp: time.Now(),
		PageableLists: []repositoryClientListItem{
			{"ListBranches", "ListOptions", "[]*github.Branch"},
			{"ListForks", "RepositoryListForksOptions", "[]*github.Repository"},
			{"ListComments", "ListOptions", "[]*github.RepositoryComment"},
			{"ListCommits", "CommitsListOptions", "[]*github.RepositoryCommit"},
			{"ListContributors", "ListContributorsOptions", "[]*github.Contributor"},
			{"ListDeployments", "DeploymentsListOptions", "[]*github.Deployment"},
			{"ListHooks", "ListOptions", "[]*github.Hook"},
			{"ListInvitations", "ListOptions", "[]*github.RepositoryInvitation"},
			{"ListKeys", "ListOptions", "[]*github.Key"},
			{"ListPagesBuilds", "ListOptions", "[]*github.PagesBuild"},
			{"ListProjects", "ProjectListOptions", "[]*github.Project"},
			{"ListReleases", "ListOptions", "[]*github.RepositoryRelease"},
			{"ListTags", "ListOptions", "[]*github.RepositoryTag"},
			{"ListTeams", "ListOptions", "[]*github.Team"},
			{"ListCollaborators", "ListCollaboratorsOptions", "[]*github.User"},
		},
		OtherLists: []repositoryClientListItem{
			{"ListParticipation", "", "*github.RepositoryParticipation"},
			{"ListLanguages", "", "map[string]int"},

			// Unused as of now
			{"ListContributorsStats", "", "[]*github.ContributorStats"},
			{"ListPunchCard", "", "[]*github.PunchCard"},
			{"ListTrafficPaths", "", "[]*github.TrafficPath"},
			{"ListTrafficReferrers", "", "[]*github.TrafficReferrer"},
			{"ListAllTopics", "", "[]string"},
			{"ListCodeFrequency", "", "[]*github.WeeklyStats"},
			{"ListCommitActivity", "", "[]*github.WeeklyCommitActivity"},
		},
		Collectors: []collector{
			{Lister: "ListReleases", ConfigName: "Downloads", IsPageable: true, IsListResult: false},
			{Lister: "ListParticipation", ConfigName: "Participation", IsPageable: false, IsListResult: false},
			{Lister: "ListForks", ConfigName: "Forks", IsPageable: true, IsListResult: true},
			{Lister: "ListLanguages", ConfigName: "Languages", IsPageable: false, IsListResult: true},
			{Lister: "ListContributors", ConfigName: "Contributors", IsPageable: true, IsListResult: true},
			{Lister: "ListBranches", ConfigName: "Branches", IsPageable: true, IsListResult: true},
		},
	})
}

//func (s *RepositoriesService) ListByOrg(ctx context.Context, org string, opt *RepositoryListByOrgOptions) ([]*Repository, *Response, error)

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type repositoryClientListItem struct {
	Function   string
	OptionType string
	ResultType string
}

type collector struct {
	Lister       string
	ConfigName   string
	IsPageable   bool
	IsListResult bool
}

var clientListTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}

package beater

import (
	"github.com/google/go-github/github"

	"github.com/elastic/beats/libbeat/common"
)


{{- range .PageableLists}}

func (rc *repositoryClient) {{.Function}}(max int) ({{.ResultType}}, error) {
	opt := &github.{{.OptionType}}{}

	var results {{.ResultType}}

	for {
		list, resp, err := rc.client.Repositories.{{.Function}}(rc.ctx, rc.GetOwner(), rc.GetName(), opt)
		if err != nil {
			return results, err
		}

		results = append(results, list...)
		if resp.NextPage == 0 || (len(results) >= max && max > 0) {
			break
		}

		opt.Page = resp.NextPage
	}

	return results, nil
}

{{- end }}

{{- range .OtherLists}}

func (rc *repositoryClient) {{.Function}}() ({{.ResultType}}, error) {
	result, _, err := rc.client.Repositories.{{.Function}}(rc.ctx, rc.GetOwner(), rc.GetName())
	return result, err
}

{{- end }}


{{- range .Collectors}}

func (bt *Githubbeat) collect{{.ConfigName}}(rc *repositoryClient) common.MapStr {
	if !bt.config.{{.ConfigName}}.Enabled {
		return nil
	}

	rawResults, err := rc.{{.Lister}}({{if .IsPageable}}bt.config.{{.ConfigName}}.Max{{end}})

	formatted := bt.extract{{.ConfigName}}(rawResults, err)
{{if .IsListResult}}
	return createListMapStr(formatted, err, bt.config.{{.ConfigName}}.List)
{{else}}
	return appendError(formatted, err)
{{end}}
}

{{- end }}
`))
