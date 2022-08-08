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

package storage

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/natefinch/atomic"

	"engine/pkg/instance"
	"engine/pkg/util"
)

// Storage ...
type Storage struct {
	dir    string
	fullMu sync.RWMutex
	incrMu sync.RWMutex
	snapMu sync.RWMutex
}

// NewStorage ...
func NewStorage(dir string) *Storage {
	return &Storage{
		dir: dir,
	}
}

// GetFullSyncLastSyncTime ...
func (s *Storage) GetFullSyncLastSyncTime() (ts int64, err error) {
	s.fullMu.RLock()
	ts, err = s.getLastSyncTime(instance.GetFullSyncFileName())
	s.fullMu.RUnlock()
	return
}

// SetFullSyncLastSyncTime ...
func (s *Storage) SetFullSyncLastSyncTime(lastSyncTs int64) (err error) {
	s.fullMu.Lock()
	err = s.setLastSyncTime(lastSyncTs, instance.GetFullSyncFileName())
	s.fullMu.Unlock()
	return err
}

// GetIncrSyncLastSyncTime ...
func (s *Storage) GetIncrSyncLastSyncTime() (ts int64, err error) {
	s.incrMu.RLock()
	ts, err = s.getLastSyncTime(instance.GetIncrSyncFileName())
	s.incrMu.RUnlock()
	return
}

// SetIncrSyncLastSyncTime ...
func (s *Storage) SetIncrSyncLastSyncTime(lastSyncTs int64) (err error) {
	s.incrMu.Lock()
	err = s.setLastSyncTime(lastSyncTs, instance.GetIncrSyncFileName())
	s.incrMu.Unlock()
	return err
}

// SaveSnapshot ...
func (s *Storage) SaveSnapshot(data []byte) error {
	s.snapMu.Lock()
	path := filepath.Join(s.dir, instance.GetSnapshotFileName())

	// f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// _, err = f.Write(data)

	err := atomic.WriteFile(path, bytes.NewReader(data))

	s.snapMu.Unlock()
	return err
}

// GetSnapshot ...
func (s *Storage) GetSnapshot() ([]byte, error) {
	s.snapMu.RLock()
	path := filepath.Join(s.dir, instance.GetSnapshotFileName())

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		// if file not exists, init system version = 0
		if os.IsNotExist(err) {
			return nil, ErrNoSyncBefore
		}
		return nil, err
	}

	s.snapMu.RUnlock()
	return bs, nil
}

// ExistSnapshot ...
func (s *Storage) ExistSnapshot() bool {
	path := filepath.Join(s.dir, instance.GetSnapshotFileName())

	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

// ErrNoSyncBefore ...
var ErrNoSyncBefore = errors.New("no sync before")

func (s *Storage) getLastSyncTime(fileName string) (int64, error) {
	path := filepath.Join(s.dir, fileName)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		// if file not exists, init system version = 0
		if os.IsNotExist(err) {
			return -1, ErrNoSyncBefore
		}
		return -1, err
	}

	ts := util.BytesToString(b)
	return strconv.ParseInt(ts, 10, 64)
}

func (s *Storage) setLastSyncTime(lastSyncTs int64, fileName string) (err error) {
	path := filepath.Join(s.dir, fileName)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	versionStr := util.Int64ToString(lastSyncTs)
	_, err = f.Write([]byte(versionStr))
	return
}
