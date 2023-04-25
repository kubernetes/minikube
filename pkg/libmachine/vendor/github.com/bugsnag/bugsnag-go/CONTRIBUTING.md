Contributing
============

-   [Fork](https://help.github.com/articles/fork-a-repo) the [notifier on github](https://github.com/bugsnag/bugsnag-go)
-   Build and test your changes
-   Commit and push until you are happy with your contribution
-   [Make a pull request](https://help.github.com/articles/using-pull-requests)
-   Thanks!


Installing the go development environment
-------------------------------------

1.  Install homebrew

    ```
    ruby -e "$(curl -fsSL https://raw.github.com/Homebrew/homebrew/go/install)"
    ```

1. Install go

    ```
    brew install go --cross-compile-all
    ```

1. Configure `$GOPATH` in `~/.bashrc`

    ```
    export GOPATH="$HOME/go"
    export PATH=$PATH:$GOPATH/bin
    ```

Installing the appengine development environment
------------------------------------------------

1. Follow the [Google instructions](https://cloud.google.com/appengine/downloads).

Downloading the code
--------------------

You can download the code and its dependencies using

```
go get -t github.com/bugsnag/bugsnag-go
```

It will be put into "$GOPATH/src/github.com/bugsnag/bugsnag-go"

Then install depend


Running Tests
-------------

You can run the tests with

```shell
go test
```

If you've made significant changes, please also test the appengine integration with

```shell
goapp test
```

Releasing a New Version
-----------------------

If you are a project maintainer, you can build and release a new version of
`bugsnag-go` as follows:

1. Commit all your changes.
2. Update the version number in `bugsnag.go`.
3. Add an entry to `CHANGELOG.md` and update the README if necessary.
4. commit tag and push

    git commit -mv1.0.x && git tag v1.0.x && git push origin v1.0.x
