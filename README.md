This is a project which allows developers to deploy applications on Kubernetes using [paketo buildpacks](https://paketo.io/) with a simple `git push`.

## How

There is no real magic in it. Here is the flow:

- User pushes the code to a git remote which we run in a pod>
- kpack is configured to poll our git remote for updates. When the update happens, a new image is built.
- kpack is configured to push the image to a docker repository which we run as a pod.

TODO:
Define what happens next. Obviously there is going to be a deployment for the application which will use this updated image.

But who is going to create that deployment? How can the user create more applications? (thus creating more remotes etc).

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
