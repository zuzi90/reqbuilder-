package reqbuilder

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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

func New(t *testing.T) *Builder {
	return &Builder{
		client:  &http.Client{},
		t:       t,
		require: require.New(t),
	}

}

// Request sends a POST request to the specified endpoint
func (b *Builder) Request(
	ctx context.Context,
	method string,
	host,
	endpoint string,
	reqBody []byte,
	result any,
	cookies *[]*http.Cookie,
	headers map[string]string,
	authorization string) (*http.Response, []*http.Cookie) {
	b.t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, host+endpoint, bytes.NewReader(reqBody))
	if err != nil {
		b.t.Log(err)
	}

	b.require.NoError(err)

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if cookies != nil && len(*cookies) != 0 {
		req.Header.Set("Authorization", authorization)
		for _, cookie := range *cookies {
			req.AddCookie(cookie)
		}
	}

	response, err := b.client.Do(req)
	if err != nil {
		b.t.Log(err)
	}
	b.require.NoError(err)

	defer func() {
		if err = response.Body.Close(); err != nil {
			b.t.Log(err)
		}
	}()

	respBody, err := b.readResponseBody(response)

	// Создаем мапу для быстрого поиска кук
	cookieMap := make(map[string]*http.Cookie)

	// Добавляем все куки из ответа сервера
	for _, c := range response.Cookies() {
		cookieMap[c.Name] = c
	}

	// Добавляем только те куки из `cookies`, которых еще нет в `cookieMap`
	if cookies != nil && len(*cookies) != 0 {
		for _, c := range *cookies {
			if _, exists := cookieMap[c.Name]; !exists {
				cookieMap[c.Name] = c
			}
		}
	}

	// Преобразуем карту обратно в срез
	allCookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, c := range cookieMap {
		allCookies = append(allCookies, c)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		return response, allCookies
	}

	err = json.Unmarshal(respBody, result)
	b.require.NoError(err)

	return response, allCookies
}

type BrotliReadCloser struct {
	*brotli.Reader
	io.Closer
}

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
	authorization string,
	result any) (*http.Response, []*http.Cookie) {
	t.Helper()

	var req *http.Request
	var err error

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err = writer.WriteField(formData, string(requestBody))
	if err != nil {
		b.t.Log(err)
		b.require.NoError(err)
	}
	func() {
		if err = writer.Close(); err != nil {
			b.t.Log(err)
			b.require.NoError(err)
		}
	}()

	// Создаём запрос с `multipart/form-data`
	req, err = http.NewRequestWithContext(ctx, method, host+endpoint, body)
	if err != nil {
		b.t.Log(err)
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

	defer func() {
		if err = response.Body.Close(); err != nil {
			b.t.Log(err)
			b.require.NoError(err)
		}
	}()

	// Создаем мапу для быстрого поиска кук
	cookieMap := make(map[string]*http.Cookie)

	// Добавляем все куки из ответа сервера
	for _, c := range response.Cookies() {
		cookieMap[c.Name] = c
	}

	// Добавляем только те куки из `cookies`, которых еще нет в `cookieMap`
	if cookies != nil && len(cookies) != 0 {
		for _, c := range cookies {
			if _, exists := cookieMap[c.Name]; !exists {
				cookieMap[c.Name] = c
			}
		}
	}

	// Преобразуем карту обратно в срез
	allCookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, c := range cookieMap {
		allCookies = append(allCookies, c)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		return response, allCookies
	}

	respBody, err := b.readResponseBody(response)
	b.require.NoError(err)

	err = json.Unmarshal(respBody, result)
	b.require.NoError(err)

	return response, allCookies

}

func (b *Builder) RequestWithoutBody() (*http.Response, []*http.Cookie) {

	return nil, nil
}

// readResponseBody decodes the response body and returns it as a byte slice.
func (b *Builder) readResponseBody(response *http.Response) ([]byte, error) {
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

type UserSignIn struct {
	Email       string `json:"email,omitempty"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type SignIn struct {
	Data string `json:"data"`
}
