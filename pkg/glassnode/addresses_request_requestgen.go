// Code generated by "requestgen -method GET -type AddressesRequest -url /v1/metrics/addresses/:metric -responseType Response"; DO NOT EDIT.

package glassnode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

func (a *AddressesRequest) SetAsset(Asset string) *AddressesRequest {
	a.Asset = Asset
	return a
}

func (a *AddressesRequest) SetSince(Since int64) *AddressesRequest {
	a.Since = Since
	return a
}

func (a *AddressesRequest) SetUntil(Until int64) *AddressesRequest {
	a.Until = Until
	return a
}

func (a *AddressesRequest) SetInterval(Interval Interval) *AddressesRequest {
	a.Interval = Interval
	return a
}

func (a *AddressesRequest) SetFormat(Format Format) *AddressesRequest {
	a.Format = Format
	return a
}

func (a *AddressesRequest) SetTimestampFormat(TimestampFormat string) *AddressesRequest {
	a.TimestampFormat = TimestampFormat
	return a
}

func (a *AddressesRequest) SetMetric(Metric string) *AddressesRequest {
	a.Metric = Metric
	return a
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (a *AddressesRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}
	// check Asset field -> json key a
	Asset := a.Asset

	// TEMPLATE check-required
	if len(Asset) == 0 {
		return nil, fmt.Errorf("a is required, empty string given")
	}
	// END TEMPLATE check-required

	// assign parameter of Asset
	params["a"] = Asset
	// check Since field -> json key s
	Since := a.Since

	// assign parameter of Since
	params["s"] = Since
	// check Until field -> json key u
	Until := a.Until

	// assign parameter of Until
	params["u"] = Until
	// check Interval field -> json key i
	Interval := a.Interval

	// assign parameter of Interval
	params["i"] = Interval
	// check Format field -> json key f
	Format := a.Format

	// assign parameter of Format
	params["f"] = Format
	// check TimestampFormat field -> json key timestamp_format
	TimestampFormat := a.TimestampFormat

	// assign parameter of TimestampFormat
	params["timestamp_format"] = TimestampFormat

	query := url.Values{}
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (a *AddressesRequest) GetParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

// GetParametersQuery converts the parameters from GetParameters into the url.Values format
func (a *AddressesRequest) GetParametersQuery() (url.Values, error) {
	query := url.Values{}

	params, err := a.GetParameters()
	if err != nil {
		return query, err
	}

	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (a *AddressesRequest) GetParametersJSON() ([]byte, error) {
	params, err := a.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (a *AddressesRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}
	// check Metric field -> json key metric
	Metric := a.Metric

	// assign parameter of Metric
	params["metric"] = Metric

	return params, nil
}

func (a *AddressesRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for k, v := range slugs {
		needleRE := regexp.MustCompile(":" + k + "\\b")
		url = needleRE.ReplaceAllString(url, v)
	}

	return url
}

func (a *AddressesRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := a.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for k, v := range params {
		slugs[k] = fmt.Sprintf("%v", v)
	}

	return slugs, nil
}

func (a *AddressesRequest) Do(ctx context.Context) (Response, error) {

	// no body params
	var params interface{}
	query, err := a.GetQueryParameters()
	if err != nil {
		return nil, err
	}

	apiURL := "/v1/metrics/addresses/:metric"
	slugs, err := a.GetSlugsMap()
	if err != nil {
		return nil, err
	}

	apiURL = a.applySlugsToUrl(apiURL, slugs)

	req, err := a.Client.NewAuthenticatedRequest(ctx, "GET", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := a.Client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse Response
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}
	return apiResponse, nil
}