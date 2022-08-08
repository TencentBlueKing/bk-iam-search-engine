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

package eval

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	log "github.com/sirupsen/logrus"

	"engine/pkg/logging/debug"
	"engine/pkg/types"
	"engine/pkg/util"
)

type EvalEngine struct {
	engines       sync.Map
	lastIndexTime time.Time
}

func NewEvalEngine() (types.Engine, error) {
	return &EvalEngine{
		engines:       sync.Map{},
		lastIndexTime: time.Now(),
	}, nil
}

func (e *EvalEngine) genKey(system, action string) string {
	return system + ":" + action
}

func (e *EvalEngine) getActionEngine(system, action string) (engine *actionEvalEngine, ok bool) {
	key := e.genKey(system, action)
	var engineI interface{}
	engineI, ok = e.engines.Load(key)
	if ok {
		engine, ok = engineI.(*actionEvalEngine)
		if ok {
			return engine, ok
		}
	}
	return nil, ok
}

func (e *EvalEngine) createActionEngine(system, action string) (engine *actionEvalEngine) {
	engine = newActionEngine(system, action)
	key := e.genKey(system, action)
	e.engines.Store(key, engine)
	return
}

func (e *EvalEngine) getOrCreateActionEngine(system, action string) (engine *actionEvalEngine) {
	var ok bool
	engine, ok = e.getActionEngine(system, action)
	if ok {
		return engine
	}
	return e.createActionEngine(system, action)
}

// Size ...
func (e *EvalEngine) Size(system, action string) uint64 {
	engine, ok := e.getActionEngine(system, action)
	if ok {
		return engine.size()
	}
	return 0
}

// BulkAdd ...
func (e *EvalEngine) BulkAdd(policies []*types.Policy) (err error) {
	var engine *actionEvalEngine
	for _, p := range policies {
		// NOTE: here, should always be abac policies, len(p.Actions) == 1; the rbac policy should not show here
		if len(p.Actions) != 1 {
			log.Errorf(
				"eval engine got an invalid policy with p.Actions != 1, p=`%+v`, skip, should check the source data",
				p,
			)
			continue
		}
		engine = e.getOrCreateActionEngine(p.System, p.Actions[0].ID)
		engine.add(p)
	}
	e.lastIndexTime = time.Now()
	return
}

func (e *EvalEngine) engineRange(f func(engine *actionEvalEngine)) {
	// range local eval engine
	e.engines.Range(func(key, value interface{}) bool {
		engine := value.(*actionEvalEngine)
		f(engine)
		return true
	})
}

// Search ...
func (e *EvalEngine) Search(
	ctx context.Context,
	req *types.SearchRequest,
	entry *debug.Entry,
) (types.SearchResult, error) {
	logger := log.WithField(util.RequestIDKey, ctx.Value(util.ContextRequestIDKey))

	system := req.System
	action := req.Action.ID

	engine, ok := e.getActionEngine(system, action)
	if !ok {
		logger.Errorf("eval")
		return &EvalSearchResult{}, nil
	}

	subjects, err := engine.search(ctx, req, entry)
	return &EvalSearchResult{subjects: subjects}, err
}

// BatchSearch ...
func (e *EvalEngine) BatchSearch(
	ctx context.Context,
	requests []*types.SearchRequest,
	entry *debug.Entry,
) (results []types.SearchResult, err error) {
	for idx, req := range requests {
		var result types.SearchResult
		result, err = e.Search(ctx, req, debug.GetSubEntryByIndex(entry, idx))
		if err != nil {
			return
		}

		results = append(results, result)
	}
	return
}

// BulkDelete ...
func (e *EvalEngine) BulkDelete(ids []int64, logger *log.Entry) error {
	e.engineRange(func(engine *actionEvalEngine) {
		engine.bulkDelete(ids)
	})
	e.lastIndexTime = time.Now()
	return nil
}

// BulkDeleteBySubjects ...
func (e *EvalEngine) BulkDeleteBySubjects(beforeUpdatedAt int64, subjects []types.Subject, logger *log.Entry) error {
	e.engineRange(func(engine *actionEvalEngine) {
		engine.bulkDeleteBySubjects(beforeUpdatedAt, subjects)
	})
	e.lastIndexTime = time.Now()
	return nil
}

// Total ...
func (e *EvalEngine) Total() (size uint64) {
	e.engineRange(func(engine *actionEvalEngine) {
		size += engine.size()
	})
	return
}

// GetLastIndexTime ...
func (e *EvalEngine) GetLastIndexTime() time.Time {
	return e.lastIndexTime
}

// TakeSnapshot ...
func (e *EvalEngine) TakeSnapshot() []types.SnapRecord {
	data := make([]types.SnapRecord, 0, 10)
	e.engines.Range(func(key, value interface{}) bool {
		systemAction := key.(string)
		parts := strings.Split(systemAction, ":")
		if len(parts) != 2 {
			// TODO: wrong
			return true
		}
		system := parts[0]
		action := parts[1]

		engine := value.(*actionEvalEngine)

		data = append(data, types.SnapRecord{
			System:                system,
			Action:                action,
			LastModifiedTimestamp: engine.getLastIndexTime().Unix(),
			EvalPolicies:          engine.dump(),
		})
		return true
	})

	return data
}

// LoadSnapshot ...
func (e *EvalEngine) LoadSnapshot(data []types.SnapRecord) error {
	for _, record := range data {
		system := record.System
		action := record.Action

		engine := e.getOrCreateActionEngine(system, action)
		engine.bulkAdd(record.EvalPolicies)
		engine.setLastIndexTime(time.Unix(record.LastModifiedTimestamp, 0))
	}
	return nil
}

