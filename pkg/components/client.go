/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心检索引擎
 * (BlueKing-IAM-Search-Engine) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package components

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/iam-go-sdk/client"
	"github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/TencentBlueKing/iam-go-sdk/util"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/parnurzeal/gorequest"

	"engine/pkg/types"
)

const (
	bkIAMVersion = "1"

	defaultTimeout = 5 * time.Second
)

// Method is the type of http method
type Method string

// POST ...
var (
	// POST http post
	POST Method = "POST"
	// GET http get
	GET Method = "GET"
)

// IAMBackendBaseResponse is the struct of iam backend response
type IAMBackendBaseResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Error will check if the response with error
func (r *IAMBackendBaseResponse) Error() error {
	if r.Code == 0 {
		return nil
	}

	return fmt.Errorf("response error[code=`%d`,  message=`%s`]", r.Code, r.Message)
}

// String will return the detail text of the response
func (r *IAMBackendBaseResponse) String() string {
	return fmt.Sprintf("response[code=`%d`, message=`%s`, data=`%s`]", r.Code, r.Message, conv.BytesToString(r.Data))
}

// IAMBackendClient is the interface of iam backend client
type IAMBackendClient interface {
	Ping() error

	GetMaxIDBeforeUpdate(updatedAt int64) (int64, error)
	ListPolicyIDBetweenUpdateAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error)
	ListPolicyBetweenID(timestamp, minID, maxID int64) ([]types.Policy, error)
	ListPolicyByIDs(ids []int64) ([]types.Policy, error)

	GetSystem(systemID string) (System, error)

	CredentialsVerify(appCode, appSecret string) (exists bool, err error)
}

type iamBackendClient struct {
	Host string

	System    string
	appCode   string
	appSecret string
}

// ListPolicyResponse ...
type ListPolicyResponse struct {
	Results []types.Policy
}

// ListPolicyIDResponse ...
type ListPolicyIDResponse struct {
	IDs []int64
}

// GetMaxIDResponse ...
type GetMaxIDResponse struct {
	ID int64
}

// CredentialsVerifyResponse ...
type CredentialsVerifyResponse struct {
	Valid bool
}

// NewIAMBackendClient will create a iam backend client
func NewIAMBackendClient(host string, appCode string, appSecret string) IAMBackendClient {
	return &iamBackendClient{
		Host: host,

		appCode:   appCode,
		appSecret: appSecret,
	}
}

func (c *iamBackendClient) call(
	method Method, path string,
	data interface{},
	timeout int64,
	responseData interface{},
) error {
	callTimeout := time.Duration(timeout) * time.Second
	if timeout == 0 {
		callTimeout = defaultTimeout
	}

	headers := map[string]string{
		"X-BK-APP-CODE":    c.appCode,
		"X-BK-APP-SECRET":  c.appSecret,
		"X-Bk-IAM-Version": bkIAMVersion,
	}

	url := fmt.Sprintf("%s%s", c.Host, path)
	start := time.Now()
	callbackFunc := client.NewMetricCallback("IAMBackend", start)

	logger.Debugf("do http request: method=`%s`, url=`%s`, data=`%s`", method, url, data)

	// request := gorequest.New().Timeout(callTimeout).Post(url).Type("json")
	request := gorequest.New().Timeout(callTimeout).Type("json")
	switch method {
	case POST:
		request = request.Post(url).Send(data)
	case GET:
		request = request.Get(url).Query(data)
	}

	// set headers
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	// do request
	baseResult := IAMBackendBaseResponse{}
	resp, _, errs := request.
		EndStruct(&baseResult, callbackFunc)

	duration := time.Since(start)

	// logFailHTTPRequest(request, resp, errs, &baseResult)

	logger.Debugf("http request result: %+v", baseResult.String())
	logger.Debugf("http request took %v ms", float64(duration/time.Millisecond))

	if len(errs) != 0 {
		return fmt.Errorf("gorequest errors=`%s`", errs)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gorequest statusCode is %d not 200", resp.StatusCode)
	}

	if baseResult.Code != 0 {
		return errors.New(baseResult.Message)
	}

	err := jsoniter.Unmarshal(baseResult.Data, responseData)
	if err != nil {
		return fmt.Errorf("http request response body data not valid: %w, data=`%v`", err, baseResult.Data)
	}
	return nil
}

