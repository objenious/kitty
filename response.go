package kitty

// NewStatusResponse allows users to respond with a specific HTTP status code
func NewStatusResponse(res interface{}, code int) *StatusResponse {
	return &StatusResponse{
		res:  res,
		code: code,
	}
}

// StatusResponse allows users to respond with a specific HTTP status code
type StatusResponse struct {
	code int
	res  interface{}
}

// StatusCode is to implement httptransport.StatusCoder
func (c *StatusResponse) StatusCode() int {
	return c.code
}
