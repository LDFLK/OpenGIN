---
description: Create and iterate on an implementation plan for a GitHub issue
---

## Background
Refer to the `implementation_plan_artifact` definition in the system prompt.
This workflow is strictly for the **PLANNING** mode.

## Context
Before writing code, we must create a clear technical plan. This ensures we understand the requirements and have a verified strategy for resolution. The plan is a living document that evolves through user feedback.

## Input
- GitHub Issue Number or URL.
- (Optional) Specific user constraints or architectural preferences.

## How to Approach

1. **Understand the Issue**:
   - specific specific details from the GitHub issue.
   - If requirements are vague, ask clarifying questions first.
   - Use /analyze-issue workflow to learn about the issue. 

2. **Draft the Plan**:
   - Create `implementation_plan.md` in the artifacts directory.
   - Follow the `implementation_plan_artifact` template:
     - **Goal Description**: Link to the issue and summarize the objective.
     - **User Review Required**: Highlight breaking changes or major decisions.
     - **Proposed Changes**: Detail file-level changes (MODIFY, NEW, DELETE) grouped by component.
     - **Verification Plan**: automated tests and manual steps.

3. **Iterative Review**:
   - **CRITICAL**: Do not proceed to implementation until the user explicitly approves the plan.
   - Use `notify_user` to present the plan.
   - If the user requests changes, update `implementation_plan.md` and request review again.

## Actions
- Set `task_boundary` Mode to `PLANNING`.
- Use `write_to_file` to create/update `implementation_plan.md`.
- Use `notify_user` to block on user approval.