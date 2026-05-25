package mage

import (
	"reflect"
	"testing"
)

// TestCoreNamespaceInterfaces verifies that core namespaces implement their interfaces
func TestCoreNamespaceInterfaces(t *testing.T) {
	tests := []struct {
		name           string
		implementation any
		interfaceType  reflect.Type
	}{
		{
			name:           "BuildNamespace",
			implementation: Build{},
			interfaceType:  reflect.TypeOf((*BuildNamespace)(nil)).Elem(),
		},
		{
			name:           "TestNamespace",
			implementation: Test{},
			interfaceType:  reflect.TypeOf((*TestNamespace)(nil)).Elem(),
		},
		{
			name:           "LintNamespace",
			implementation: Lint{},
			interfaceType:  reflect.TypeOf((*LintNamespace)(nil)).Elem(),
		},
		{
			name:           "FormatNamespace",
			implementation: Format{},
			interfaceType:  reflect.TypeOf((*FormatNamespace)(nil)).Elem(),
		},
		{
			name:           "DocsNamespace",
			implementation: Docs{},
			interfaceType:  reflect.TypeOf((*DocsNamespace)(nil)).Elem(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			implType := reflect.TypeOf(tt.implementation)
			if !implType.Implements(tt.interfaceType) {
				t.Errorf("%s does not implement %s interface", implType, tt.interfaceType)
			}
		})
	}
}

// TestNamespaceFactories verifies that factory functions work
func TestNamespaceFactories(t *testing.T) {
	tests := []struct {
		name    string
		factory func() any
		check   func(any) bool
	}{
		{
			name:    "NewBuildNamespace",
			factory: func() any { return NewBuildNamespace() },
			check:   func(v any) bool { _, ok := v.(BuildNamespace); return ok },
		},
		{
			name:    "NewTestNamespace",
			factory: func() any { return NewTestNamespace() },
			check:   func(v any) bool { _, ok := v.(TestNamespace); return ok },
		},
		{
			name:    "NewLintNamespace",
			factory: func() any { return NewLintNamespace() },
			check:   func(v any) bool { _, ok := v.(LintNamespace); return ok },
		},
		{
			name:    "NewFormatNamespace",
			factory: func() any { return NewFormatNamespace() },
			check:   func(v any) bool { _, ok := v.(FormatNamespace); return ok },
		},
		{
			name:    "NewDocsNamespace",
			factory: func() any { return NewDocsNamespace() },
			check:   func(v any) bool { _, ok := v.(DocsNamespace); return ok },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := tt.factory()
			if ns == nil {
				t.Errorf("%s returned nil", tt.name)
				return
			}

			if !tt.check(ns) {
				t.Errorf("%s returned wrong type: %T", tt.name, ns)
			}
		})
	}
}
