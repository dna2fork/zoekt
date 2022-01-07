# scmsync

scm sync helper

### preparation

**Gitlab**

create a private token for specified user, use https to clone project; for example: `https://<username>:<private_token>@git.server/path/to/project.git
or directly use git@git.server:path/to/project

**Perforce** use p4 login to generate perforce ticket

`config`:

- `interval`: how frequently sync source code (in seconds)
- `git <url>  <localDiskPath> <projectName> <branchName> <username> [h]`
- `git <url>  <localDiskPath> <projectName> <username> [h]`
- `p4  <host> <clientName>     <projectName> <username> [h]`

where optional `h` means dump full commits into `.git|.p4/.zoekt/commits`

### startService

```
cp config.exmpale config
# edit config
python scmsync.py
```
