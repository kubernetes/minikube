#### Adding a New Dependency
Minikube uses `Godep` to manage vendored dependencies.
`Godep` can be a bit finnicky with a project with this many dependencies.
Here is a rough set of steps that usually works to add a new dependency.

1. Make a clean GOPATH, with minikube in it.
  This isn't strictly necessary, but it usually helps.

  ```shell
  mkdir -p $HOME/newgopath/src/k8s.io
  export GOPATH=$HOME/newgopath
  cd $HOME/newgopath/src/k8s.io
  git clone https://github.com/kubernetes/minikube.git
  cd minikube
  ```
  
1. Install the package versions specified in Godeps/Godeps.json
  ```shell
  godep restore ./...
  ```
  NOTE:  If you encounter a HTTP 301 error, you may need to set the following:
  `git config --global http.https://gopkg.in.followRedirects true`
  
1. `go get` your new dependency.
  ```shell
  go get mynewdepenency
  ```

1. Use it in code, build and test.

1. Import the dependency from GOPATH into vendor/
  ```shell
  godep save ./...
  ```

  NOTE:  godep seems to have an issue with not recovering files in `vendor/golang.org/x/sys/windows/`
  IF you plan on committing your updated dependencies, make sure not to remove these files in your commit
  

  If it is a large dependency, please commit the vendor/ directory changes separately.
  This makes review easier in GitHub.
  
  NOTE: We have recently added a deprecation message regarding boot2docker.  Make sure that this deprecation message ends up in the vendored code at `/vendor/github.com/docker/machine/libmachine/provision/boot2docker.go`: [https://github.com/kubernetes/minikube/blob/master/vendor/github.com/docker/machine/libmachine/provision/boot2docker.go#L220](https://github.com/kubernetes/minikube/blob/master/vendor/github.com/docker/machine/libmachine/provision/boot2docker.go#L220)

  ```shell
  git add vendor/
  git commit -m "Adding dependency foo"
  git add --all
  git commit -m "Adding cool feature"
  ```
