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

package boltdb

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/polarismesh/polaris-server/common/log"
)

// BoltHandler encapsulate operations around boltdb
type BoltHandler interface {

	// SaveValue insert data object, each data object should be identified by unique key
	SaveValue(typ string, key string, object interface{}) error

	// DeleteValue delete data object by unique key
	DeleteValues(typ string, key []string) error

	// UpdateValue update properties of data object
	UpdateValue(typ string, key string, properties map[string]interface{}) error

	// LoadValues load data objects by unique keys, return value is 'key->object' map
	LoadValues(typ string, keys []string, typObject interface{}) (map[string]interface{}, error)

	// LoadValuesByFilter filter data objects by condition, return value is 'key->object' map
	LoadValuesByFilter(typ string, fields []string,
		typObject interface{}, filter func(map[string]interface{}) bool) (map[string]interface{}, error)

	// LoadValues load all saved data objects, return value is 'key->object' map
	LoadValuesAll(typ string, typObject interface{}) (map[string]interface{}, error)

	// IterateFields iterate all saved data objects
	IterateFields(typ string, field string, typObject interface{}, process func(interface{})) error

	// CountValues count all data objects
	CountValues(typ string) (int, error)

	// Execute execute scripts directly
	Execute(writable bool, process func(tx *bolt.Tx) error) error

	// BeginTransaction begin boltdb transaction
	Transaction() (*bolt.Tx, error)

	// Close close boltdb
	Close() error
}

// BoltConfig config to initialize boltdb
type BoltConfig struct {
	// FileName boltdb store file
	FileName string
}

const (
	confPath    = "path"
	defaultPath = "./polaris.bolt"
)

// Parse parse yaml config
func (c *BoltConfig) Parse(opt map[string]interface{}) {
	if value, ok := opt[confPath]; ok {
		c.FileName = value.(string)
	} else {
		c.FileName = defaultPath
	}
}

const (
	defaultTimeoutForFileLock = 5 * time.Second
)

// NewBoltHandler create the boltdb handler
func NewBoltHandler(config *BoltConfig) (BoltHandler, error) {
	db, err := openBoltDB(config.FileName)
	if nil != err {
		return nil, err
	}
	return &boltHandler{db: db}, nil
}

type boltHandler struct {
	db *bolt.DB
}

func openBoltDB(path string) (*bolt.DB, error) {
	return bolt.Open(path, 0600, &bolt.Options{
		Timeout: defaultTimeoutForFileLock,
	})
}

// SaveValue insert data object, each data object should be identified by unique key
func (b *boltHandler) SaveValue(typ string, key string, value interface{}) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		var typBucket *bolt.Bucket
		var err error
		typBucket, err = tx.CreateBucketIfNotExists([]byte(typ))
		if nil != err {
			return err
		}
		keyBuf := []byte(key)
		var bucket *bolt.Bucket
		//先清理老数据
		bucket = typBucket.Bucket(keyBuf)
		if nil != bucket {
			if err = typBucket.DeleteBucket(keyBuf); nil != err {
				return err
			}
		}
		//创建全新bucket
		bucket, err = typBucket.CreateBucket(keyBuf)
		if nil != err {
			return err
		}
		var buffers map[string][]byte
		buffers, err = serializeObject(bucket, value)
		if nil != err {
			return err
		}
		if len(buffers) > 0 {
			for k, v := range buffers {
				err = bucket.Put([]byte(k), v)
				if nil != err {
					return err
				}
			}
		}
		return err
	})
}

// LoadValues load data objects by unique keys, return value is 'key->object' map
func (b *boltHandler) LoadValues(typ string, keys []string, typObject interface{}) (map[string]interface{}, error) {
	var values = make(map[string]interface{})
	if len(keys) == 0 {
		return values, nil
	}
	err := b.db.View(func(tx *bolt.Tx) error {
		return loadValues(tx, typ, keys, typObject, values)
	})
	return values, err
}

func loadValues(tx *bolt.Tx, typ string, keys []string, typObject interface{}, values map[string]interface{}) error {
	for _, key := range keys {
		bucket := getBucket(tx, typ, key)
		if nil == bucket {
			continue
		}
		toObj, err := deserializeObject(bucket, typObject)
		if nil != err {
			return err
		}
		values[key] = toObj
	}
	return nil
}

// LoadValuesByFilter filter data objects by condition, return value is 'key->object' map
func (b *boltHandler) LoadValuesByFilter(typ string, fields []string,
	typObject interface{}, filter func(map[string]interface{}) bool) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := b.db.View(func(tx *bolt.Tx) error {
		return loadValuesByFilter(tx, typ, fields, typObject, filter, values)
	})
	return values, err
}

