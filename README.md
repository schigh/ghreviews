# ghreviews (WIP)

Little CLI tool to see if you have PR reviews out there on the Github

## Current usage:

There will be `init` and `configure` functionality in the future, but for now, do the following:

1. Create a folder in your `$HOME` directory called `.ghreviews`
2. Create a yaml config file in that folder called `config.yml`:

```yaml
token: 'yourgithubtokengoeshere' #  https://github.com/settings/tokens
username: 'yourusername'
repos:
  - name: 'my_repo'
    owner: 'myusername'
  - name: 'someone_elses_repo'
    owner: 'theirusername'

```

3. Create a Github personal access token.  For the repos, set the name and owner to match the parameters in the repo URL:

`https://github.com/<owner>/<name>`

4. Get the dependencies: `go mod download`

5. build the app: `go build -o /path/to/go/bin/ghreviews .`

6. Profit, probably

