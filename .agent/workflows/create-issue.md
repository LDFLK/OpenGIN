---
description: Create a new GitHub Issue
---

## Background

Refer to workflow `/understand-github-setting` and `/git-setup-opengin` to learn more about how the 
Github is configured to this project. 

## Context

Create a new Github issue using the mcp tools provide. 

## Setup 

Use Github MCP tools to create the issue. 

## Input

User will give the title and body of the issue and optionally the labels.

## How to Approach

Use the Github MCP tools to create the issue. 

## Actions

If any command needs to be executed please refer to the guideline in workflow /ide-configs.

Look into the user prompt where the user may say this is associated with an existing code block or a line of code. We must also make sure if this is the case, the issue must be added inline to that corresponding file as a comment, and put a tag in front FIXME (for bugs) or TODO (for enhancements) based on the created issue. 

Example 

```python
def calculate_total(price, discount):
    # FIXME: Issue #402 - If discount is > 100%, price becomes negative. Add validation.
    final_price = price - (price * discount)
    return final_price
```