func loadValuesByFilter(tx *bolt.Tx, typ string, fields []string, typObject interface{},
	filter func(map[string]interface{}) bool, values map[string]interface{}) error {
	typeBucket := tx.Bucket([]byte(typ))
	if nil == typeBucket {
		return nil
	}
	keys, err := getKeys(typeBucket)
	if nil != err {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		bucket := typeBucket.Bucket([]byte(key))
		if nil == bucket {
			log.Warnf("[BlobStore] bucket not found for key %s, type %s", key, typ)
			continue
		}
		var matchResult bool
		matchResult, err = matchObject(bucket, fields, typObject, filter)
		if nil != err {
			return err
		}
		if !matchResult {
			continue
		}
		var targetObj interface{}
		targetObj, err = deserializeObject(bucket, typObject)
		if nil != err {
			return err
		}
		values[key] = targetObj
	}
	return nil
}

func reflectProtoMsg(typObject interface{}, fieldName string) (proto.Message, error) {
	intoType := indirectType(reflect.TypeOf(typObject))
	field, ok := intoType.FieldByName(fieldName)
	if !ok {
		return nil, errors.New(fmt.Sprintf("field %s not found in object %v", fieldName, intoType))
	}
	rawFieldType := field.Type
	if !rawFieldType.Implements(messageType) {
		return nil, errors.New(fmt.Sprintf("field %s type not match in object %v, want %v, get %v",
			fieldName, intoType, messageType, field.Type))
	}
	return reflect.New(rawFieldType.Elem()).Interface().(proto.Message), nil
}

func reflectMapMsg(bucket *bolt.Bucket, bucketField string) (map[string]string, error) {
	subBucket := bucket.Bucket([]byte(bucketField))
	if nil == subBucket {
		return nil, nil
	}
	values := make(map[string]string)
	err := subBucket.ForEach(func(k, v []byte) error {
		values[string(k)] = string(v)
		return nil
	})
	if nil != err {
		return nil, err
	}
	return values, nil
}

func getFieldObject(bucket *bolt.Bucket, typObject interface{}, field string) (interface{}, error) {
	bucketField := toBucketField(field)
	valueBytes := bucket.Get([]byte(bucketField))
	if len(valueBytes) == 0 {
		return reflectMapMsg(bucket, bucketField)
	}
	typByte := valueBytes[0]
	switch typByte {
	case typeString:
		value, _ := decodeStringBuffer(bucketField, valueBytes)
		return value, nil
	case typeBool:
		value, _ := decodeBoolBuffer(bucketField, valueBytes)
		return value, nil
	case typeTime:
		value, _ := decodeTimeBuffer(bucketField, valueBytes)
		return value, nil
	case typeProtobuf:
		msg, err := reflectProtoMsg(typObject, field)
		if nil != err {
			return false, err
		}
		value, err := decodeMessageBuffer(msg, field, valueBytes)
		if nil != err {
			return false, err
		}
		return value, nil
	case typeInt, typeInt8, typeInt16, typeInt32, typeInt64:
		value, _ := decodeIntBuffer(field, valueBytes, typByte)
		return value, nil
	case typeUint, typeUint8, typeUint16, typeUint32, typeUint64:
		value, _ := decodeUintBuffer(field, valueBytes, typByte)
		return value, nil
	default:
		log.Warnf(
			"[BlobStore] matchObject unrecognized field %s, type is %d", field, typByte)
		return nil, nil
	}
}

func matchObject(bucket *bolt.Bucket,
	fields []string, typObject interface{}, filter func(map[string]interface{}) bool) (bool, error) {
	if len(fields) == 0 {
		return true, nil
	}
	if nil == filter {
		return true, nil
	}
	fieldValues := make(map[string]interface{}, 0)
	for _, field := range fields {
		value, err := getFieldObject(bucket, typObject, field)
		if nil != err {
			return false, err
		}
		if nil == value {
			continue
		}
		fieldValues[field] = value
	}
	return filter(fieldValues), nil
}

// IterateFields iterate all saved data objects
func (b *boltHandler) IterateFields(typ string, field string, typObject interface{}, filter func(interface{})) error {
	if nil == filter {
		return nil
	}
	return b.db.View(func(tx *bolt.Tx) error {
		typeBucket := tx.Bucket([]byte(typ))
		if nil == typeBucket {
			return nil
		}
		keys, err := getKeys(typeBucket)
		if nil != err {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		for _, key := range keys {
			bucket := typeBucket.Bucket([]byte(key))
			if nil == bucket {
				log.Warnf("[BlobStore] bucket not found for key %s, type %s", key, typ)
				continue
			}
			var fieldObj interface{}
			fieldObj, err = getFieldObject(bucket, typObject, field)
			if nil != err {
				return err
			}
			filter(fieldObj)
		}
		return nil
	})
}

// Close close boltdb
func (b *boltHandler) Close() error {
	if nil != b.db {
		return b.db.Close()
	}
	return nil
}

// DeleteValue delete data object by unique key
func (b *boltHandler) DeleteValues(typ string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		return deleteValues(tx, typ, keys)
	})
}

