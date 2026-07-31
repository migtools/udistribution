package registry

import (
	"fmt"

	"github.com/distribution/distribution/v3/registry/api/errcode"
)

// UdistributionError represents abstracted registry errors
type UdistributionError struct {
	Code    string
	Message string
	Detail  interface{}
}

func (e *UdistributionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ConvertDistributionError converts distribution errors to udistribution errors
func ConvertDistributionError(err error) error {
	if err == nil {
		return nil
	}

	if errs, ok := err.(errcode.Errors); ok {
		var udistErrs []error
		for _, e := range errs {
			// Each error in Errors slice needs to be cast back to errcode.Error
			if ecErr, ok := e.(errcode.Error); ok {
				udistErrs = append(udistErrs, &UdistributionError{
					Code:    string(rune(ecErr.Code)),
					Message: ecErr.Message,
					Detail:  ecErr.Detail,
				})
			} else {
				// If it's not an errcode.Error, just wrap it as-is
				udistErrs = append(udistErrs, e)
			}
		}
		return fmt.Errorf("registry errors: %v", udistErrs)
	}

	if e, ok := err.(errcode.Error); ok {
		return &UdistributionError{
			Code:    string(rune(e.Code)),
			Message: e.Message,
			Detail:  e.Detail,
		}
	}

	return err
}

// IsErrorCode checks if an error matches a specific error code
func IsErrorCode(err error, code string) bool {
	if ue, ok := err.(*UdistributionError); ok {
		return ue.Code == code
	}
	return false
}
