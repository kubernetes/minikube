# Your inspiring proposal title

* First proposed: 2021-05-17
* Authors: Andriy Dzikh (@andriyDev)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

## Summary

As of June 2021, public users have no way to view the flake rates of integration tests. This can make it tricky to determine whether an individual PR is causing a new error, or if the test failure is just a flake, or if the test is entirely broken. While each test failure should be investigated, sometimes an unrelated test fails, and knowing that the test has been "flake-y" can increase confidence in a particular PR.

This proposal is for a system to inform users, both public and internal, of the flake rates of various tests on the master branch.

## Goals

*   Comments on PRs describing the flake rates of failing tests
*   Charts to visualize the flake rates of any test

## Design Details

### Overview

The full overview of the system is as follows:
*   Jenkins integration test jobs running on master use gopogh summaries (already created) and processes them into a series of test run entries, including whether the test passed and its duration
*   These test run entries are appended into a full dataset of all test runs at `gs://minikube-flake-rate/data.csv`
*   A Jenkins job runs regularly to compute the flake rates of tests in `gs://minikube-flake-rate/data.csv` and outputs the results into `gs://minikube-flake-rate/flake_rates.csv`, including the environment (e.g. `Docker_Linux`), the test name, the flake rate as a percentage, and the average duration
*   An HTML+JS file, hosted on `gs://minikube-flake-rate/flake_chart.html`, will read the full test data (`gs://minikube-flake-rate/data.csv`), and parse it into a chart displaying the daily flake rates and average durations of the requested tests (specified by url query arguments)
*   When the Jenkins integration test jobs run on a PR, they will instead write a comment on the PR including all failed tests and their flake rates. These comments will also include a link to the flake charts for each test/environment

### Test Data Collection

Our system needs a way to collect data from our existing integration tests. As of June 2021, all integration Jenkins jobs run the integration tests, then use gopogh to create HTML files for viewing, and JSON files for summarizing the test results. The new system will then take these JSON summaries, and pass them into a script named `upload_tests.sh`. This script will process the summary into a CSV file of its test runs and related data, and upload this to a dataset of all test runs at `gs://minikube-flake-rate/data.csv`. This file will be publicly accessible to all users to read (and later chart the data).

### Flake Rate Computation

On a regular schedule (every 4 hours for example), a Jenkins job named `Flake Rate Computation` will download `gs://minikube-flake-rate/data.csv` and compute a failure percentage for each test/environment combination, based on the number of failures occurring in the past 15 days (this will be configurable). Note that this will be the past 15 dates that the test actually ran, since this can allow a test to be skipped for a long period of time and then unskipped while maintaining the old flake rate. This will also compute the average duration of the test for the past 15 days. The resulting data will then be stored in `gs://minikube-flake-rate/flake_rates.csv`.

### Charts

To allow users to see the daily "flakiness" of a test/environment combination, we will have an HTML file at `gs://minikube-flake-rate/flake_chart.html` and a JS file at `gs://minikube-flake-rate/flake_chart.js`. These will fetch `gs://minikube-flake-rate/data.csv` and parse it into Google Charts allowing us to visualize the "flakiness" over time. This can help track down exactly when a test became "flake-y" by telling us the commits associated with each test date. The flake rate charts will use two query parameters (e.g. `google.com?these=are&three=query&param=eters`): test which will control which test to view (`TestFunctional/parallel/LogsCmd`), and env which will control the environment to view (e.g. `Docker_Linux`). By hosting this in a GCS bucket, we can avoid needing to create actual servers to manage this. Since these files are incredibly lightweight, there is little concern over the workload of hosting these files.

### PR Comments

As PRs can have many failures, it is useful to be told the flake rates of some of these tests. Some of our tests could be more stable, and knowing that a failed test is known to be unreliable can be informative for both the PR creators and the PR reviewers. To that end, whenever an integration test running on a PR completes, it will call a script named `report_flakes.sh`. This script will use a provided gopogh summary (for the test run that should be reported about) and the public `gs://minikube-flake-rate/flake_rates.csv` to comment on the PR about all failed tests, their flake rates, and links to the flake charts for the test and the environment the failure occurred on.

An important note, as of June 2021, there is no system for synchronizing after all integration tests run. To workaround this, each integration test (from a set of "important" environments) will trigger `report_flakes.sh` on its own. This means that each environment's test failures will be reported in a separate list. The set of "important" environments should be kept as small as possible in order to prevent spam and keep PRs readable. This is merely a temporary solution - a more permanent design will need to be considered in the future.

### Additional Information

The raw data `gs://minikube-flake-rate/data.csv` can become quite large if stored as simple CSV data. Since this is a CSV file, it will contain columns for each field which includes commit hash, test date, test name, etc. Some of these fields can be repetitive like commit hash and test date. Since test runs are generally added such that all the tests for a single commit hash are added consecutively, we can use a sentinel value to repeat values. Specifically, if the previous row had the same value for the current column, we can replace the current column value with an empty space. When parsing the reverse can be performed - whenever a blank space is found, simply repeat the value of the previous row.

```
Input:
hash,2021-06-10,Docker_Linux,TestFunctional,Passed,0.5
hash,2021-06-10,Docker_Linux_containerd,TestFunctional,Failed,0.6

Output:
hash,2021-06-10,Docker_Linux,TestFunctional,Passed,0.5
,,DockerLinux_containerd,,Failed,0.6
```

This optimization will be done in `optimize_data.sh`.


## Alternatives Considered

Another optimization technique that can be used on `gs://minikube-flake-rate/data.csv` is to use a string table. The string table would be stored at `gs://minikube-flake-rate/data_strings.txt` and would contain an ordered list of unique strings. The index of each string can then be used in place of the actual text in `gs://minikube-flake-rate/data.csv`. The index into the string table will very likely be shorter than the text it represents, saving space. For non-consecutive strings, this can be a very big saving. For example, test names are repeated very often in `gs://minikube-flake-rate/data.csv`, but almost never consecutively. With this technique, the dataset can be compressed even further.

The trouble with this technique is complexity - any users of the dataset would need to also manage the string table. More importantly, if a new string needs to be added to the string table, the order is critical, meaning synchronization can be a problem (since our integration tests run in parallel). Due to these concerns, this option was rejected (although this may be a more feasible option in the future depending on how integration test synchronization is handled).