func deleteValues(tx *bolt.Tx, typ string, keys []string) error {
	typeBucket := tx.Bucket([]byte(typ))
	if nil == typeBucket {
		return nil
	}
	for _, key := range keys {
		keyBytes := []byte(key)
		if nil != typeBucket.Bucket(keyBytes) {
			err := typeBucket.DeleteBucket(keyBytes)
			if nil != err {
				return err
			}
		}
	}
	return nil
}

func getBucket(tx *bolt.Tx, typ string, key string) *bolt.Bucket {
	bucket := tx.Bucket([]byte(typ))
	if nil == bucket {
		return nil
	}
	return bucket.Bucket([]byte(key))
}

func convertInt64Value(value interface{}, kind reflect.Kind) int64 {
	switch kind {
	case reflect.Int:
		return int64(value.(int))
	case reflect.Int8:
		return int64(value.(int8))
	case reflect.Int16:
		return int64(value.(int16))
	case reflect.Int32:
		return int64(value.(int32))
	case reflect.Int64:
		return value.(int64)
	}
	return 0
}

func convertUint64Value(value interface{}, kind reflect.Kind) uint64 {
	switch kind {
	case reflect.Uint:
		return uint64(value.(uint))
	case reflect.Uint8:
		return uint64(value.(uint8))
	case reflect.Uint16:
		return uint64(value.(uint16))
	case reflect.Uint32:
		return uint64(value.(uint32))
	case reflect.Uint64:
		return value.(uint64)
	}
	return 0
}

func getKeys(bucket *bolt.Bucket) ([]string, error) {
	keys := make([]string, 0)
	err := bucket.ForEach(func(k, v []byte) error {
		keys = append(keys, string(k))
		return nil
	})
	return keys, err
}

// CountValues count all data objects
func (b *boltHandler) CountValues(typ string) (int, error) {
	var count int
	err := b.db.View(func(tx *bolt.Tx) error {
		typeBucket := tx.Bucket([]byte(typ))
		if nil == typeBucket {
			return nil
		}
		return typeBucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
	})
	return count, err
}

// UpdateValue update properties of data object
func (b *boltHandler) UpdateValue(typ string, key string, properties map[string]interface{}) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		var err error
		typeBucket := tx.Bucket([]byte(typ))
		if nil == typeBucket {
			return nil
		}
		bucket := typeBucket.Bucket([]byte(key))
		if nil == bucket {
			return nil
		}
		if len(properties) == 0 {
			return nil
		}
		for propKey, propValue := range properties {
			bucketKey := toBucketField(propKey)
			propType := reflect.TypeOf(propValue)
			kind := propType.Kind()
			switch kind {
			case reflect.String:
				err = bucket.Put([]byte(bucketKey), encodeStringBuffer(propValue.(string)))
			case reflect.Bool:
				err = bucket.Put([]byte(bucketKey), encodeBoolBuffer(propValue.(bool)))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				err = bucket.Put([]byte(bucketKey),
					encodeIntBuffer(convertInt64Value(propValue, kind), numberKindToType[kind]))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				err = bucket.Put([]byte(bucketKey),
					encodeUintBuffer(convertUint64Value(propValue, kind), numberKindToType[kind]))
			case reflect.Map:
				err = encodeRawMap(bucket, bucketKey, propValue.(map[string]string))
			case reflect.Ptr:
				if propType.Implements(messageType) {
					//protobuf类型
					var msgBuf []byte
					msgBuf, err = encodeMessageBuffer(propValue.(proto.Message))
					if nil != err {
						return err
					}
					err = bucket.Put([]byte(bucketKey), msgBuf)
				}
			case reflect.Struct:
				if propType.AssignableTo(timeType) {
					//时间类型
					err = bucket.Put([]byte(bucketKey), encodeTimeBuffer(propValue.(time.Time)))
				}
			}
			if nil != err {
				return err
			}
		}
		return nil
	})
}

// LoadValues load all saved data objects, return value is 'key->object' map
func (b *boltHandler) LoadValuesAll(typ string, typObject interface{}) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := b.db.View(func(tx *bolt.Tx) error {
		typeBucket := tx.Bucket([]byte(typ))
		if nil == typeBucket {
			return nil
		}
		keys, err := getKeys(typeBucket)
		if nil != err {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		for _, key := range keys {
			bucket := typeBucket.Bucket([]byte(key))
			if nil == bucket {
				log.Warnf("[BlobStore] bucket not found for key %s, type %s", key, typ)
				continue
			}
			var targetObj interface{}
			targetObj, err = deserializeObject(bucket, typObject)
			if nil != err {
				return err
			}
			values[key] = targetObj
		}
		return nil
	})
	return values, err
}

// Execute execute scripts directly
func (b *boltHandler) Execute(writable bool, process func(tx *bolt.Tx) error) error {
	if writable {
		return b.db.Update(process)
	}
	return b.db.View(process)
}

// BeginTransaction begin boltdb transaction
func (b *boltHandler) Transaction() (*bolt.Tx, error) {
	return b.db.Begin(true)
}
