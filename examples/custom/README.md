# Custom Tasks Example

This example shows how to extend MAGE-X with your own custom tasks while still benefiting from all the built-in functionality.

## Features Demonstrated

### 1. Custom Namespaces
- `Custom` - Project-specific deployment tasks
- `DB` - Database operations (setup, seed, reset)
- `Docker` - Docker build and deployment
- `Generate` - Code generation tasks

### 2. Environment-Specific Deployment
```bash
# Deploy to different environments
mage customDeploy dev
mage customDeploy staging
mage customDeploy prod
```

### 3. Database Operations
```bash
# Setup database
mage dbSetup

# Rollback last change
mage dbRollback

# Seed database
mage dbSeed

# Reset database (drop, setup, seed)
mage dbReset
```

### 4. Docker Workflows
```bash
# Build Docker image
mage dockerBuild

# Run locally in Docker
mage dockerRun

# Push to registry
mage dockerPush
```

### 5. Code Generation
```bash
# Generate mocks
mage generateMocks

# Generate Swagger docs
mage generateSwagger

# Generate protobuf code
mage generateProto

# Run all generators
mage generateAll
```

### 6. Development Workflow
```bash
# One-time setup
mage setup

# Run with hot reload
mage dev

# Full CI pipeline
mage ci
```

## Key Patterns

### Task Dependencies
```go
// Ensure build is done before deploy
mg.Deps(Build)

// Run tasks in sequence
mg.SerialDeps(Lint, Test, Build)

// Run specific namespace task
mg.Deps(mg.Namespace(Test{}).CI)
```

### Environment Variables
```go
// Check required env vars
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
    return fmt.Errorf("DATABASE_URL is required")
}

// Set env vars for other tasks
os.Setenv("GO_BUILD_TAGS", "prod")
```

### User Confirmation
```go
// Confirm dangerous operations
fmt.Print("Are you sure? (yes/no): ")
var response string
fmt.Scanln(&response)
if response != "yes" {
    return fmt.Errorf("cancelled")
}
```

### Error Handling
```go
// Check if tools exist
if err := sh.Run("which", "air"); err != nil {
    // Install if missing
    sh.Run("go", "install", "github.com/cosmtrek/air@latest")
}
```

## Configuration

Create `.mage.yaml` to configure both MAGE-X and your custom tasks:

```yaml
project:
  name: myapp
  binary: myapp

build:
  tags:
    - prod
  platforms:
    - linux/amd64
    - darwin/amd64

docker:
  registry: myregistry.com
  repository: myorg/myapp

tools:
  custom:
    air: github.com/cosmtrek/air@latest
    dbmate: github.com/amacneil/dbmate@latest
```

## Best Practices

1. **Group Related Tasks**: Use namespaces to organize tasks
2. **Use Dependencies**: Leverage `mg.Deps()` to ensure prerequisites
3. **Provide Feedback**: Use clear output messages
4. **Handle Errors**: Return meaningful error messages
5. **Document Tasks**: Add comments to explain what each task does

## Try It Out

```bash
# See all available tasks
mage -l

# Run setup
mage setup

# Start development
mage dev

# Deploy to staging
mage custom:deploy staging

# Run full CI pipeline
mage ci
```

This example shows just a few possibilities. You can create any custom tasks your project needs while still getting all the benefits of MAGE-X!
