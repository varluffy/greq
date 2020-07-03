/**
* Created by GoLand.
* User: luffy
* Date: 2019-05-13
* Time: 11:40
 */
package greq

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type Response struct {
	req      *http.Request
	resp     *http.Response
	respBody []byte
	took     time.Duration
	ctx      context.Context
	err      error
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Took() time.Duration {
	return r.took
}

func (r *Response) Request() *http.Request {
	return r.req
}

func (r *Response) Response() *http.Response {
	return r.resp
}

func (r *Response) Context() context.Context {
	return r.ctx
}

func (r *Response) StatusCode() int {
	if r.resp != nil {
		return r.resp.StatusCode
	}
	return http.StatusServiceUnavailable
}

func (r *Response) ToBytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.respBody != nil {
		return r.respBody, nil
	}
	defer r.resp.Body.Close()
	body, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.respBody = body
	return body, err
}

func (r *Response) ToString() (string, error) {
	bytes, err := r.ToBytes()
	return string(bytes), err
}

func (r *Response) ToJSON(v interface{}) error {
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	return d.Decode(v)
}

func (r *Response) ToXML(v interface{}) error {
	bs, err := r.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(bs, v)
}

func (r *Response) Cookies() []*http.Cookie {
	if r.resp != nil {
		return r.resp.Cookies()
	}
	return nil
}

func (r *Response) Header() http.Header {
	if r.resp != nil {
		return r.resp.Header
	}
	return http.Header{}
}

func (r *Response) DumpRequest(body bool) ([]byte, error) {
	return httputil.DumpRequest(r.req, body)
}

func (r *Response) DumpResponse(body bool) ([]byte, error) {
	return httputil.DumpResponse(r.resp, body)
}
