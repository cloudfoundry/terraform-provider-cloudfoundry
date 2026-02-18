package mta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var (
	jsonCheck = regexp.MustCompile("(?i:(?:application|text)/json)")
)

// APIClient manages communication with the MTA REST API API v1.3.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.
	// API Services
	DefaultApi *DefaultApiService
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.DefaultApi = (*DefaultApiService)(&c.common)

	return c
}

// parameterToString convert interface{} parameters to string, using a delimiter if format is provided.
func parameterToString(obj any, collectionFormat string) string {
	var delimiter string

	switch collectionFormat {
	case "csv":
		delimiter = ","
	}

	if reflect.TypeOf(obj).Kind() == reflect.Slice {
		return strings.Trim(strings.ReplaceAll(fmt.Sprint(obj), " ", delimiter), "[]")
	}

	return fmt.Sprintf("%v", obj)
}

// callAPI do the request.
func (c *APIClient) callAPI(request *http.Request) (resp *http.Response, err error) {
	if request.Method == "POST" {
		resp, err = c.DefaultApi.GetCsrfToken(request.Context())
		if err != nil {
			return resp, err
		}
		csrfToken := resp.Header.Values("x-csrf-token")[0]
		cookies := resp.Cookies()
		cookieValue := cookies[0].Name + "=" + cookies[0].Value + "; " + cookies[1].Name + "=" + cookies[1].Value
		request.Header.Add("x-csrf-token", csrfToken)
		request.Header.Add("Cookie", cookieValue)
	}
	resp, err = c.cfg.HTTPClient.Do(request)
	return resp, err
}

// Change base path to allow switching to mocks.
func (c *APIClient) ChangeBasePath(path string) {
	c.cfg.BasePath = path
}

// prepareRequest build the request.
func (c *APIClient) prepareRequest(
	ctx context.Context,
	path string, method string,
	postBody any,
	headerParams map[string]string,
	queryParams url.Values,
	formParams url.Values,
	fileName string,
	fileBytes []byte) (localVarRequest *http.Request, err error) {

	var body *bytes.Buffer

	// Detect postBody type and post.
	if postBody != nil {
		contentType := headerParams["Content-Type"]
		if contentType == "" {
			contentType = detectContentType(postBody)
			headerParams["Content-Type"] = contentType
		}

		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}

	// add form parameters and file if available.
	if strings.HasPrefix(headerParams["Content-Type"], "multipart/form-data") && len(formParams) > 0 || (len(fileBytes) > 0 && fileName != "") {
		if body != nil {
			return nil, errors.New("cannot specify postBody and multipart form at the same time")
		}
		body = &bytes.Buffer{}
		w := multipart.NewWriter(body)

		if len(fileBytes) > 0 && fileName != "" {
			w.Boundary()
			//_, fileNm := filepath.Split(fileName)
			part, err := w.CreateFormFile("file", filepath.Base(fileName))
			if err != nil {
				return nil, err
			}
			_, err = part.Write(fileBytes)
			if err != nil {
				return nil, err
			}
			// Set the Boundary in the Content-Type
			headerParams["Content-Type"] = w.FormDataContentType()
		}

		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
		w.Close()
	}

	// Setup path and query parameters
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Adding Query Param
	query := url.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	url.RawQuery = query.Encode()

	// Generate a new request
	if body != nil {
		localVarRequest, err = http.NewRequest(method, url.String(), body)
	} else {
		localVarRequest, err = http.NewRequest(method, url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers.Set(h, v)
		}
		localVarRequest.Header = headers
	}

	// Override request host, if applicable
	if c.cfg.Host != "" {
		localVarRequest.Host = c.cfg.Host
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.cfg.UserAgent)

	if ctx != nil {
		// add context to the request
		localVarRequest = localVarRequest.WithContext(ctx)
	}

	for header, value := range c.cfg.DefaultHeader {
		localVarRequest.Header.Add(header, value)
	}

	return localVarRequest, nil
}

func (c *APIClient) decode(v any, b []byte, contentType string) (err error) {
	if strings.Contains(contentType, "application/json") {
		if err = json.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	}
	return errors.New("undefined response type")
}

func (c *APIClient) returnResponse(resp *http.Response, returnValue any, varBody []byte) error {
	if resp.StatusCode == 204 {
		return nil
	} else if resp.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err := c.decode(&returnValue, varBody, resp.Header.Get("Content-Type"))
		return err
	} else {
		newErr := GenericError{
			body:  varBody,
			error: resp.Status,
		}
		return newErr
	}
}

func (c *APIClient) sendRequestGetResponse(ctx context.Context, request Request, returnValue any) (operationId string, localVarHttpResponse *http.Response, err error) {
	r, err := c.prepareRequest(ctx, request.path, request.method, request.postBody, request.headers, request.queryParams, request.formParams, request.fileName, request.fileBytes)
	if err != nil {
		return operationId, nil, err
	}

	localVarHttpResponse, err = c.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return operationId, localVarHttpResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return operationId, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode == 202 {
		operationId, err = decodeOperationJobID(localVarHttpResponse)
		return operationId, localVarHttpResponse, err
	}

	err = c.returnResponse(localVarHttpResponse, &returnValue, localVarBody)
	return operationId, localVarHttpResponse, err
}

func (c *APIClient) get(ctx context.Context, request Request, returnValue any) (localVarHttpResponse *http.Response, err error) {
	request.method = GET
	_, localVarHttpResponse, err = c.sendRequestGetResponse(ctx, request, &returnValue)
	return localVarHttpResponse, err
}

func (c *APIClient) post(ctx context.Context, request Request, returnValue any) (operationId string, localVarHttpResponse *http.Response, err error) {
	request.method = POST
	return c.sendRequestGetResponse(ctx, request, &returnValue)
}

// Set request body from an interface{}.
func setBody(body any, contentType string) (bodyBuf *bytes.Buffer, err error) {
	bodyBuf = &bytes.Buffer{}

	if jsonCheck.MatchString(contentType) {
		err = json.NewEncoder(bodyBuf).Encode(body)
	}
	if err != nil {
		return nil, err
	}
	if bodyBuf.Len() == 0 {
		err = fmt.Errorf("invalid body type %s", contentType)
		return nil, err
	}
	return bodyBuf, nil
}

// detectContentType method is used to figure out `Request.Body` content type for request header.
func detectContentType(body any) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Pointer:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

// GenericError Provides access to the body, error and model on returned errors.
type GenericError struct {
	body  []byte
	error string
	model any
}

// Error returns non-empty string if there was an error.
func (e GenericError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response.
func (e GenericError) Body() []byte {
	return e.body
}

// Model returns the unpacked model of the error.
func (e GenericError) Model() any {
	return e.model
}
