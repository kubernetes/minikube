<!-- markdownlint-disable single-title -->
# v2.0.0 (Unreleased)

# v2.0.0-beta.65 (2025-06-06)

ENHANCEMENTS

* Adds support for additional AWS region ([#1321](https://github.com/hashicorp/aws-sdk-go-base/pull/1321))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.64 (2025-04-04)

ENHANCEMENTS

* Adds support for additional AWS regions ([#1301](https://github.com/hashicorp/aws-sdk-go-base/pull/1301))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.63 (2025-03-18)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.62 (2025-02-04)

ENHANCEMENTS

* Adds support for additional AWS regions ([#1262](https://github.com/hashicorp/aws-sdk-go-base/pull/1262))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.61 (2025-01-15)

ENHANCEMENTS

* Adds support for AWS region `mx-central-1` ([#1248](https://github.com/hashicorp/aws-sdk-go-base/pull/1248))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.60 (2025-01-08)

ENHANCEMENTS

* Adds support for AWS region `ap-southeast-7` ([#1240](https://github.com/hashicorp/aws-sdk-go-base/pull/1240))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.59 (2024-10-28)

ENHANCEMENTS

* Update generated `endpoints` ([#1205](https://github.com/hashicorp/aws-sdk-go-base/pull/1205))

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.58 (2024-09-23)

ENHANCEMENTS

* Adds per-partition service endpoint metadata to the top-level `endpoints` package ([#1182](https://github.com/hashicorp/aws-sdk-go-base/pull/1182))

# v2.0.0-beta.57 (2024-09-18)

ENHANCEMENTS

* Adds top-level `endpoints` package containing AWS partition and Region metadata ([#1176](https://github.com/hashicorp/aws-sdk-go-base/pull/1176))

# v2.0.0-beta.56 (2024-09-11)

ENHANCEMENTS

* Adds support for IAM role chaining ([#1170](https://github.com/hashicorp/aws-sdk-go-base/pull/1170))

# v2.0.0-beta.55 (2024-08-27)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Adds support for AWS region `ap-southeast-5` ([#1156](https://github.com/hashicorp/aws-sdk-go-base/pull/1156))

# v2.0.0-beta.54 (2024-06-19)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.53 (2024-05-09)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Adds `Backoff` parameter to configure the backoff strategy the retryer will use to determine the delay between retry attempts ([#1045](https://github.com/hashicorp/aws-sdk-go-base/pull/1045))

# v2.0.0-beta.52 (2024-04-11)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Adds `MaxBackoff` parameter to configure the maximum backoff delay that is allowed to occur between retrying a failed request ([#1011](https://github.com/hashicorp/aws-sdk-go-base/pull/1011))

# v2.0.0-beta.51 (2024-04-04)

BUG FIXES

* Correctly handles user agents passed using `TF_APPEND_USER_AGENT` which contain `/`,  `(`, `)`, or space ([#990](https://github.com/hashicorp/aws-sdk-go-base/pull/990))

# v2.0.0-beta.50 (2024-03-19)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Changes the default AWS SDK for Go v2 API client [`RateLimiter`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2@v1.26.0/aws/retry#RateLimiter) to [`ratelimit.None`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/ratelimit#pkg-variables) in order to maintain behavioral compatibility with AWS SDK for Go v1 ([#977](https://github.com/hashicorp/aws-sdk-go-base/pull/977))

# v2.0.0-beta.49 (2024-03-12)

NOTES

* Updates to Go 1.21 (used by Terraform starting with `v1.6.0` and the AWS provider since `5.37.0`), which, for Windows, requires at least Windows 10 or Windows Server 2016--support for previous versions has been discontinued--and, for macOS, requires macOS 10.15 Catalina or later--support for previous versions has been discontinued ([#960](https://github.com/hashicorp/aws-sdk-go-base/pull/960)).
* Updates dependencies.

BREAKING CHANGES

* Removes the `UseLegacyWorkflow` configuration option ([#962](https://github.com/hashicorp/aws-sdk-go-base/pull/962))

# v2.0.0-beta.48 (2024-02-21)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Adds `TokenBucketRateLimiterCapacity` parameter to configure token bucket rate limiter capacity ([#933](https://github.com/hashicorp/aws-sdk-go-base/pull/933))

# v2.0.0-beta.47 (2024-02-14)

BUG FIXES

* Ensures that each AWS SDK for Go v2 API client gets an independent `Retryer` ([#918](https://github.com/hashicorp/aws-sdk-go-base/pull/918))
* Updates dependencies.

# v2.0.0-beta.46 (2024-01-03)

BUG FIXES

* Updates dependencies.

ENHANCEMENTS

* Adds support for AWS region `ca-west-1` ([#858](https://github.com/hashicorp/aws-sdk-go-base/pull/858))

# v2.0.0-beta.45 (2023-12-12)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.44 (2023-12-01)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.43 (2023-11-28)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.42 (2023-11-21)

BUG FIXES

* Updates dependencies.

# v2.0.0-beta.41 (2023-11-16)

BUG FIXES

* Fixes error with stripping SSO Start URLs in shared configuration files ([#787](https://github.com/hashicorp/aws-sdk-go-base/pull/787))

# v2.0.0-beta.40 (2023-11-14)

ENHANCEMENTS

* Adds `HttpsProxy` and `NoProxy` parameters to configure separate HTTPS proxy and proxy bypass ([#775](https://github.com/hashicorp/aws-sdk-go-base/pull/775))

# v2.0.0-beta.39 (2023-11-08)

ENHANCEMENTS

* Adds `SsoEndpoint` parameter to override default AWS SSO endpoint ([#741](https://github.com/hashicorp/aws-sdk-go-base/pull/741))

# v2.0.0-beta.38 (2023-11-01)

ENHANCEMENTS

* Improves the readability of an error message returned from `GetAwsAccountIDAndPartition` ([#713](https://github.com/hashicorp/aws-sdk-go-base/pull/713))
* Adds `tfawserr.ErrCodeContains` ([#733](https://github.com/hashicorp/aws-sdk-go-base/pull/733))

# v2.0.0-beta.37 (2023-10-11)

NOTES

* Updates dependencies, including an update to `aws-sdk-go-v2` which fixes an [issue](https://github.com/aws/aws-sdk-go-v2/issues/2166) with FIPS endpoint resolution in GovCloud regions.

# v2.0.0-beta.36 (2023-09-22)

BREAKING CHANGES

* The `ValidateRegion` function has been moved to the `validation` package and renamed to `SupportedRegion` ([#650](https://github.com/hashicorp/aws-sdk-go-base/pull/650))

ENHANCEMENTS

* Adds `JSONNoDuplicateKeys` function to the `validation` package ([#650](https://github.com/hashicorp/aws-sdk-go-base/pull/650))
* logging: S3 object bodies are no longer logged. Body size is logged instead.

# v2.0.0-beta.35 (2023-09-05)

ENHANCEMENTS

* Adds `AllowedAccountIds` and `ForbiddenAccountIds` fields and a `VerifyAccountIDAllowed` method to `Config` ([#638](https://github.com/hashicorp/aws-sdk-go-base/pull/638))
* Adds `tfawserr.ErrHTTPStatusCodeEquals` ([#642](https://github.com/hashicorp/aws-sdk-go-base/pull/642))

# v2.0.0-beta.34 (2023-08-17)

BREAKING CHANGES

* Now requires a `logging.Logger` implementation to enable logging ([#605](https://github.com/hashicorp/aws-sdk-go-base/pull/605))

# v2.0.0-beta.33 (2023-08-04)

BREAKING CHANGES

* Public functions return `diag.Diagnostics` instead of `error` ([#553](https://github.com/hashicorp/aws-sdk-go-base/pull/553))

# v2.0.0-beta.32 (2023-07-20)

NOTES

Dependency Updates

# v2.0.0-beta.31 (2023-07-06)

ENHANCEMENT

* Adds `tfawserr.ErrMessageContains` for AWS services that don't define Go error types ([#533](https://github.com/hashicorp/aws-sdk-go-base/pull/533))

# v2.0.0-beta.30 (2023-06-22)

ENHANCEMENTS

* Adds more sensitive value masking in HTTP request and response logs ([#523](https://github.com/hashicorp/aws-sdk-go-base/pull/523))
* Adds `tfawserr.ErrCodeEquals` for AWS services that don't define Go error types ([#524](https://github.com/hashicorp/aws-sdk-go-base/pull/524))

# v2.0.0-beta.29 (2023-06-08)

ENHANCEMENT

* Enables support for Adaptive retry mode ([#489](https://github.com/hashicorp/aws-sdk-go-base/pull/489))

# v2.0.0-beta.28 (2023-06-01)

BUG FIXES

* Limits HTTP response body size in logs to 4 KB ([#490](https://github.com/hashicorp/aws-sdk-go-base/pull/490))

ENHANCEMENTS

* Updates limit of HTTP requsest body size in logs to 1 KB ([#490](https://github.com/hashicorp/aws-sdk-go-base/pull/490))

# v2.0.0-beta.27 (2023-05-24)

BUG FIXES

* Reintroduces special handling to work around very high AWS API retry counts removed in v2.0.0-beta.26 ([#481](https://github.com/hashicorp/aws-sdk-go-base/pull/481))

# v2.0.0-beta.26 (2023-05-23)

BREAKING CHANGES

* Removes special handling to work around very high AWS API retry counts. ([#462](https://github.com/hashicorp/aws-sdk-go-base/pull/462))

# v2.0.0-beta.25 (2023-03-23)

ENHANCEMENTS

* Enables more logging during setup. ([#386](https://github.com/hashicorp/aws-sdk-go-base/pull/386))

# v2.0.0-beta.24 (2023-02-23)

BUG FIXES

* Avoids retries on `Expired Token` errors. ([#362](https://github.com/hashicorp/aws-sdk-go-base/pull/362))

# v2.0.0-beta.23 (2023-02-09)

BUG FIXES

* Truncates HTTP request bodies in logs. ([#351](https://github.com/hashicorp/aws-sdk-go-base/pull/351))

ENHANCEMENTS

* Adds support for AWS region `ap-southeast-4`. ([#348](https://github.com/hashicorp/aws-sdk-go-base/pull/348))

# v2.0.0-beta.22 (2023-02-02)

BREAKING CHANGES

* Adds `context.Context` return value to `GetAwsConfig` with configured logger. Adds `context.Context` parameter to `awsbasev1.GetSession`. ([#341](https://github.com/hashicorp/aws-sdk-go-base/pull/341))

BUG FIXES

* Scrubs sensitive values from HTTP request and response logs. ([#341](https://github.com/hashicorp/aws-sdk-go-base/pull/341))

ENHANCEMENTS

* Uses structured logging. ([#341](https://github.com/hashicorp/aws-sdk-go-base/pull/341))

# v2.0.0-beta.21 (2023-01-13)

ENHANCEMENTS

* Adds support for a congfigurable HTTP client. ([#340](https://github.com/hashicorp/aws-sdk-go-base/pull/340))

# v2.0.0-beta.20 (2022-11-22)

ENHANCEMENTS

* Adds support for AWS region `ap-south-2`. ([#339](https://github.com/hashicorp/aws-sdk-go-base/pull/339))

# v2.0.0-beta.19 (2022-11-16)

ENHANCEMENTS

* Adds support for AWS region `eu-south-2`. ([#337](https://github.com/hashicorp/aws-sdk-go-base/pull/337))

# v2.0.0-beta.18 (2022-11-15)

ENHANCEMENTS

* Adds support for AWS region `eu-central-2`. ([#335](https://github.com/hashicorp/aws-sdk-go-base/pull/335))

# v2.0.0-beta.17 (2022-08-31)

ENHANCEMENTS

* Adds support for `max_attempts` in shared config files. ([#278](https://github.com/hashicorp/aws-sdk-go-base/pull/278))
* Prevents silent failures when `RoleARN` missing from `AssumeRole` or `AssumeRoleWithWebIdentity`. ([#277](https://github.com/hashicorp/aws-sdk-go-base/pull/277))
* Adds support for `SourceIdentity` with `AssumeRole`. ([#311](https://github.com/hashicorp/aws-sdk-go-base/pull/311))
* Adds support for AWS region `me-central-1`. ([#328](https://github.com/hashicorp/aws-sdk-go-base/pull/328))
* Adds support for passing HTTP User-Agent products in `useragent.Context`. ([#318](https://github.com/hashicorp/aws-sdk-go-base/pull/318))

# v2.0.0-beta.16 (2022-04-27)

BREAKING CHANGES

* Removes boolean `SkipEC2MetadataApiCheck` and adds `EC2MetadataServiceEnableState` of type `imds.ClientEnableState`. ([#240](https://github.com/hashicorp/aws-sdk-go-base/pull/240))

ENHANCEMENTS

* Adds support for assuming IAM role with web identity. ([#178](https://github.com/hashicorp/aws-sdk-go-base/pull/178))

# v2.0.0-beta.15 (2022-04-12)

ENHANCEMENTS

* Adds parameter `SuppressDebugLog` to suppress logging. ([#232](https://github.com/hashicorp/aws-sdk-go-base/pull/232))

# v2.0.0-beta.14 (2022-04-07)

ENHANCEMENTS

* Adds support for custom CA bundles in shared config files for AWS SDK for Go v1. ([#226](https://github.com/hashicorp/aws-sdk-go-base/pull/226))

# v2.0.0-beta.13 (2022-03-09)

NOTES

* Filters CR characters out of AWS SDK for Go v1 logs. ([#174](https://github.com/hashicorp/aws-sdk-go-base/pull/174))

# v2.0.0-beta.12 (2022-03-02)

NOTES

* Filters CR characters out of AWS SDK for Go v2 logs. ([#157](https://github.com/hashicorp/aws-sdk-go-base/pull/157))

# v2.0.0-beta.11 (2022-02-28)

BUG FIXES

* No longer overrides shared config and credentials files when using defaults. ([#151](https://github.com/hashicorp/aws-sdk-go-base/pull/151))

# v2.0.0-beta.10 (2022-02-25)

ENHANCEMENTS

* Adds logging for explicitly set authentication parameters. ([#146](https://github.com/hashicorp/aws-sdk-go-base/pull/146))
* Adds warning log when `Profile` and static credentials environment variables are set. ([#146](https://github.com/hashicorp/aws-sdk-go-base/pull/146))

# v2.0.0-beta.9 (2022-02-23)

BUG FIXES

* Now returns an error if an invalid profile is specified. ([#128](https://github.com/hashicorp/aws-sdk-go-base/pull/128))

ENHANCEMENTS

* Retrieves region from IMDS when credentials sourced from IMDS. ([#131](https://github.com/hashicorp/aws-sdk-go-base/pull/131))

# v2.0.0-beta.8 (2022-02-18)

BUG FIXES

* Restores expansion of `~/` in file paths. ([#118](https://github.com/hashicorp/aws-sdk-go-base/pull/118))
* Fixes error when setting custom CA bundle. ([#122](https://github.com/hashicorp/aws-sdk-go-base/pull/122))

ENHANCEMENTS

* Adds expansion of environment variables in file paths. ([#118](https://github.com/hashicorp/aws-sdk-go-base/pull/118))
* Updates list of valid regions. ([#111](https://github.com/hashicorp/aws-sdk-go-base/pull/111))
* Adds parameter `CustomCABundle`. ([#122](https://github.com/hashicorp/aws-sdk-go-base/pull/122))

# v2.0.0-beta.7 (2022-02-14)

BUG FIXES

* Updates HTTP client to correctly handle IMDS authentication from inside a container. ([#116](https://github.com/hashicorp/aws-sdk-go-base/pull/116))

# v2.0.0-beta.6 (2022-02-09)

BREAKING CHANGES

* Removes config parameter `DebugLogging` and always enables logging.
  Client applications are expected to filter logs by setting log levels. ([#97](https://github.com/hashicorp/aws-sdk-go-base/pull/97))

ENHANCEMENTS

* Adds support for setting maximum retries using environment variable `AWS_MAX_ATTEMPTS`. ([#105](https://github.com/hashicorp/aws-sdk-go-base/pull/105))

# v2.0.0-beta.5 (2022-01-31)

BUG FIXES

* Was not correctly setting additional user-agent string parameters on AWS SDK v1 `Session`. ([#95](https://github.com/hashicorp/aws-sdk-go-base/pull/95))

# v2.0.0-beta.4 (2022-01-31)

ENHANCEMENTS

* Adds support for IPv6 IMDS endpoints with parameter `EC2MetadataServiceEndpointMode` and environment variable `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE`. ([#92](https://github.com/hashicorp/aws-sdk-go-base/pull/92))
* Adds parameter `EC2MetadataServiceEndpoint` and environment variable `AWS_EC2_METADATA_SERVICE_ENDPOINT`.
  Deprecates environment variable `AWS_METADATA_URL`. ([#92](https://github.com/hashicorp/aws-sdk-go-base/pull/92))
* Adds parameter `StsRegion`. ([#91](https://github.com/hashicorp/aws-sdk-go-base/pull/91))
* Adds parameters `UseDualStackEndpoint` and `UseFIPSEndpoint`. ([#88](https://github.com/hashicorp/aws-sdk-go-base/pull/88))

BREAKING CHANGES

* Renames parameter `SkipMetadataApiCheck` to `SkipEC2MetadataApiCheck`. ([#92](https://github.com/hashicorp/aws-sdk-go-base/pull/92))
* Renames assume role parameter `DurationSeconds` to `Duration`. ([#84](https://github.com/hashicorp/aws-sdk-go-base/pull/84))

# v2.0.0-beta.3 (2021-11-03)

ENHANCEMENTS

* Adds parameter `UserAgent` to append to user-agent string. ([#86](https://github.com/hashicorp/aws-sdk-go-base/pull/86))

# v2.0.0-beta.2 (2021-09-27)

ENHANCEMENTS

* Adds parameter `HTTPProxy`. ([#81](https://github.com/hashicorp/aws-sdk-go-base/pull/81))
* Adds parameter `APNInfo` to add APN data to user-agent string. ([#82](https://github.com/hashicorp/aws-sdk-go-base/pull/82))

BREAKING CHANGES

* Moves assume role parameters to `AssumeRole` struct. ([#78](https://github.com/hashicorp/aws-sdk-go-base/pull/78))
