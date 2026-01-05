---
description: Github Setup for OpenGIN Repository
---

## Basic Configuration

The Github Upstream for this repo is https://github.com/LDFLK/OpenGIN and this has been forked for my work https://github.com/vibhatha/OpenGIN. 

I have set the upstream as ldf and my origin as vibhatha. Please refer to the following output from `git remote -v`. 

```bash
ldf     git@github.com:LDFLK/OpenGIN.git (fetch)
ldf     git@github.com:LDFLK/OpenGIN.git (push)
vibhatha        git@github.com:vibhatha/OpenGIN.git (fetch)
vibhatha        git@github.com:vibhatha/OpenGIN.git (push)

```

I have also given access to the Github MCP to read, write Github issues. 


## Best Practices

1. Maintain branches for each feature/bug

We always create a branch for feature development. There will be no commits
for the main branch, even in your fork. The Github issue number is referred 
to as a suffix and we always make branches `gh-<issue-number>` format. 

2. Make sure to rebase and keep things in sync. 

We follow these to keep ourselves sync. 

Once you are in the branch, 

Make sure everything you need to be committed is there and pushed to your branch. 

Do a rebase with the upstream (I have configured it as ldf).

```bash
git pull --rebase ldf main

```

Then resolve any conflicts if there is any. As the Antigravity agent, you can 
notify the user if there are conflicts.