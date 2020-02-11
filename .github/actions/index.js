const core = require("@actions/core");
const github = require("@actions/github");

async function run() {
  try {
    const [
      gitHubRepoOwner,
      gitHubRepoName
    ] = process.env.GITHUB_REPOSITORY.split("/");
    const gitHubSha = process.env.GITHUB_SHA;
    const gitHubToken = core.getInput("github-token");

    const octokit = new github.GitHub(gitHubToken);

    octokit.checks.create({
      owner: gitHubRepoOwner,
      repo: gitHubRepoName,
      name: "Check Created by API",
      head_sha: gitHubSha,
      status: "completed",
      conclusion: "success",
      output: {
        title: "Check Created by API",
        summary: `# All good ![step 1](https://www.imore.com/sites/imore.com/files/styles/w1600h900crop/public/field/image/2019/07/pokemonswordshieldstartertrio.jpg "Step 1")`
      }
    });

    core.setOutput("time", new Date().toTimeString());
  } catch (error) {
    core.setFailed(error.message);
  }
}

run();