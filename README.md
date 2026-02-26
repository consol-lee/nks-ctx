# kubectl nks-ctx

A kubectl plugin to manage [Ncloud Kubernetes Service (NKS)](https://www.ncloud.com/v2/product/containers/nks) cluster contexts. Syncs all NKS clusters in your NCP account to kubeconfig and lets you switch between them — similar to [kubectx](https://github.com/ahmetb/kubectx).

## Overview

`nks-ctx` discovers NKS clusters across all supported regions (Korea, Singapore, Japan for public; capital/southern for gov) and configures your kubeconfig so you can switch contexts with a single command.

- **List and sync** all NKS clusters from the NCP API into `~/.kube/config`
- **Switch context** by cluster name: `kubectl nks-ctx <cluster-name>`
- **Multi-environment**: use `--profile` for finance or government NCP profiles
- **Skips** clusters already in kubeconfig to avoid duplicate entries

## Installation

### Via Krew

```bash
kubectl krew install nks-ctx
```

### Manual Installation

1. Download the latest release for your platform from the [Releases](../../releases) page.
2. Extract the executable and place it in your PATH:

```bash
# Example: extract zip and move to PATH
unzip kubectl-nks_ctx-linux-amd64.zip
chmod +x kubectl-nks_ctx
sudo mv kubectl-nks_ctx /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/consol-lee/nks-ctx.git
cd nks-ctx
make build
make install
```

## Prerequisites

- **kubectl**
- **[ncp-iam-authenticator](https://github.com/NaverCloudPlatform/ncp-iam-authenticator)** installed and in PATH
- **NCP API credentials** (Access Key / Secret Key)

## How It Works

1. **Load credentials** from environment variables or `~/.ncloud/configure`.
2. **List clusters** by calling the NKS API for each region (KR, SGN, JPN for public; v2, krs-v2 for gov).
3. **Update kubeconfig** via `ncp-iam-authenticator` for each cluster (skips if already present).
4. **Display** the cluster list; `*` marks the current context.

Example:

```bash
$ kubectl nks-ctx
Synced 2 cluster(s), skipped 1 already configured. (3 total)

  my-cluster-dev
* my-cluster-staging
  my-cluster-prod

$ kubectl nks-ctx my-cluster-prod
Switched to context "my-cluster-prod"
```

## Configuration

Credentials are read from:

**1. Environment variables**

```bash
export NCLOUD_ACCESS_KEY="your-access-key"
export NCLOUD_SECRET_KEY="your-secret-key"
export NCLOUD_API_GW="https://ncloud.apigw.ntruss.com"
```

**2. NCP config file (`~/.ncloud/configure`)**

```ini
[DEFAULT]
ncloud_access_key_id=your-access-key
ncloud_secret_access_key=your-secret-key
ncloud_api_url=https://ncloud.apigw.ntruss.com

[finance]
ncloud_access_key_id=finance-access-key
ncloud_secret_access_key=finance-secret-key
ncloud_api_url=https://fin-ncloud.apigw.fin-ntruss.com
```

Use a profile with `--profile`:

```bash
kubectl nks-ctx --profile finance
```

Kubeconfig is stored at `~/.kube/config` (or `$KUBECONFIG`). Existing entries from other providers are preserved; only NKS cluster entries are added or updated.

## Development

### Building

```bash
make build              # Current platform only
make build-all          # All platforms: darwin-amd64, darwin-arm64, linux-amd64
```

### Testing

```bash
make test
make test-verbose
make test-cover
make test-linux         # Run tests in a Linux container (podman/docker)
```

See [TESTING.md](TESTING.md) for details.

### Release assets

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
make release-assets
```

Produces in `dist/`: source zip and tar.gz, binary zips per platform, and `checksums.txt`. Upload these to your release page.

## License

This project is licensed under the [MIT License](LICENSE). Dependencies use Apache-2.0 or compatible licenses.

## About

kubectl plugin to manage NKS (Ncloud Kubernetes Service) cluster contexts with NCP IAM authentication.
