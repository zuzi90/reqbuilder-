package reqbuilder

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

// Builder is a helper for sending HTTP requests in tests.
type Builder struct {
	client  *http.Client
	t       *testing.T
	require *require.Assertions
}

func New(require *require.Assertions) *Builder {
	return &Builder{
		client:  &http.Client{},
		require: require,
	}

}

// Request sends a POST request to the specified endpoint.
func (b *Builder) Request(
	t *testing.T,
	ctx context.Context,
	method string,
	host,
	endpoint string,
	reqBody []byte,
	cookies []*http.Cookie,
	headers map[string]string,
	authorization string) (*http.Response, []*http.Cookie) {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, host+endpoint, bytes.NewReader(reqBody))
	if err != nil {
		t.Log(err)
	}

	b.require.NoError(err)

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if cookies != nil && len(cookies) != 0 {
		req.Header.Set("Authorization", authorization)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	response, err := b.client.Do(req)
	if err != nil {
		t.Log(err)
	}
	b.require.NoError(err)

	cookieMap := make(map[string]*http.Cookie)

	for _, c := range response.Cookies() {
		cookieMap[c.Name] = c
	}

	if cookies != nil && len(cookies) != 0 {
		for _, c := range cookies {
			if _, exists := cookieMap[c.Name]; !exists {
				cookieMap[c.Name] = c
			}
		}
	}

	allCookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, c := range cookieMap {
		allCookies = append(allCookies, c)
	}

	return response, allCookies
}

type BrotliReadCloser struct {
	*brotli.Reader
	io.Closer
}

// MultipartRequest sends a request with a `multipart/form-data` body to the specified endpoint.
func (b *Builder) MultipartRequest(
	t *testing.T,
	ctx context.Context,
	method,
	host,
	endpoint string,
	requestBody []byte,
	formData string,
	cookies []*http.Cookie,
	headers map[string]string,
	authorization string) (*http.Response, []*http.Cookie) {
	t.Helper()

	var req *http.Request
	var err error

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err = writer.WriteField(formData, string(requestBody))
	if err != nil {
		t.Log(err)
		b.require.NoError(err)
	}
	func() {
		if err = writer.Close(); err != nil {
			t.Log(err)
			b.require.NoError(err)
		}
	}()

	req, err = http.NewRequestWithContext(ctx, method, host+endpoint, body)
	if err != nil {
		t.Log(err)
		b.require.NoError(err)
	}

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if cookies != nil && len(cookies) != 0 {
		req.Header.Set("Authorization", authorization)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	response, err := b.client.Do(req)
	b.require.NoError(err)

	// Create a map for quick cookie search
	cookieMap := make(map[string]*http.Cookie)

	// Add all cookies from the server response
	for _, c := range response.Cookies() {
		cookieMap[c.Name] = c
	}

	// Add only those cookies from `cookies` that are not yet in `cookieMap`
	if cookies != nil && len(cookies) != 0 {
		for _, c := range cookies {
			if _, exists := cookieMap[c.Name]; !exists {
				cookieMap[c.Name] = c
			}
		}
	}

	// Convert the map back to a slice
	allCookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, c := range cookieMap {
		allCookies = append(allCookies, c)
	}

	return response, allCookies

}

// RequestWithoutBody sends a request without a body to the specified endpoint.
func (b *Builder) RequestWithoutBody(
	t *testing.T,
	ctx context.Context,
	method,
	host,
	endpoint string,
	headers map[string]string,
	cookies []*http.Cookie,
	authorization string,
) (*http.Response, []*http.Cookie) {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, host+endpoint, nil)
	if err != nil {
		t.Log(err)
	}

	b.require.NoError(err)

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if cookies != nil && len(cookies) != 0 {
		req.Header.Set("Authorization", authorization)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	response, err := b.client.Do(req)
	if err != nil {
		t.Log(err)
	}

	b.require.NoError(err)

	cookieMap := make(map[string]*http.Cookie)

	for _, c := range response.Cookies() {
		cookieMap[c.Name] = c
	}

	if cookies != nil && len(cookies) != 0 {
		for _, c := range cookies {
			if _, exists := cookieMap[c.Name]; !exists {
				cookieMap[c.Name] = c
			}
		}
	}

	allCookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, c := range cookieMap {
		allCookies = append(allCookies, c)
	}

	return response, allCookies
}

// SignIn sends a request to the specified endpoint and returns the response and cookies.
func (b *Builder) SignIn(
	t *testing.T,
	ctx context.Context,
	method,
	host,
	endpoint string,
	requestBody []byte,
	headers map[string]string,
) (*http.Response, []*http.Cookie) {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, host+endpoint, bytes.NewReader(requestBody))
	if err != nil {
		t.Log(err)
	}

	b.require.NoError(err)

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	response, err := b.client.Do(req)
	if err != nil {
		t.Log(err)
		b.require.NoError(err)
	}

	return response, response.Cookies()
}

// ReadResponseBody decodes the response body and returns it as a byte slice.
func (b *Builder) ReadResponseBody(response *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	case "br":
		reader = BrotliReadCloser{
			Reader: brotli.NewReader(response.Body),
			Closer: response.Body,
		}
	case "zstd":
		decoder, err := zstd.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer decoder.Close()
		reader = io.NopCloser(decoder)
	case "deflate":
		reader = flate.NewReader(response.Body)
		defer reader.Close()
	default:
		reader = response.Body
	}

	return io.ReadAll(reader)
}
