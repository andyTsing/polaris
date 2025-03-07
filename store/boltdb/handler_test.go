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
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	v1 "github.com/polarismesh/polaris-server/common/api/v1"
	"github.com/polarismesh/polaris-server/common/model"
	"strconv"
	"testing"
	"time"
)

func TestBoltHandler_SaveNamespace(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	nsValue := &model.Namespace{
		Name:       "Test",
		Comment:    "test ns",
		Token:      "111111",
		Owner:      "user1",
		Valid:      true,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	err = handler.SaveValue(tblNameNamespace, nsValue.Name, nsValue)
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_LoadNamespace(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	nsValue := &model.Namespace{
		Name:       "Test",
		Comment:    "test ns",
		Token:      "111111",
		Owner:      "user1",
		Valid:      true,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	nsValues, err := handler.LoadValues(tblNameNamespace, []string{nsValue.Name}, &model.Namespace{})
	if nil != err {
		t.Fatal(err)
	}
	targetNsValue := nsValues[nsValue.Name]
	targetNs := targetNsValue.(*model.Namespace)
	fmt.Printf("loaded ns is %+v\n", targetNs)
	if nsValue.Name != targetNs.Name {
		fmt.Printf("name not equals\n")
	}
}

func TestBoltHandler_DeleteNamespace(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	nsValue := &model.Namespace{
		Name: "Test",
	}
	err = handler.DeleteValues(tblNameNamespace, []string{nsValue.Name})
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_Service(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	svcValue := &model.Service{
		ID:         "idSvc1234",
		Namespace:  "Test",
		Name:       "TestSvc",
		Comment:    "test svc",
		Token:      "111111",
		Owner:      "user1",
		Valid:      true,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
		Meta:       map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
	}
	err = handler.SaveValue("service", svcValue.ID, svcValue)
	if nil != err {
		t.Fatal(err)
	}
	nsValues, err := handler.LoadValues("service", []string{svcValue.ID}, &model.Service{})
	if nil != err {
		t.Fatal(err)
	}
	targetSvcValue := nsValues[svcValue.ID]
	targetSvc := targetSvcValue.(*model.Service)
	fmt.Printf("loaded svc is %+v\n", targetSvc)
	if svcValue.Name != targetSvc.Name || len(svcValue.Meta) != len(targetSvc.Meta) {
		fmt.Printf("name not equals\n")
	}
	fmt.Printf("trget meta is %v\n", targetSvc.Meta)

	_, _ = handler.LoadValuesByFilter("service", []string{"Meta"}, &model.Service{}, func(m map[string]interface{}) bool {
		values := m["Meta"]
		fmt.Printf("values are %v\n", values)
		return true
	})

	err = handler.DeleteValues("service", []string{svcValue.ID})
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_Location(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	id := "12345"
	locValue := &model.Location{
		Proto: &v1.Location{
			Region: &wrappers.StringValue{Value: "huabei"},
			Zone:   &wrappers.StringValue{Value: "shenzhen"},
			Campus: &wrappers.StringValue{Value: "longgang1"},
		},
		RegionID: 111,
		ZoneID:   1112,
		CampusID: 1113,
		Valid:    true,
	}
	err = handler.SaveValue("location", id, locValue)
	if nil != err {
		t.Fatal(err)
	}
	locValues, err := handler.LoadValues("location", []string{id}, &model.Location{})
	if nil != err {
		t.Fatal(err)
	}
	targetLocValue := locValues[id]
	targetLoc := targetLocValue.(*model.Location)
	fmt.Printf("loaded loc is %+v\n", targetLoc)
	err = handler.DeleteValues("location", []string{id})
	if nil != err {
		t.Fatal(err)
	}
}

const (
	tblService = "service"
)

func TestBoltHandler_CountValues(t *testing.T) {
	count := 5
	var idToServices = make(map[string]*model.Service)
	var ids = make([]string, 0)
	for i := 0; i < count; i++ {
		svcValue := &model.Service{
			ID:         "idSvcCount" + strconv.Itoa(i),
			Namespace:  "Test",
			Name:       "TestSvc" + strconv.Itoa(i),
			Comment:    "test svc",
			Token:      "111111",
			Owner:      "user1",
			Valid:      true,
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
			Meta:       map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
		}
		idToServices[svcValue.ID] = svcValue
		ids = append(ids, svcValue.ID)
	}
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	for id, svc := range idToServices {
		err = handler.SaveValue(tblService, id, svc)
		if nil != err {
			t.Fatal(err)
		}
	}
	services, err := handler.LoadValues(tblService, ids, &model.Service{})
	if nil != err {
		t.Fatal(err)
	}
	if len(services) != count {
		t.Fatal("load value count not match")
	}
	nCount, err := handler.CountValues(tblService)
	if nil != err {
		t.Fatal(err)
	}
	if nCount != count {
		t.Fatal("count not match")
	}
	err = handler.DeleteValues("service", ids)
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_LoadValuesByFilter(t *testing.T) {
	count := 5
	var idToServices = make(map[string]*model.Service)
	var ids = make([]string, 0)
	for i := 0; i < count; i++ {
		svcValue := &model.Service{
			ID:         "idSvcCount" + strconv.Itoa(i),
			Namespace:  "Test",
			Name:       "TestSvc" + strconv.Itoa(i),
			Comment:    "test svc",
			Token:      "111111",
			Owner:      "user" + strconv.Itoa(i),
			Valid:      true,
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
			Meta:       map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
		}
		idToServices[svcValue.ID] = svcValue
		ids = append(ids, svcValue.ID)
	}
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	for id, svc := range idToServices {
		err = handler.SaveValue(tblService, id, svc)
		if nil != err {
			t.Fatal(err)
		}
	}
	values, err := handler.LoadValuesByFilter(tblService, []string{"Owner"},
		&model.Service{}, func(props map[string]interface{}) bool {
			owner := props["Owner"].(string)
			return owner == "user1" || owner == "user2"
		})
	if nil != err {
		t.Fatal(err)
	}
	if len(values) != 2 {
		t.Fatal("filter count not match 2")
	}
	err = handler.DeleteValues("service", ids)
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_IterateFields(t *testing.T) {
	count := 5
	var idToServices = make(map[string]*model.Service)
	var ids = make([]string, 0)
	for i := 0; i < count; i++ {
		svcValue := &model.Service{
			ID:         "idSvcCount" + strconv.Itoa(i),
			Namespace:  "Test",
			Name:       "TestSvc" + strconv.Itoa(i),
			Comment:    "test svc",
			Token:      "111111",
			Owner:      "user" + strconv.Itoa(i),
			Valid:      true,
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
			Meta:       map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
		}
		idToServices[svcValue.ID] = svcValue
		ids = append(ids, svcValue.ID)
	}
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	for id, svc := range idToServices {
		err = handler.SaveValue(tblService, id, svc)
		if nil != err {
			t.Fatal(err)
		}
	}
	names := make([]string, 0)
	err = handler.IterateFields(tblService, "Name", &model.Service{}, func(value interface{}) {
		names = append(names, value.(string))
	})
	if nil != err {
		t.Fatal(err)
	}
	if len(names) != count {
		t.Fatalf("iterate count not match %d", count)
	}
	err = handler.DeleteValues("service", ids)
	if nil != err {
		t.Fatal(err)
	}
}

func TestBoltHandler_UpdateValue(t *testing.T) {
	count := 5
	var idToServices = make(map[string]*model.Service)
	var ids = make([]string, 0)
	for i := 0; i < count; i++ {
		svcValue := &model.Service{
			ID:         "idSvcCount" + strconv.Itoa(i),
			Namespace:  "Test",
			Name:       "TestSvc" + strconv.Itoa(i),
			Comment:    "test svc",
			Token:      "111111",
			Owner:      "user" + strconv.Itoa(i),
			Valid:      true,
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
			Meta:       map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
		}
		idToServices[svcValue.ID] = svcValue
		ids = append(ids, svcValue.ID)
	}
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	for id, svc := range idToServices {
		err = handler.SaveValue(tblService, id, svc)
		if nil != err {
			t.Fatal(err)
		}
	}
	targetId := ids[0]
	afterComment := "comment1"
	err = handler.UpdateValue(tblService, targetId, map[string]interface{}{
		"Comment": afterComment,
	})
	if nil != err {
		t.Fatal(err)
	}

	values, err := handler.LoadValues(tblService, []string{targetId}, &model.Service{})
	if nil != err {
		t.Fatal(err)
	}
	value, ok := values[targetId]
	if !ok {
		t.Fatalf("not exists %s", targetId)
	}

	if value.(*model.Service).Comment != afterComment {
		t.Fatalf("after comment not match")
	}

}
