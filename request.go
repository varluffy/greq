/**
* Created by GoLand.
* User: luffy
* Date: 2019-05-13
* Time: 11:32
 */
package greq

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

const (
	TypeJSON       = "application/json;charset=utf-8"
	TypeXML        = "application/xml;charset=utf-8"
	TypeUrlencoded = "application/x-www-form-urlencoded;charset=utf-8"
	TypeForm       = "application/x-www-form-urlencoded;charset=utf-8"
	TypeFormData   = "application/x-www-form-urlencoded;charset=utf-8"
	TypeHTML       = "text/html;charset=utf-8"
	TypeText       = "text/plain;charset=utf-8"
	TypeMultipart  = "multipart/form-data;charset=utf-8"
	TypeStream     = "application/octet-stream;charset=utf-8"
)

type Request struct {
	target  string
	method  string
	header  http.Header
	params  url.Values
	body    io.Reader
	client  *http.Client
	cookies []*http.Cookie
	proxy   string
	ctx     context.Context
	file    *file
	err     error
	req     *http.Request
}

type file struct {
	name     string
	path     string
	filename string
}

func NewRequest(method, target string) *Request {
	req := &Request{}
	req.header = http.Header{}
	req.SetContentType(TypeUrlencoded)
	req.target = target
	req.method = strings.ToUpper(method)
	req.params = url.Values{}
	req.SetDefaultClient()
	return req
}

func (r *Request) SetContentType(contentType string) {
	r.SetHeader("Content-Type", contentType)
}

func (r *Request) SetHeader(key, value string) {
	r.header.Set(key, value)
}

func (r *Request) AddHeader(key, value string) {
	r.header.Add(key, value)
}

func (r *Request) SetHttpHeader(header http.Header) {
	r.header = header
}

func (r *Request) SetBody(body io.Reader) {
	r.body = body
}

func (r *Request) AddParam(key, value string) {
	r.params.Add(key, value)
}

func (r *Request) SetParam(key, value string) {
	r.params.Set(key, value)
}

func (r *Request) SetParams(params url.Values) {
	r.params = params
}

func (r *Request) SetProxy(proxyURL string) {
	r.proxy = proxyURL
}

func (r *Request) SetFile(name, filename, path string) {
	r.file = &file{
		name:     name,
		filename: filename,
		path:     path,
	}
}

func (r *Request) SetClient(client *http.Client) {
	r.client = client
}

func (r *Request) GetClient() *http.Client {
	if r.client == nil {
		r.SetDefaultClient()
	}
	return r.client
}

func (r *Request) SetDefaultClient() {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	r.client = &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func (r *Request) getTransport() *http.Transport {
	transport, _ := r.GetClient().Transport.(*http.Transport)
	return transport
}

func (r *Request) EnableInsecureTLS(enable bool) {
	transport := r.getTransport()
	if transport == nil {
		return
	}
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	transport.TLSClientConfig.InsecureSkipVerify = enable
}

func (r *Request) SetTimeout(d time.Duration) {
	r.GetClient().Timeout = d
}

func (r *Request) AddCookie(cookie *http.Cookie) {
	r.cookies = append(r.cookies, cookie)
}

func (r *Request) SetCookies(cookies []*http.Cookie) {
	r.cookies = cookies
}

func (r *Request) SetBodyJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		r.err = err
	}
	r.SetBody(bytes.NewBuffer(data))
	r.SetContentType(TypeJSON)
}

func (r *Request) SetBodyXML(v interface{}) {
	data, err := xml.Marshal(v)
	if err != nil {
		r.err = err
	}
	r.SetBody(bytes.NewBuffer(data))
	r.SetContentType(TypeXML)
}
func (r *Request) SetContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *Request) Do() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	var (
		req      *http.Request
		err      error
		body     io.Reader
		rawQuery string
	)
	if r.method == http.MethodGet || r.method == http.MethodHead {
		if len(r.params) > 0 {
			rawQuery = r.params.Encode()
		}
	} else {
		if r.body != nil {
			body = r.body
			if len(r.params) > 0 {
				rawQuery = r.params.Encode()
			}
		} else if r.file != nil {
			file, err := os.Open(r.file.path)
			if err != nil {
				return nil, err
			}
			defer file.Close()
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			fileWriter, err := bodyWriter.CreateFormFile(r.file.name, r.file.filename)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return nil, err
			}
			r.SetContentType(bodyWriter.FormDataContentType())
			for key, values := range r.params {
				for _, value := range values {
					_ = bodyWriter.WriteField(key, value)
				}
			}
			err = bodyWriter.Close()
			if err != nil {
				return nil, err
			}
			body = bodyBuf
		} else if r.params != nil {
			body = strings.NewReader(r.params.Encode())
		}
	}

	if rawQuery != "" {
		if strings.IndexByte(r.target, '?') == -1 {
			r.target = r.target + "?" + rawQuery
		} else {
			r.target = r.target + "&" + r.target
		}
	}
	if r.proxy != "" {
		u, err := url.Parse(r.proxy)
		if err != nil {
			return nil, err
		}
		transport := r.getTransport()
		transport.Proxy = http.ProxyURL(u)
	}

	req, err = http.NewRequest(r.method, r.target, body)
	if err != nil {
		return nil, err
	}

	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	req.Header = r.header

	if len(r.cookies) > 0 {
		for _, cookie := range r.cookies {
			req.AddCookie(cookie)
		}
	}
	r.req = req
	return r.GetClient().Do(req)
}

func (r *Request) Exec() *Response {
	before := time.Now()
	resp, err := r.Do()
	after := time.Now()
	took := after.Sub(before)
	return &Response{req: r.req, resp: resp, took: took, ctx: r.ctx, err: err}
}
