package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type config struct {
	Token    string `yaml:"token"`
	UserName string `yaml:"username"`
	Repos    []struct {
		Name  string `yaml:"name"`
		Owner string `yaml:"owner"`
	} `yaml:"repos"`
}

type pr struct {
	uri      string
	requests []string
}

type reviewRequest map[string][]string

type repoMap map[string][]pr

const (
	defaultConfigFileName = ".ghreviews/config.yml"
	displayFileName       = "requests.html"
)

var (
	cfgPath string
)

// print plain text
func printfPlain(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// print error text
func printfError(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func loadEnv() (*config, error) {

	// if the cfgPath is blank, try to load the default
	if cfgPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}

		cfgPath = path.Join(usr.HomeDir, defaultConfigFileName)
	}

	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setup() {
	printfPlain("running setup TBD")
	os.Exit(0)
}

func configure() {
	printfPlain("running configure TBD")
	os.Exit(0)
}

func auth(cfg *config) *githubv4.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.Token,
	})
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient)
}

func getPullRequests(ghClient *githubv4.Client, name string, owner string, repos repoMap) error {
	var q struct {
		Repository struct {
			Name         githubv4.String
			PullRequests struct {
				Nodes []struct {
					URL            githubv4.URI
					ReviewRequests struct {
						Nodes []struct {
							RequestedReviewer struct {
								User struct {
									Login githubv4.String
								} `graphql:"... on User"`
							}
						}
					} `graphql:"reviewRequests(last: 10)"`
				}
			} `graphql:"pullRequests(states: [OPEN], last: 50)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	v := map[string]interface{}{
		"name":  githubv4.String(name),
		"owner": githubv4.String(owner),
	}

	err := ghClient.Query(context.Background(), &q, v)
	if err != nil {
		return err
	}

	repos[name] = make([]pr, 0)

	for _, prNode := range q.Repository.PullRequests.Nodes {
		_pr := pr{
			uri:      prNode.URL.String(),
			requests: make([]string, 0),
		}

		for _, rrNode := range prNode.ReviewRequests.Nodes {
			_pr.requests = append(_pr.requests, string(rrNode.RequestedReviewer.User.Login))
		}

		repos[name] = append(repos[name], _pr)
	}

	return nil
}

func getUserRequests(cfg *config, repos repoMap) reviewRequest {
	rr := make(reviewRequest)
	for k, v := range repos {
		for _, pr := range v {
			for _, username := range pr.requests {
				if username == cfg.UserName {
					rr[k] = append(rr[k], pr.uri)
				}
			}
		}
	}

	return rr
}

func notify(rr reviewRequest) {
	if len(rr) == 0 {
		return
	}
	sb := &strings.Builder{}
	for k, v := range rr {
		sb.WriteString(k)
		sb.Write([]byte{':', '\n'})
		for _, uri := range v {
			sb.Write([]byte{'\t', '-', ' '})
			sb.WriteString(uri)
			sb.WriteByte('\n')
		}
	}

	printfPlain("Your review is requested for the following PRs")
	printfPlain(sb.String())
}

func main() {
	if len(os.Args) > 1 {
		switch {
		case os.Args[1] == "init":
			setup()
		case os.Args[1] == "config":
			configure()
		}
	}
	flag.StringVar(&cfgPath, "c", "", "path to config file")
	flag.Parse()

	// load config
	cfg, err := loadEnv()
	if err != nil {
		printfError("encountered error: %v", err)
		os.Exit(1)
	}

	// auth with github
	ghClient := auth(cfg)
	repos := make(repoMap)

	// get pull requests
	for _, repo := range cfg.Repos {
		if err := getPullRequests(ghClient, repo.Name, repo.Owner, repos); err != nil {
			printfError("error encountered: %v", err)
			os.Exit(1)
		}
	}

	requests := getUserRequests(cfg, repos)
	notify(requests)
}
