These are some notes on how the dockerfile in this directory could be used
to create a git remote inside the cluster. This way you don't need to push code
anywhere outside the cluster (e.g. GitHub).

This is not complete yet (and probably not working).

TODO: Complete this feature.

## Deploy the git remote

```bash
docker run -d git_remote
```

## Setup your git repo

In your .git/config file put this remote (use your container's IP address):

```
[remote "local"]
	url = root@172.17.0.2:myproject.git
	fetch = +refs/heads/*:refs/remotes/local/*

```
