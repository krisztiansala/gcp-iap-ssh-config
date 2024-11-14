# GCP IAP SSH Config Tool

A command-line utility that simplifies the setup of SSH configurations for Google Cloud Platform (GCP) instances using Identity-Aware Proxy (IAP) tunneling - mainly use for VS Code Remote SSH access.

## Prerequisites

- Go 1.16 or later
- Google Cloud SDK (`gcloud`) installed and configured
- Appropriate IAP permissions configured

## Installation

### Using Go

```bash
go install github.com/krisztiansala/gcp-iap-ssh-config@latest
```
### From Source
```bash
git clone https://github.com/krisztiansala/gcp-iap-ssh-config.git
cd gcp-iap-ssh-config
go build -o setup-ssh-config
```
## Usage
```bash
setup-ssh-config --project PROJECT_ID --instance INSTANCE_NAME --zone ZONE
```
### Required Flags

- `--project, -p`: Your GCP project ID
- `--instance, -i`: The name of your GCP instance
- `--zone, -z`: The zone where your instance is located

### Optional Flags

- `--force, -f`: Force update existing SSH config entry
- `--dry-run`: Print the config without modifying the SSH config file
- `--config`: Path to SSH config file (default: ~/.ssh/config)

### Example
```bash
setup-ssh-config -p my-project-123 -i my-instance -z us-central1-a
```

## Troubleshooting

1. Ensure you have the latest version of Google Cloud SDK installed
2. Verify that IAP is enabled for your project
3. Check that you have the necessary IAP permissions
4. Ensure you are logged in with `gcloud auth login`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Security

This tool manages SSH configurations. Always review the changes made to your SSH config file and ensure they align with your security requirements.

## Related Documentation

- [GCP IAP Documentation](https://cloud.google.com/iap/docs)
- [GCP Compute Engine SSH Documentation](https://cloud.google.com/compute/docs/connect/standard-ssh)