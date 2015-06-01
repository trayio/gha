The GitHub Authenticator makes it easier to use the GitHub API from shell
scripts. It prints out your OAuth token to standard out. It saves your token to
`~/.github_<username>` (chmod 600) - if the file does not exist it will prompt
for your password and 2FA token and then print it. This allows you to always
use it as follows:

```sh
token=$(gha -user=robbiev)
curl -H "Authorization: token $token" ...
```
