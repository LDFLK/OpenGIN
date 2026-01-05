---
description: Create a GitHub Pull Request
---

## Background

Refer to workflow `/understand-github-setting` and `/git-setup-opengin` to learn more about how the 
GitHub is configured to this project. 

## Context

Create a GitHub Pull Request using the mcp tools provided.

## Setup 

Use GitHub MCP tools to create the PR. 

## Input

User will give the title, body, and potentially the head and base branches. If not provided, try to infer the head branch from the current context and assume base is 'main' or the default branch.

## How to Approach

Use the GitHub MCP tools to create the pull request. 

## Actions

If any command needs to be executed please refer to the guideline in workflow /ide-configs.

## Guideline

### PR Title 

The format is as follows;

GH-<issue-number>: <issue-title>

### Description

Include a summary of changes done in this PR. For that please use the content in the `implementation_plan.md` file and what the user has described in the issue. 

If there are any commits in the branch already, use the commit messages to extract some information as well. 

At the end of the description please add the following

This PR closes #<issue-number>.