package api

// NewResponse constructs a Response with the provided status code and body payload.
func NewResponse(code int, body interface{}) *Response {
	return &Response{
		Code: code,
		body: body,
	}
}

// SetBody updates the response body payload and returns the response for chaining.
func (resp *Response) SetBody(body interface{}) *Response {
	resp.body = body
	return resp
}