// EvalSearchResult ES 查询结果
type EvalSearchResult struct {
	subjects []types.Subject
}

// GetSubjects ...
func (e *EvalSearchResult) GetSubjects(allowedSubjectUIDs *set.StringSet) []types.Subject {
	subjects := make([]types.Subject, 0, len(e.subjects))
	for _, subject := range e.subjects {
		if allowedSubjectUIDs.Has(subject.UID) {
			continue
		}

		subjects = append(subjects, subject)
		allowedSubjectUIDs.Add(subject.UID)
	}
	return subjects
}

// actionEvalEngine ...
type actionEvalEngine struct {
	system string
	action string

	policies      map[int64]*types.Policy
	lastIndexTime time.Time

	mu *sync.RWMutex
}

// newActionEngine ...
func newActionEngine(system, action string) *actionEvalEngine {
	return &actionEvalEngine{
		system: system,
		action: action,

		policies:      make(map[int64]*types.Policy, 10),
		lastIndexTime: time.Now(),

		mu: new(sync.RWMutex),
	}
}

// empty ...
func (e *actionEvalEngine) empty() bool {
	return len(e.policies) == 0
}

// size ...
func (e *actionEvalEngine) size() uint64 {
	return uint64(len(e.policies))
}

// add ...
func (e *actionEvalEngine) add(p *types.Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.policies[p.ID] = p
	e.lastIndexTime = time.Now()
}

// bulkAdd ...
func (e *actionEvalEngine) bulkAdd(policies []*types.Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, p := range policies {
		e.policies[p.ID] = p
	}
	e.lastIndexTime = time.Now()
}

// search ...
func (e *actionEvalEngine) search(
	ctx context.Context,
	req *types.SearchRequest,
	entry *debug.Entry,
) ([]types.Subject, error) {
	if e.empty() {
		return nil, nil
	}

	// TODO: add cache here!!!!!!

	subjects := make([]types.Subject, 0, 100)

	// TODO: 从 sync.Pool 中初始化, 如果从sync.Pool中初始化, 已经扩容的slice会一直占内存吗
	evaledResults := make(map[string]bool, 100)

	// TODO: 从 sync.Pool 中初始化
	obj := expression.NewObjectSet()
	for _, resourceNode := range req.Resource {
		obj.Set(resourceNode.Type, resourceNode.Attribute)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	subjectUIDs := set.NewStringSet()
	for _, p := range e.policies {
		// 1. 如果已经过期了, 那么不计算
		if p.ExpiredAt < req.NowTimestamp {
			continue
		}

		// filter the subject not the req.SubjectType
		if req.SubjectType != types.SubjectTypeAll && req.SubjectType != p.Subject.Type {
			continue
		}

		// 2. 如果subject已经有了, 那么跳过
		// NOTE:这里是个线性执行的, 所以判断subject重复没有问题, 但是如果切换成并行计算, 那么会有问题
		if subjectUIDs.Has(p.Subject.UID) {
			continue
		}

		// 3. 已经执行过的表达式, 不再执行 (使用模板等配置出来的表达式可能是一样的)
		if hasPermission, ok := evaledResults[p.ExpressionSignature]; ok {
			if !hasPermission {
				// 无权限
				continue
			} else {
				// 有权限(此时subject肯定是没有出现过的), 加入到subjects
				subjects = append(subjects, p.Subject)
				subjectUIDs.Add(p.Subject.UID)

				continue
			}
		}

		// 4. eval, 判断有没有权限
		allowed := p.Expression.Eval(obj)
		evaledResults[p.ExpressionSignature] = allowed
		if allowed {
			subjects = append(subjects, p.Subject)
			subjectUIDs.Add(p.Subject.UID)

			debug.AddPolicy(entry, p)
		}
	}
	return subjects, nil
}

// bulkDelete ...
func (e *actionEvalEngine) bulkDelete(ids []int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, id := range ids {
		delete(e.policies, id)
	}

	e.lastIndexTime = time.Now()
}

func toSubjectSet(subjects []types.Subject) *set.StringSet {
	subjectSet := set.NewFixedLengthStringSet(len(subjects))
	for _, subject := range subjects {
		subjectSet.Add(subject.Type + ":" + subject.ID)
	}
	return subjectSet
}

// bulkDeleteBySubjects ...
func (e *actionEvalEngine) bulkDeleteBySubjects(beforeUpdatedAt int64, subjects []types.Subject) {
	subjectSet := toSubjectSet(subjects)
	e.bulkDeleteByMatchFunc(func(policy *types.Policy) bool {
		return subjectSet.Has(policy.Subject.UID) && policy.UpdatedAt < beforeUpdatedAt
	})
}

func (e *actionEvalEngine) bulkDeleteByMatchFunc(matchFunc func(policy *types.Policy) bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	deleteIDs := make([]int64, 0, 10)
	for _, p := range e.policies {
		if matchFunc(p) {
			deleteIDs = append(deleteIDs, p.ID)
		}
	}

	for _, id := range deleteIDs {
		delete(e.policies, id)
	}

	e.lastIndexTime = time.Now()
}

// dump ...
func (e *actionEvalEngine) dump() []*types.Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ps := make([]*types.Policy, 0, len(e.policies))

	for _, p := range e.policies {
		ps = append(ps, p)
	}
	return ps
}

// getLastIndexTime ...
func (e *actionEvalEngine) getLastIndexTime() time.Time {
	return e.lastIndexTime
}

// setLastIndexTime ...
func (e *actionEvalEngine) setLastIndexTime(lastIndexTime time.Time) {
	e.lastIndexTime = lastIndexTime
}
