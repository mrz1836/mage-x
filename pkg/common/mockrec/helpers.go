// Package mockrec provides helper functions for consolidating repeated gomock recorder patterns
package mockrec

import (
	"reflect"

	"go.uber.org/mock/gomock"
)

// RecordCall is a helper function that consolidates the repeated pattern:
//
//	mr.mock.ctrl.T.Helper()
//	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, methodName, reflect.TypeOf((*Interface)(nil).Method), args...)
//
// This pattern appears 71+ times across mock files and can be safely consolidated.
func RecordCall(ctrl *gomock.Controller, receiver interface{}, methodName string, methodType reflect.Type, args ...interface{}) *gomock.Call {
	ctrl.T.Helper()
	return ctrl.RecordCallWithMethodType(receiver, methodName, methodType, args...)
}

// RecordNoArgsCall is a helper for methods with no arguments
func RecordNoArgsCall(ctrl *gomock.Controller, receiver interface{}, methodName string, methodType reflect.Type) *gomock.Call {
	ctrl.T.Helper()
	return ctrl.RecordCallWithMethodType(receiver, methodName, methodType)
}
