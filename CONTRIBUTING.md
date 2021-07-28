# Contributing to FlashPipe

Contributions from the community are welcome. This project uses [Developer Certificate of Origin (DCO)](https://developercertificate.org) to certify that contributors have the right to submit the code they are contributing.

When submitting Pull Requests, ensure all commits contain a `Signed-off-by` line in its commit message to pass the automated check by [Probot: DCO](https://probot.github.io/apps/dco/).

Following are the guidelines for contributions:
- If you are a first time contributor on GitHub, check out the [First Contributions repository](https://github.com/firstcontributions/first-contributions).
- Wherever possible ensure that changes in commits are related.
  - If there are various changes, avoid a single commit for all of them.
  - Splitting the changes into different commits allows for providing specific details in each commit message and eases the review process.
- Work on changes in a different branch (in your forked repository) other than `main` and submit PRs from that branch. In general, I use `rebase and merge` for PRs into a different branch before the changes make it into the `main` branch and a Docker image release. This ensures the `main` branch's history is clean and your fork can continue to track it easily for further changes.
- If there are various unrelated changes, it is better to submit them as separate PRs. It is easier to review and include small individual chunks of changes into the `main` branch.
- If you have something big, please open an issue first so that we can have a discussion about it. Don't get me wrong - I truly welcome contributions and are thrilled to have them. Having a discussion beforehand ensures we are on the same page before starting a big endeavour, and hopefuly avoids any surprises during the PR review process.