func (c *iamBackendClient) callWithReturnMapData(
	method Method, path string,
	data interface{},
	timeout int64,
) (map[string]interface{}, error) {
	var responseData map[string]interface{}
	err := c.call(method, path, data, timeout, &responseData)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return responseData, nil
}

// Ping will check the iam backend service is ping-able
func (c *iamBackendClient) Ping() (err error) {
	url := fmt.Sprintf("%s%s", c.Host, "/ping")

	resp, _, errs := gorequest.New().Timeout(defaultTimeout).Get(url).EndBytes()
	if len(errs) != 0 {
		return fmt.Errorf("ping fail! errs=%v", errs)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping fail! status_code=%d", resp.StatusCode)
	}
	return nil
}

// GetMaxIDBeforeUpdate ...
func (c *iamBackendClient) GetMaxIDBeforeUpdate(updatedAt int64) (int64, error) {
	path := "/api/v1/engine/policies/ids/max"
	query := map[string]interface{}{
		"updated_at": updatedAt,
		"type":       policyAPIType,
	}

	data, err := c.callWithReturnMapData(GET, path, query, 10)
	if err != nil {
		return -1, err
	}

	maxID := GetMaxIDResponse{}
	err = mapstructure.Decode(data, &maxID)
	if err != nil {
		return -1, err
	}
	return maxID.ID, nil
}

// ListPolicyIDBetweenUpdateAt ...
func (c *iamBackendClient) ListPolicyIDBetweenUpdateAt(
	beginUpdatedAt,
	endUpdatedAt int64,
) ([]int64, error) {
	path := "/api/v1/engine/policies/ids"
	query := map[string]interface{}{
		"begin_updated_at": beginUpdatedAt,
		"end_updated_at":   endUpdatedAt,
		"type":             policyAPIType,
	}
	data, err := c.callWithReturnMapData(GET, path, query, 10)
	if err != nil {
		return nil, err
	}

	ids := ListPolicyIDResponse{}
	err = mapstructure.Decode(data, &ids)
	if err != nil {
		return nil, err
	}

	return ids.IDs, nil
}

// ListPolicyBetweenID 查询指定id之间的策略数据
func (c *iamBackendClient) ListPolicyBetweenID(
	timestamp,
	minID,
	maxID int64,
) ([]types.Policy, error) {
	path := "/api/v1/engine/policies"
	query := map[string]interface{}{
		"timestamp": timestamp,
		"min_id":    minID,
		"max_id":    maxID,
		"type":      policyAPIType,
	}

	data, err := c.callWithReturnMapData(GET, path, query, 10)
	if err != nil {
		return nil, err
	}

	responsePolicies := ListPolicyResponse{}
	err = mapstructure.Decode(data, &responsePolicies)
	if err != nil {
		return nil, err
	}

	return responsePolicies.Results, nil
}

// ListPolicyByIDs 查询指定id的策略数据
func (c *iamBackendClient) ListPolicyByIDs(ids []int64) ([]types.Policy, error) {
	path := "/api/v1/engine/policies"
	query := map[string]interface{}{
		"ids":  util.Int64ArrayToString(ids, ","),
		"type": policyAPIType,
	}
	data, err := c.callWithReturnMapData(GET, path, query, 10)
	if err != nil {
		return nil, err
	}

	responsePolicies := ListPolicyResponse{}
	err = mapstructure.Decode(data, &responsePolicies)
	if err != nil {
		return nil, err
	}

	return responsePolicies.Results, nil
}

// GetSystem ...
func (c *iamBackendClient) GetSystem(systemID string) (system System, err error) {
	path := fmt.Sprintf("/api/v1/engine/systems/%s", systemID)
	query := map[string]interface{}{}
	data, err := c.callWithReturnMapData(GET, path, query, 10)
	if err != nil {
		return
	}

	err = mapstructure.Decode(data, &system)
	if err != nil {
		return
	}

	return
}

// CredentialsVerify ...
func (c *iamBackendClient) CredentialsVerify(appCode, appSecret string) (exists bool, err error) {
	path := "/api/v1/engine/credentials/verify"
	req := map[string]interface{}{
		"type": "app",
		"data": map[string]interface{}{
			"app_code":   appCode,
			"app_secret": appSecret,
		},
	}
	data, err := c.callWithReturnMapData(POST, path, req, 10)
	if err != nil {
		return
	}

	responseCredentialsVerify := CredentialsVerifyResponse{}
	err = mapstructure.Decode(data, &responseCredentialsVerify)
	if err != nil {
		return
	}

	return responseCredentialsVerify.Valid, nil
}
