# Github Development Process - Fork / Travis Model

This process assumes you are going to fork github.com:cisco/arc to your own account and
work from it there.

The advantage to this is that you can setup your own personal travis to build development
branches as you commit code, before you create a pull request in cisco/arc.

You can setup your personal travis here: https://travis-ci.org/profile/[username]


## Creating the fork

Just head over to the https://github.com/cisco/arc and click the "Fork" button. You want the fork to be located in [username]/arc.


## Cloning the repo

```shell
# Clone your fork to your local machine
git clone git@github.com:[username]/arc src/github.com/cisco/arc

u# Add cisco/arc as the upstream repo
git remote add upstream git@github.com:cisco/arc

# Verify that the "upstream" remote exists.
git remote -v

```


## Make changes

```shell
# Create the development branch
git checkout -b [branch_name]

... make changes ...

# Commits go to your local development branch. For each commit to the branch use the
# general form "Issue #N: ..." where N is the issue number of the problem being worked.
git commit -m "Issue #N, ..."

# If you have setup a personal travis build, this will kickoff the build.
git push origin [branch_name]
```


## Pull request

```shell
# Sync your master branch with the upstream master.
git fetch upstream
git checkout master
git merge --ff-only upstream/master

# Rebase your changes to pull in any changes on the upstream master.
git checkout [branch_name]
git rebase master

... re-build / test ...

# Push your development branch upstream. This will kickoff a travis build for cisco/arc.
git push upstream [branch_name]
```

Once your development branch is available goto https://github.com/cisco/arc/pulls and create a pull request.
Create the pull request with the following

- Reviewers: arc-committers
- Assignees: _yourself_
- Projects:  arc

This will cause email to be sent out to the reviewers.

When ready to merge a PR use the **Rebase** option.

After the PR is merged, delete the branch associated with the PR. We will only use one development branch per PR.

If the issue associated with the PR is complete, mark it as closed.


## Cleanup the development branch

```shell
# After the PR has merged, you no longer need your development branch.
git checkout master

# Since the development branch has been deleted upstream, this will remove it from your local copy.
git fetch --all --prune

# This will delete your development branch from your personal github account.
git push origin :[branch_name]

# This will delete your development from your local workspace.
git branch -D [branch_name]
```


## Syncing master to upstream

```shell
# Pull a copy of the upstream repo.
git fetch upstream

# Merge to code from upstream. You fast forward only, since you should not be making changes on this branch.
git checkout master
git merge --ff-only upstream/master

# Push the merged branch up to your github account to keep it in sync with the upstream repo.
git push
```
