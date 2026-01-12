// Package mockrec provides helper functions for consolidating repeated
// gomock recorder patterns, reducing boilerplate in mock implementations.
//
// # Overview
//
// The package eliminates the repeated pattern found in gomock-generated code:
//
//	mr.mock.ctrl.T.Helper()
//	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, methodName, reflect.TypeOf((*Interface)(nil).Method), args...)
//
// This pattern appears 71+ times across mock files and is safely consolidated.
//
// # Recording Calls
//
// Use RecordCall for methods with arguments:
//
//	call := mockrec.RecordCall(ctrl, mock, "MethodName", methodType, arg1, arg2)
//	call.Return(expectedResult)
//
// Use RecordNoArgsCall for methods without arguments:
//
//	call := mockrec.RecordNoArgsCall(ctrl, mock, "MethodName", methodType)
//	call.Return(expectedResult)
//
// # Benefits
//
//   - Reduces code duplication across mock implementations
//   - Maintains proper test helper registration
//   - Simplifies mock recorder method implementations
package mockrec
