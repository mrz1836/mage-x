# Simple Build Example

This example demonstrates the basic usage of the Go-Mage namespace interface architecture. It shows how to use factory functions to create namespace instances and call their methods.

## Overview

The example provides a simple build script that uses the new namespace interface pattern instead of the old direct struct instantiation.

## Key Features

- **Interface-Based**: Uses `NewBuildNamespace()` instead of `Build{}`
- **Type Safety**: Compile-time verification through interfaces
- **Consistent Pattern**: All namespaces follow the same factory pattern
- **Error Handling**: Proper error propagation through the pipeline

## Usage

```bash
# Individual commands
mage build      # Build the application
mage test       # Run tests
mage lint       # Run linter
mage format     # Format code
mage clean      # Clean build artifacts

# Composite commands
mage all        # Run complete pipeline
mage install    # Build and install
mage prebuild   # Run pre-build tasks
```

## Code Comparison

### Old Pattern (still works)
```go
func Build() error {
    build := mage.Build{}
    return build.Default()
}
```

### New Pattern (recommended)
```go
func Build() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}
```

## Benefits Demonstrated

1. **Cleaner Syntax**: Factory functions provide cleaner instantiation
2. **Interface Contracts**: Clear contracts through interface definitions
3. **Future Extensibility**: Easy to swap implementations later
4. **Better Testing**: Can inject mock implementations for testing

## Files

- `magefile.go` - Complete build script using namespace interfaces
- `README.md` - This documentation

## Next Steps

After understanding this basic example, explore:
- [Multi-Platform Build](../multi-platform/) - Building for multiple platforms
- [CI Pipeline](../ci-pipeline/) - Complete CI/CD pipeline
- [Custom Implementations](../../custom/) - Creating custom namespace implementations

## Migration Note

This example can be gradually migrated from existing magefiles:

1. Keep existing functions working
2. Add new interface-based functions
3. Update existing functions to use new pattern
4. Remove old pattern when comfortable

The migration is completely safe with zero breaking changes.