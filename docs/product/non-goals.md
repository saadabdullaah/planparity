# Version-one non-goals

PlanParity version one deliberately excludes:

- Applying, destroying, importing, remediating, or otherwise changing infrastructure.
- Mutating or managing state, including remote backends and backups.
- Hosted analysis, user accounts, organizations, authentication services, or external databases.
- Cloud-specific provider emulators or a guarantee of zero network activity.
- Terraform/OpenTofu binary downloading or version management.
- CI-platform product integrations, policy languages, signed attestations, SARIF, and PDF reports.
- Plugins, Terragrunt, Pulumi, CloudFormation, Bicep, and other IaC engines.
- AI-generated conclusions.
- macOS and mobile support.

These exclusions are product boundaries, not an unprioritized backlog. Adding one before version one requires explicitly removing or deferring comparable scope.

