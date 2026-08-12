# Security policy

Please report suspected vulnerabilities through GitHub's private vulnerability
reporting for this repository. Do not include exploit details in a public issue.

Tourminal treats repositories and `.tour` files as untrusted input. It never
executes commands declared by a tour, restricts source paths to the selected
workspace, limits input sizes, and removes terminal control characters from
repository-provided text. Reports that bypass any of these boundaries are
especially useful.
