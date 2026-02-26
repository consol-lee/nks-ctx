# Testing Guide

This document explains how to test the kubectl-nks-ctx project.

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Test Specific Package
```bash
go test ./pkg/ncp
go test ./pkg/kubeconfig
```

### Run with Verbose Output
```bash
go test -v ./...
```

### Check Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test
```bash
go test -run TestNewClient ./pkg/ncp
```

## Test Types

### 1. Unit Tests

Test each package's functionality independently.

#### NCP Authentication Tests
```bash
go test -v ./pkg/ncp -run TestGenerateHMACSignature
```

#### Client Tests
```bash
go test -v ./pkg/ncp -run TestClient
```

#### Kubeconfig Manager Tests
```bash
go test -v ./pkg/kubeconfig
```

### 2. Integration Tests

To test in a real environment:

1. **Set Environment Variables**
```bash
export NCLOUD_ACCESS_KEY="your-access-key"
export NCLOUD_SECRET_KEY="your-secret-key"
export NCLOUD_API_GW="https://fin-ncloud.apigw.fin-ntruss.com"
```

2. **Run Integration Tests**
```bash
go test -tags=integration ./pkg/ncp
```

### 3. Tests Using Mocks

HTTP client tests use the `httptest` package to create mock servers.

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Handle mock response
}))
defer server.Close()
```

## Test Writing Guide

### Test File Naming Convention
- Test files follow the `*_test.go` format
- Test functions start with `Test*`
- Benchmarks start with `Benchmark*`

### Test Example

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Test Execution Notes

1. **Environment Variables**: Some tests use environment variables, so make sure they are set before running tests.

2. **Temporary Files**: Tests use `t.TempDir()` to create temporary directories. They are automatically cleaned up after tests.

3. **Parallel Execution**: You can run tests in parallel with `go test -parallel`, but be careful with tests that use shared resources.

## Running Tests in CI/CD

### GitHub Actions Example
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: go test -v ./...
      - run: go test -coverprofile=coverage.out ./...
```

## Debugging

### Debug Specific Test
```bash
# Verbose output
go test -v -run TestSpecificTest ./pkg/ncp

# Re-run only failed tests
go test -v -run TestSpecificTest ./pkg/ncp -count=1
```

### Set Test Timeout
```bash
go test -timeout 30s ./...
```

## Using Makefile for Tests

```bash
make test        # Run all tests
make test-cover  # Run tests with coverage
```
