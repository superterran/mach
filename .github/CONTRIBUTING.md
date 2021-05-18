# Contributing

We love pull requests from everyone! 

Fork, then clone the repo:

```bash
    git clone git@github.com:your-username/mach.git
```
Set up your machine:

```bash   
    cd mach
    make build
```

Use cobra to create new feature, if needed:

```bash
cobra add <command>
```

Make sure the tests pass:

```bash
    go test ./...
```

While creating new commands, don't forget to create a `_test.go` file.

Make your changes, Add tests for your change, Make the tests pass:

```bash
    go test ./...
```

Make your change. Add tests for your change. Make the tests pass:

```bash
    rake
```

Push to your fork and [submit a pull request][pr].

[pr]: https://github.com/superterran/mach/compare/

At this point you're waiting on us. We like to at least comment on pull requests
within three business days (and, typically, one business day). We may suggest
some changes or improvements or alternatives.

Some things that will increase the chance that your pull request is accepted:

* Write tests.
* Write a [good commit message][commit].

[commit]: http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html