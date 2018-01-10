# Github Development Process - Cloning Model

This process assumes you are going to clone directly from https://github.com:cisco/arc.
This is opposed to forking model as described [here](github_forking_process.md).


## Cloning the repo

Clone the repo to your local machine. This assumes your are in the root of your go workspace.
```shell
git clone git@github.com:cisco/arc src/github.com/cisco/arc
```


## Make changes

Create the development branch
```shell
git checkout -b [development_branch]
```

... make changes ...

Commits go to your local development branch. For each commit to the branch use the
general form "Issue #N: ..." where N is the issue number of the problem being worked.
```shell
git commit -m "Issue #N: ..."
```

Before you create the pull request, sync your development branch with the master.
```shell
git checkout master
git pull --ff-only

git checkout [development_branch]
git rebase master
```

... re-build / test ...

Push your development branch to github. This will kickoff a travis-ci build for cisco/arc.
```shell
git push origin [development_branch]
```


## Pull request

Once your development branch is available go to https://github.com/cisco/arc/pulls and create a pull request
with the following attributes:

- Reviewers: arc-committers
- Assignees: _yourself_
- Projects:  arc

This will cause email to be sent out to the reviewers.

When ready to merge the PR use the **Rebase** option.

After the PR is merged, delete the branch associated with the PR. We will only use one development branch per PR.

If the issue associated with the PR is complete, mark it as closed.


## Cleanup the development branch

After the PR has merged, you no longer need your development branch. Switch to your master branch.
```shell
git checkout master
```

Since the development branch has been deleted upstream, this will remove it from your local workspace.
```shell
git fetch --all --prune
```

This will delete your development branch from your local workspace.
```shell
git branch -D [development_branch]
```

Pull your master branch using fast forward only. Since we are not making changes on master we should always be able to fast forward.
```shell
git pull --ff-only
```


## See also

- [Atlassian Tutorial: Merging vs. Rebasing](https://www.atlassian.com/git/tutorials/merging-vs-rebasing)
- [About merge methods on GitHub](https://help.github.com/articles/about-merge-methods-on-github/)
