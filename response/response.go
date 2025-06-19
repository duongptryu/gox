package response

import (
	"github.com/gin-gonic/gin"
)

// successRes represents the simplified API response structure
type successRes struct {
	IsError bool        `json:"is_error"`
	Data    interface{} `json:"data"`
	Paging  interface{} `json:"paging,omitempty"`
	Filter  interface{} `json:"filter,omitempty"`
}

// NewSuccessResponse creates a new success response with data, paging, and filter
func NewSuccessResponse(data, paging, filter interface{}) *successRes {
	return &successRes{IsError: false, Data: data, Paging: paging, Filter: filter}
}

// NewSimpleSuccessResponse creates a simple success response with just data
func NewSimpleSuccessResponse(data interface{}) *successRes {
	return NewSuccessResponse(data, nil, nil)
}

// errorRes represents the error response structure
type errorRes struct {
	IsError bool        `json:"is_error"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details interface{}) *errorRes {
	return &errorRes{
		IsError: true,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// JSON sends the error response as JSON with the specified status code
func (r *errorRes) JSON(c *gin.Context, statusCode int) {
	c.JSON(statusCode, r)
}
