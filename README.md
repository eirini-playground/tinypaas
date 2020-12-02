This is a project which allows developers to deploy applications on Kubernetes using simple `git push`.

[![asciicast](https://asciinema.org/a/vKZlkl1xdJnj2ZWKrpoYrxMcF.svg)](https://asciinema.org/a/vKZlkl1xdJnj2ZWKrpoYrxMcF)

## How

There is no real magic in it. Here is the flow:

- User pushes the code to a git remote which runs inside the kubernetes cluster and is dedicated to his app.
- The git remote is configured to create a kpack "Image" as soon as code is pushed to it.
- When the Image is built, one of our components creates an [Eirini](https://github.com/cloudfoundry-incubator/eirini) LRP for it and Eirini deploys the application.
- Kpack is configured to push the image to a docker repository which we run as a pod.

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
