package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// Static test errors to satisfy err113 linter
var (
	ErrTestOriginal   = errors.New("original error")
	ErrTestConnection = errors.New("connection refused")
	ErrTestExit       = errors.New("exit status 1")
	ErrTestSyntax     = errors.New("syntax error")
	ErrTestFirst      = errors.New("first error")
	ErrTestSecond     = errors.New("second error")
	ErrTestThird      = errors.New("third error")
)

// FactoryTestSuite tests the error factory functionality
type FactoryTestSuite struct {
	testhelpers.BaseSuite

	factory *CommonErrorFactory
}

// SetupSuite configures the test suite
func (ts *FactoryTestSuite) SetupSuite() {
	ts.Options = testhelpers.DefaultOptions()
	ts.Options.CreateTempDir = false // We don't need temp dirs for error tests
	ts.Options.CreateGoModule = false
	ts.BaseSuite.SetupSuite()
}

// SetupTest runs before each test
func (ts *FactoryTestSuite) SetupTest() {
	ts.factory = NewCommonErrorFactory()
}

// TestCommonErrorFactory tests the common error factory patterns
func (ts *FactoryTestSuite) TestCommonErrorFactory() {
	ts.Run("NotFound", func() {
		err := ts.factory.NotFound("user", "123")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "user not found: 123")
	})

	ts.Run("AlreadyExists", func() {
		err := ts.factory.AlreadyExists("file", "test.txt")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "file already exists: test.txt")
	})

	ts.Run("InvalidArgument", func() {
		err := ts.factory.InvalidArgument("email", "invalid-email", "missing @ symbol")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "invalid email 'invalid-email': missing @ symbol")
	})

	ts.Run("Wrap", func() {
		wrappedErr := ts.factory.Wrap(ErrTestOriginal, "operation failed")
		ts.AssertError(wrappedErr)
		ts.Require().Contains(wrappedErr.Error(), "operation failed: original error")
		ts.Require().ErrorIs(wrappedErr, ErrTestOriginal)
	})

	ts.Run("Wrapf", func() {
		wrappedErr := ts.factory.Wrapf(ErrTestConnection, "failed to connect to %s:%d", "localhost", 8080)
		ts.AssertError(wrappedErr)
		ts.Require().Contains(wrappedErr.Error(), "failed to connect to localhost:8080: connection refused")
		ts.Require().ErrorIs(wrappedErr, ErrTestConnection)
	})
}

// TestPackageLevelFunctions tests the package-level convenience functions
func (ts *FactoryTestSuite) TestPackageLevelFunctions() {
	ts.Run("NotFound package function", func() {
		err := NotFound("config", "database.yml")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "config not found: database.yml")
	})

	ts.Run("CommandFailed", func() {
		err := CommandFailed("go", []string{"build", "."}, ErrTestExit)
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "command failed: go build .")
	})

	ts.Run("ValidationFailed", func() {
		err := ValidationFailed("port", "99999", "must be between 1 and 65535")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "validation failed for field 'port' with value '99999': must be between 1 and 65535")
	})
}

// TestDomainSpecificFactories tests domain-specific error factories
func (ts *FactoryTestSuite) TestDomainSpecificFactories() {
	ts.Run("BuildErrors", func() {
		err := GetBuildErrors().CompilationFailed("main.go", ErrTestSyntax)
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "compilation failed for main.go")
	})

	ts.Run("TestErrors", func() {
		err := GetTestErrors().CoverageBelowThreshold(75.5, 80.0)
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "coverage 75.50% below threshold 80.00%")
	})

	ts.Run("SecurityErrors", func() {
		err := GetSecurityErrors().UnauthorizedAccess("admin-panel", "guest")
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "unauthorized access to admin-panel by user guest")
	})
}

// TestErrorChain tests error chaining functionality
func (ts *FactoryTestSuite) TestErrorChain() {
	err1 := ErrTestFirst
	err2 := ErrTestSecond
	err3 := ErrTestThird

	chain := Chain(err1, err2, nil, err3) // nil should be ignored

	ts.Require().Equal(3, chain.Count())
	ts.Require().Equal(err1, chain.First())
	ts.Require().Equal(err3, chain.Last())

	allErrors := chain.ToSlice()
	ts.Require().Len(allErrors, 3)
	ts.Require().Equal(err1, allErrors[0])
	ts.Require().Equal(err2, allErrors[1])
	ts.Require().Equal(err3, allErrors[2])
}

// TestMageErrorIntegration tests integration with the MageError system
func (ts *FactoryTestSuite) TestMageErrorIntegration() {
	ts.Run("WithCode", func() {
		err := ErrorWithCode(ErrFileNotFound, "configuration file missing")
		ts.AssertError(err)

		var mageErr MageError
		ts.Require().ErrorAs(err, &mageErr)
		ts.Require().Equal(ErrFileNotFound, mageErr.Code())
	})

	ts.Run("WithCodef", func() {
		err := ErrorWithCodef(ErrTimeout, "operation timed out after %d seconds", 30)
		ts.AssertError(err)
		ts.Require().Contains(err.Error(), "operation timed out after 30 seconds")

		var mageErr MageError
		ts.Require().ErrorAs(err, &mageErr)
		ts.Require().Equal(ErrTimeout, mageErr.Code())
	})
}

// TestErrorCodeConsistency verifies error codes are correctly applied
func (ts *FactoryTestSuite) TestErrorCodeConsistency() {
	testCases := []struct {
		name         string
		createError  func() error
		expectedCode ErrorCode
	}{
		{
			name:         "NotFound",
			createError:  func() error { return NotFound("item", "123") },
			expectedCode: ErrNotFound,
		},
		{
			name:         "AlreadyExists",
			createError:  func() error { return AlreadyExists("item", "123") },
			expectedCode: ErrAlreadyExists,
		},
		{
			name:         "InvalidArgument",
			createError:  func() error { return InvalidArgument("field", "value", "reason") },
			expectedCode: ErrInvalidArgument,
		},
		{
			name:         "FileNotFound",
			createError:  func() error { return FileNotFound("/path/to/file") },
			expectedCode: ErrFileNotFound,
		},
		{
			name:         "PermissionDenied",
			createError:  func() error { return PermissionDenied("resource", "read") },
			expectedCode: ErrPermissionDenied,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			err := tc.createError()
			ts.AssertError(err)

			code := GetCode(err)
			ts.Require().Equal(tc.expectedCode, code)
		})
	}
}

// TestFactoryTestSuite runs the error factory test suite
func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
