/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package sqldb

import (
	"database/sql"
	"github.com/polarismesh/polaris-server/common/log"
	"github.com/polarismesh/polaris-server/store"
)

// query
type QueryHandler func(query string, args ...interface{}) (*sql.Rows, error)

// 批量查询数据的回调函数
type BatchHandler func(objects []interface{}) error

// 批量查询数据的对外接口
// 每次最多查询200个
func BatchQuery(label string, data []interface{}, handler BatchHandler) error {
	//start := time.Now()
	maxCount := 200
	beg := 0
	remain := len(data)
	if remain == 0 {
		return nil
	}

	progress := 0
	for {
		if remain > maxCount {
			if err := handler(data[beg : beg+maxCount]); err != nil {
				return err
			}

			beg += maxCount
			remain -= maxCount
			progress += maxCount
			if progress%20000 == 0 {
				log.Infof("[Store][database][Batch] query (%s) progress(%d / %d)", label, progress, len(data))
			}
		} else {
			if err := handler(data[beg : beg+remain]); err != nil {
				return err
			}
			break
		}
	}
	//log.Infof("[Store][database][Batch] consume time: %v", time.Now().Sub(start))
	return nil
}

/**
 * @brief 批量操作
 * @note 每次最多操作100个
 */
func BatchOperation(label string, data []interface{}, handler BatchHandler) error {
	if data == nil {
		return nil
	}
	maxCount := 100
	progress := 0
	for begin := 0; begin < len(data); begin += maxCount {
		end := begin + maxCount
		if end > len(data) {
			end = len(data)
		}
		if err := handler(data[begin:end]); err != nil {
			return err
		}
		progress += end - begin
		if progress%maxCount == 0 {
			log.Infof("[Store][database][Batch] operation (%s) progress(%d/%d)", label, progress, len(data))
		}
	}
	return nil
}

// 单独查询count个数的执行函数
func queryEntryCount(conn *BaseDB, str string, args []interface{}) (uint32, error) {
	var count uint32
	var err error
	Retry("queryRow", func() error {
		err = conn.QueryRow(str, args...).Scan(&count)
		return err
	})
	switch {
	case err == sql.ErrNoRows:
		log.Errorf("[Store][database] not found any entry(%s)", str)
		return 0, err
	case err != nil:
		log.Errorf("[Store][database] query entry count(%s) err: %s", str, err.Error())
		return 0, err
	default:
		return count, nil
	}
}

// 别名查询转换
var aliasFilter2Where = map[string]string{
	"service":   "source.name",
	"namespace": "source.namespace",
	"alias":     "alias.name",
	"owner":     "alias.owner",
}

// 别名查询字段转换函数
func serviceAliasFilter2Where(filter map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range filter {
		if d, ok := aliasFilter2Where[k]; ok {
			out[d] = v
		} else {
			out[k] = v
		}
	}

	return out
}

/**
 * @brief 检查数据库处理返回的行数
 */
func checkDataBaseAffectedRows(result sql.Result, counts ...int64) error {
	n, err := result.RowsAffected()
	if err != nil {
		log.Errorf("[Store][Database] get rows affected err: %s", err.Error())
		return err
	}

	for _, c := range counts {
		if n == c {
			return nil
		}
	}

	log.Errorf("[Store][Database] get rows affected result(%d) is not match expect(%+v)", n, counts)
	return store.NewStatusError(store.AffectedRowsNotMatch, "affected rows not matched")
}
