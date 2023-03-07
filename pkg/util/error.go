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

package util

import (
	"time"

	"github.com/getsentry/sentry-go"

	"engine/pkg/errorx"
)

// Error Codes
const (
	NoError           = 0
	ParamError        = 1903002
	BadRequestError   = 1903400
	UnauthorizedError = 1903401
	ForbiddenError    = 1903403
	NotFoundError     = 1903404
	ConflictError     = 1903409
	SystemError       = 1903500
)

// ReportToSentry is a shortcut to build and send an event to sentry
func ReportToSentry(message string, extra map[string]interface{}) {
	// report to sentry
	ev := sentry.NewEvent()
	ev.Message = message
	ev.Level = "error"
	ev.Timestamp = time.Now()
	ev.Extra = extra
	errorx.ReportEvent(ev)
}
