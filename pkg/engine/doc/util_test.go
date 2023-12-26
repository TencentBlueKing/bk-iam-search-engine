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

package doc

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Doc", func() {
	Describe("splitBKIAMPath", func() {
		It("empty", func() {
			paths := splitBKIAMPath("")
			assert.Empty(GinkgoT(), paths)
		})

		It("one level", func() {
			paths := splitBKIAMPath("/biz,1/")
			assert.Len(GinkgoT(), paths, 1)
			assert.Contains(GinkgoT(), paths, "/biz,1/")
		})

		It("two level", func() {
			paths := splitBKIAMPath("/biz,1/set,2/")
			assert.Len(GinkgoT(), paths, 3)
			assert.Contains(GinkgoT(), paths, "/biz,1/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,2/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,*/")
		})

		It("three level", func() {
			paths := splitBKIAMPath("/biz,1/set,2/module,3/")
			assert.Len(GinkgoT(), paths, 5)
			assert.Contains(GinkgoT(), paths, "/biz,1/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,2/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,2/module,3/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,*/")
			assert.Contains(GinkgoT(), paths, "/biz,1/set,2/module,*/")
		})
	})
})

func BenchmarkSplitBKIAMPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		splitBKIAMPath("/biz,1/set,2/module,3/")
	}
}
