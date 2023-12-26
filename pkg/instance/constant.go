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

package instance

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// NOTE: we want to use the same config file, but different instance, so here we use env

const (
	instanceTypeRbac = "rbac"

	policyAPITypeAbac = "abac"
	policyAPITypeRbac = "rbac"

	policyRbacBeginID = 500000000
)

var (
	// api param type={engineAPIType}
	policyAPIType = policyAPITypeAbac

	// abac, begin 1
	// rbac, begin 500000000
	policyBeginID = 1

	// for sync local file
	fullSyncFileName = "last_sync_time.full"
	incrSyncFileName = "last_sync_time.incr"
	snapshotFileName = "snapshot.json"
)

func init() {
	instanceType := os.Getenv("INSTANCE_TYPE")

	if instanceType == instanceTypeRbac {
		policyAPIType = policyAPITypeRbac
		policyBeginID = policyRbacBeginID
		fullSyncFileName = "last_sync_time.rbac.full"
		incrSyncFileName = "last_sync_time.rbac.incr"
		snapshotFileName = "snapshot.rbac.json"
	}
	log.Infof("init Component with policyAPIType=%s", policyAPIType)
}

func GetPolicyAPIType() string {
	return policyAPIType
}

func GetPolicyBeginID() (beginID int64) {
	return int64(policyBeginID)
}

func GetFullSyncFileName() string {
	return fullSyncFileName
}

func GetIncrSyncFileName() string {
	return incrSyncFileName
}

func GetSnapshotFileName() string {
	return snapshotFileName
}
