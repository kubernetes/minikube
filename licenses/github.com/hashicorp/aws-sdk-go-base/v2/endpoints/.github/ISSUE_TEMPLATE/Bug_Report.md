---
name: 🐛 Bug Report
about: If something isn't working as expected 🤔.

---

<!--- Please keep this note for the community --->

### Community Note

* Please vote on this issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment

<!--- Thank you for keeping this note for the community --->

### Environment and Versions

* Terraform Executor: Local workstation / EC2 Instance / ECS / Terraform Cloud or Enterprise / Other (please specify)
  * If outside Terraform Cloud/Enterprise, which operating system and version: Linux/macOS/Windows
* Terraform CLI version: `terraform -v`
* Terraform AWS Provider version: `terraform providers -v`
* Terraform Backend/Provider Configuration:

```hcl
# Copy-paste your Terraform S3 Backend or AWS Provider configurations here
```

* AWS environment variables (if any):

```sh
AWS_XXX=example
```

* AWS configuration files (if any):

```txt
# Copy-paste your AWS shared configuration file contents here
```

### Expected Behavior

What should have happened?

### Actual Behavior

What actually happened?

### Steps to Reproduce

<!--- Please list the steps required to reproduce the issue. --->

1. `terraform init`
1. `terraform apply`

### Debug Output

Please provide a link to a GitHub Gist containing the complete debug output. Please do NOT paste the debug output in the issue; just paste a link to the Gist.

To obtain the debug output, see the [Terraform documentation on debugging](https://www.terraform.io/docs/internals/debugging.html).

### References

Are there any other GitHub issues (open or closed) or pull requests that should be linked here? Terraform or AWS documentation?
