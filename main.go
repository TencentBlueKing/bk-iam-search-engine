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

package main

import "engine/cmd"

// @title IAM-Search-Engine API
// @version 1.0
// @description 蓝鲸权限中心后台 engine 服务 API 文档

// NOTE:
//   1. build a search engine (bleve)
//   2. parse then make the expression into 3 parts
//      2.1 any (is_any check)
//      2.2 id in  / id eq
//      2.3 can't resolve, eval only
//   3. B1: merge all_OR policies, move some policies from 2.3 to 2.2
//   4. B2: if 2.1/2.2 subject has permission, no eval in 2.3
//   5. B3: expression has the same signature, if 2.3 has evaled policy, no eval again
//   6. B4: single AND policy with remote_resource_dependency(bk_job to bk_cmdb)  => 不处理
//   8. B6: turn off the analysis of each term, we don't need it

// TODO:
//   7. B5: the order of complex expression(sort by subject->score)
//   9. B7: merge any, remove other exprs in all OR

// NOTE: 数据拉取和同步, 会成为这个项目最大的问题

// TODO: 数据同步
//     1. 有一个syncer, 定期同步每个 system+action 的策略列表
//     2. 有一个watcher, 订阅变更事件, 增删改对应的策略, 同步
//     3. 同步的方向:
//        3.1 es
//        3.2 ToEval()
//        困难点: 这二者是可能相互转换的

// TODO: toEval的策略集合维护困难  (最大的问题)
//     1. 每次拉起后, 构建? (此时需要查询, 注意判定, 放到内存)
//     2. 一旦程序崩溃, 重启后, 需要重新构建
//     3. 重建到服务可用, 之间可能很长一段时间接口不可用

func main() {

	cmd.Execute()
	// router := server.NewRouter(c)
	//
	// // By default it serves on :8080 unless a
	// // PORT environment variable was defined.
	// _ = router.Run("0.0.0.0:9202")
}
