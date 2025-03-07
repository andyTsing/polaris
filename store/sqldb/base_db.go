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
	"context"
	"database/sql"
	"fmt"
	"github.com/polarismesh/polaris-server/common/log"
	"github.com/polarismesh/polaris-server/plugin"
	"strings"
	"time"
)

// db抛出的异常，需要重试的字符串组
var errMsg = []string{"Deadlock", "bad connection", "invalid connection"}

// 对sql.DB的封装
type BaseDB struct {
	*sql.DB
	cfg            *dbConfig
	isolationLevel sql.IsolationLevel
	parsePwd       plugin.ParsePassword
}

/**
 * @brief store的配置
 */
type dbConfig struct {
	dbType           string
	dbUser           string
	dbPwd            string
	dbAddr           string
	dbName           string
	maxOpenConns     int
	maxIdleConns     int
	connMaxLifetime  int
	txIsolationLevel int
}

// 新建一个BaseDB
func NewBaseDB(cfg *dbConfig, parsePwd plugin.ParsePassword) (*BaseDB, error) {
	baseDb := &BaseDB{cfg: cfg, parsePwd: parsePwd}
	if cfg.txIsolationLevel > 0 {
		baseDb.isolationLevel = sql.IsolationLevel(cfg.txIsolationLevel)
		log.Infof("[Store][database] use isolation level: %s", baseDb.isolationLevel.String())
	}

	if err := baseDb.openDatabase(); err != nil {
		return nil, err
	}

	return baseDb, nil
}

// 与数据库进行连接
func (b *BaseDB) openDatabase() error {
	c := b.cfg

	// 使用密码解析插件
	if b.parsePwd != nil {
		pwd, err := b.parsePwd.ParsePassword(c.dbPwd)
		if err != nil {
			log.Errorf("[Store][database][ParsePwdPlugin] parse password err: %s", err.Error())
			return err
		}
		c.dbPwd = pwd
	}

	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s", c.dbUser, c.dbPwd, c.dbAddr, c.dbName)
	db, err := sql.Open(c.dbType, dns)
	if err != nil {
		log.Errorf("[Store][database] sql open err: %s", err.Error())
		return err
	}
	if pingErr := db.Ping(); pingErr != nil {
		log.Errorf("[Store][database] database ping err: %s", pingErr.Error())
		return pingErr
	}
	if c.maxOpenConns > 0 {
		log.Infof("[Store][database] db set max open conns: %d", c.maxOpenConns)
		db.SetMaxOpenConns(c.maxOpenConns)
	}
	if c.maxIdleConns > 0 {
		log.Infof("[Store][database] db set max idle conns: %d", c.maxIdleConns)
		db.SetMaxIdleConns(c.maxIdleConns)
	}
	if c.connMaxLifetime > 0 {
		log.Infof("[Store][database] db set conn max life time: %d", c.connMaxLifetime)
		db.SetConnMaxLifetime(time.Second * time.Duration(c.connMaxLifetime))
	}

	b.DB = db
	return nil
}

// 重写db.Exec函数
// 提供重试功能
func (b *BaseDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error
	Retry("exec "+query, func() error {
		result, err = b.DB.Exec(query, args...)
		return err
	})

	return result, err
}

// 重写db.Query函数
func (b *BaseDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error
	Retry("query "+query, func() error {
		rows, err = b.DB.Query(query, args...)
		return err
	})

	return rows, err
}

// 重写db.Begin
func (b *BaseDB) Begin() (*BaseTx, error) {
	var tx *sql.Tx
	var err error
	var option *sql.TxOptions
	if b.isolationLevel > 0 {
		option = &sql.TxOptions{Isolation: sql.IsolationLevel(b.isolationLevel)}
	}
	Retry("begin", func() error {
		tx, err = b.DB.BeginTx(context.Background(), option)
		return err
	})

	return &BaseTx{Tx: tx}, err
}

// 对sql.Tx的封装
type BaseTx struct {
	*sql.Tx
}

// 重试主函数
// 最多重试20次，每次等待5ms*重试次数
func Retry(label string, handle func() error) {
	var err error
	maxTryTimes := 20
	for i := 1; i <= maxTryTimes; i++ {
		err = handle()
		if err == nil {
			return
		}

		repeated := false // 是否重试
		for _, msg := range errMsg {
			if strings.Contains(err.Error(), msg) {
				log.Warnf("[Store][database][%s] get error msg: %s. Repeated doing(%d)", label, err.Error(), i)
				time.Sleep(time.Millisecond * 5 * time.Duration(i))
				repeated = true
				break
			}
		}
		if !repeated {
			return
		}
	}
}

// 事务重试
func RetryTransaction(label string, handle func() error) error {
	var err error
	Retry(label, func() error {
		err = handle()
		return err
	})
	return err
}
