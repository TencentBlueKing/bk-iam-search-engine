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

package types

import (
	"fmt"

	"github.com/TencentBlueKing/iam-go-sdk/expression"
	jsoniter "github.com/json-iterator/go"

	"engine/pkg/util"
)

// BkIAMPathSuffix ...
const (
	BkIAMPathSuffix = "._bk_iam_path_"
	BkIAMPathKey    = "_bk_iam_path_"
)

// ExpressionType ...
type ExpressionType string

// Any ...
const (
	Any  ExpressionType = "any"
	Doc  ExpressionType = "doc"
	Eval ExpressionType = "eval"
)

// Policy ...
type Policy struct {
	Version    string              `json:"version" mapstructure:"version"`
	ID         int64               `json:"id" mapstructure:"id"`
	System     string              `json:"system" mapstructure:"system"`
	Action     Action              `json:"action" mapstructure:"action"`
	Subject    Subject             `json:"subject" mapstructure:"subject"`
	TemplateID int64               `json:"template_id" mapstructure:"template_id"`
	Expression expression.ExprCell `json:"expression" mapstructure:"expression"`
	ExpiredAt  int64               `json:"expired_at" mapstructure:"expired_at"`
	UpdatedAt  int64               `json:"updated_at" mapstructure:"updated_at"`

	ExpressionSignature string
	ExpressionLength    int

	ExpressionType ExpressionType // 表达式类型 any, doc, eval
}

// FillUniqueFields ...
func (p *Policy) FillUniqueFields() error {
	p.Subject.FillUID()

	// NOTE: json bytes
	exprBytes, err := jsoniter.Marshal(p.Expression)
	if err != nil {
		return fmt.Errorf("policy: %d expression marshal error %w", p.ID, err)
	}
	p.ExpressionSignature = util.GetBytesMD5Hash(exprBytes)
	p.ExpressionLength = len(exprBytes)

	return nil
}

// Subject ...
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`

	UID string
}

// FillUID
func (s *Subject) FillUID() {
	s.UID = s.Type + ":" + s.ID
}

// ResourceCountReachLimit ...
func ResourceCountReachLimit(req *SearchRequest, allowedSubjectUIDs *util.StringSet) bool {
	return req.Limit > 0 && allowedSubjectUIDs.Size() >= req.Limit
}

// SnapRecord ...
type SnapRecord struct {
	System                string    `json:"system"`
	Action                string    `json:"action"`
	LastModifiedTimestamp int64     `json:"last_modified_timestamp"`
	EvalPolicies          []*Policy `json:"eval_policies"`
}

func (s *SnapRecord) FillPoliciesUniqueFields() (err error) {
	for _, p := range s.EvalPolicies {
		err = p.FillUniqueFields()
		if err != nil {
			return err
		}
	}
	return nil
}
