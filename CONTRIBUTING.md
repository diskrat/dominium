# Contributing Guidelines

## Branching Strategy
- `main` is protected. Direct commits are forbidden.
- All changes must be submitted via Pull Request (PR) targeting the `main` branch.
- Use GitFlow branch names (for a concise explanation, read the [Atlassian Gitflow Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow) guide):
  - `feature/<description>` for new features
  - `bugfix/<description>` for bug fixes
  - `docs/<description>` for documentation updates

## Workflow

1. Update local `main`:
   ```bash
   git checkout main
   git pull origin main
   ```

2. Create working branch:
   ```bash
   git checkout -b <type>/<description>
   ```

3. Commit logical changes:
   Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format. Keep commits focused and atomic.

   ```bash
   git add <files>
   git commit -m "<type>: <description>"
   ```
   *Valid types: feat, fix, docs, refactor, test*

4. Push working branch:
   If pushing this branch for the first time, set the upstream:
   ```bash
   git push -u origin <type>/<description>
   ```
   For subsequent pushes on the same branch:
   ```bash
   git push
   ```

5. Open Pull Request to `main`.
6. Await Code Review and approval before merging.
