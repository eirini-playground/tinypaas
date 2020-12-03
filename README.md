This is a project which allows developers to deploy applications on Kubernetes using simple `git push`.

[![asciicast](https://asciinema.org/a/vKZlkl1xdJnj2ZWKrpoYrxMcF.svg)](https://asciinema.org/a/vKZlkl1xdJnj2ZWKrpoYrxMcF)

**Disclaimer**: This is not supposed to be used in production. It's just a PoC created as part of a hackday. Use it as inspiration only!

## How

There is no real magic in it. Here is the flow:

- User creates a new application with the `tinypaas create` command
- [kpack](https://github.com/pivotal/kpack) creates a docker image for this application
- Eirini creates a statefulset for with the created image.
- Our Kubernetes controller updates an Ingress resource to give a nice url to the application.
- `tinypaas list` command gives you the created url for your application
