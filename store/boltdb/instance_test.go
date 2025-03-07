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
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	api "github.com/polarismesh/polaris-server/common/api/v1"
	"github.com/polarismesh/polaris-server/common/model"
)

const (
	insCount = 5
)

func TestInstanceStore_AddInstance(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}
	for i := 0; i < insCount; i++ {

		nowt := time.Now().Format("2006-01-02 15:04:05")

		err = insStore.AddInstance(&model.Instance{
			Proto: &api.Instance{
				Id:                &wrappers.StringValue{Value: "insid" + strconv.Itoa(i)},
				Host:              &wrappers.StringValue{Value: "1.1.1." + strconv.Itoa(i)},
				Port:              &wrappers.UInt32Value{Value: uint32(i + 1)},
				Protocol:          &wrappers.StringValue{Value: "grpc"},
				Weight:            &wrappers.UInt32Value{Value: uint32(i + 1)},
				EnableHealthCheck: &wrappers.BoolValue{Value: true},
				Healthy:           &wrappers.BoolValue{Value: true},
				Isolate:           &wrappers.BoolValue{Value: true},
				Metadata: map[string]string{
					"insk1": "insv1",
					"insk2": "insv2",
				},
				Ctime:    &wrappers.StringValue{Value: nowt},
				Mtime:    &wrappers.StringValue{Value: nowt},
				Revision: &wrappers.StringValue{Value: "revision" + strconv.Itoa(i)},
			},
			ServiceID:         "svcid1",
			ServicePlatformID: "svcPlatId1",
			Valid:             true,
			ModifyTime:        time.Now(),
		})
		if nil != err {
			t.Fatal(err)
		}
	}
}

func TestInstanceStore_BatchAddInstances(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	instances := make([]*model.Instance, 0)
	for i := insCount; i < insCount+5; i++ {

		nowt := time.Now().Format("2006-01-02 15:04:05")

		ins := &model.Instance{
			Proto: &api.Instance{
				Id:                &wrappers.StringValue{Value: "insid" + strconv.Itoa(i)},
				Host:              &wrappers.StringValue{Value: "1.1.1." + strconv.Itoa(i)},
				Port:              &wrappers.UInt32Value{Value: uint32(i)},
				Protocol:          &wrappers.StringValue{Value: "grpc"},
				Weight:            &wrappers.UInt32Value{Value: uint32(i)},
				EnableHealthCheck: &wrappers.BoolValue{Value: true},
				Healthy:           &wrappers.BoolValue{Value: true},
				Isolate:           &wrappers.BoolValue{Value: true},
				Metadata: map[string]string{
					"insk1": "insv1",
					"insk2": "insv2",
				},
				Ctime:    &wrappers.StringValue{Value: nowt},
				Mtime:    &wrappers.StringValue{Value: nowt},
				Revision: &wrappers.StringValue{Value: "revision" + strconv.Itoa(i)},
			},
			ServiceID:         "svcid2",
			ServicePlatformID: "svcPlatId1",
			Valid:             true,
			ModifyTime:        time.Now(),
		}

		instances = append(instances, ins)
	}

	err = insStore.BatchAddInstances(instances)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstanceStore_GetExpandInstances(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	_, ii, err := insStore.GetExpandInstances(nil, nil, 0, 20)
	if nil != err {
		t.Fatal(err)
	}

	for _, i := range ii {
		fmt.Printf("instances is %+v\n", i)
	}
}

func TestInstanceStore_GetMoreInstances(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	tt, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 15:04:05")
	m, err := insStore.GetMoreInstances(tt, false, false, []string{"svcid2"})
	if err != nil {
		t.Fatal(err)
	}

	if len(m) != 5 {
		t.Fatal(fmt.Sprintf("get more instances error, except len is %d, got %d", 5, len(m)))
	}

}

func TestInstanceStore_SetInstanceHealthStatus(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	err = insStore.SetInstanceHealthStatus("insid7", 0, "rev-no-healthy")
	if err != nil {
		t.Fatal("set instance healthy error")
	}

	ins, err := insStore.GetInstance("insid7")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Proto.GetHealthy().GetValue() != false {
		t.Fatal(fmt.Sprintf("set instance healthy status error, except %t, got %t",
			false, ins.Proto.GetHealthy().GetValue()))
	}
}

func TestInstanceStore_BatchSetInstanceIsolate(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	err = insStore.BatchSetInstanceIsolate([]interface{}{"insid7", "insid1"}, 0, "rev-no-Isolate")
	if err != nil {
		t.Fatal("set instance isolate error")
	}

	ins, err := insStore.GetInstance("insid7")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Proto.GetIsolate().GetValue() != false {
		t.Fatal(fmt.Sprintf("set instance isolate  error, except %t, got %t",
			false, ins.Proto.GetIsolate().GetValue()))
	}

	ins, err = insStore.GetInstance("insid1")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Proto.GetIsolate().GetValue() != false {
		t.Fatal(fmt.Sprintf("set instance isolate  error, except %t, got %t",
			false, ins.Proto.GetIsolate().GetValue()))
	}
}

func TestInstanceStore_GetInstancesMainByService(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	ii, err := insStore.GetInstancesMainByService("svcid1", "1.1.1.1")
	if nil != err {
		t.Fatal(err)
	}

	for _, i := range ii {
		fmt.Printf("get instance %+v\n", i)
	}
}

func TestInstanceStore_UpdateInstance(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	insM := &model.Instance{
		Proto: &api.Instance{
			Id:                &wrappers.StringValue{Value: "insid" + strconv.Itoa(0)},
			Service:           &wrappers.StringValue{Value: "svcid1"},
			Namespace:         &wrappers.StringValue{Value: "testns"},
			Host:              &wrappers.StringValue{Value: "1.2.3." + strconv.Itoa(0)},
			Port:              &wrappers.UInt32Value{Value: uint32(8080)},
			Protocol:          &wrappers.StringValue{Value: "trpc"},
			Weight:            &wrappers.UInt32Value{Value: uint32(0)},
			EnableHealthCheck: &wrappers.BoolValue{Value: true},
			Healthy:           &wrappers.BoolValue{Value: true},
			Isolate:           &wrappers.BoolValue{Value: true},
			Metadata: map[string]string{
				"modifyK1": "modifyV1",
				"modifyK2": "modifyV1",
			},
		},
		ServiceID:         "svcid1",
		ServicePlatformID: "svcPlatId1",
		Valid:             true,
		ModifyTime:        time.Now(),
	}

	err = insStore.UpdateInstance(insM)
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	ins, err := insStore.GetInstance("insid0")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Proto.GetHost().GetValue() != "1.2.3.0" ||
		ins.Proto.GetPort().GetValue() != 8080 {
		t.Fatal(fmt.Sprintf("udpate instance error"))
	}
}

func TestInstanceStore_GetInstancesBrief(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}
	sStore := &serviceStore{handler: handler}

	err = sStore.AddService(&model.Service{
		ID:        "svcid1",
		Name:      "svcname1",
		Namespace: "testsvc",
		Business:  "testbuss",
		Ports:     "8080",
		Meta: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
		Comment:    "testcomment",
		Department: "testdepart",
		Token:      "testtoken1",
		Owner:      "testowner",
		Revision:   "testrevision1",
		Reference:  "",
		Valid:      true,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = sStore.AddService(&model.Service{
		ID:        "svcid2",
		Name:      "svcname2",
		Namespace: "testsvc",
		Business:  "testbuss",
		Ports:     "8080",
		Meta: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
		Comment:    "testcomment",
		Department: "testdepart",
		Token:      "testtoken2",
		Owner:      "testowner",
		Revision:   "testrevision2",
		Reference:  "",
		Valid:      true,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	m, err := insStore.GetInstancesBrief(map[string]bool{
		"insid1": true,
		"insid8": true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ins1 := m["insid1"]
	ins2 := m["insid8"]

	if ins1.Proto.GetService().GetValue() != "svcname1" {
		t.Fatal(fmt.Sprintf("get instance brief error, except %s, got %s",
			"svcname1", ins1.Proto.GetService().GetValue()))
	}

	if ins2.Proto.GetService().GetValue() != "svcname2" {
		t.Fatal(fmt.Sprintf("get instance brief error, except %s, got %s",
			"svcname2", ins2.Proto.GetService().GetValue()))
	}

	for _, instance := range m {
		fmt.Printf("get instance from brief %+v\n", instance)
	}

	// delete services
	sStore.DeleteService("svcid1", "", "")
	sStore.DeleteService("svcid2", "", "")
}

func TestInstanceStore_GetInstancesCount(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	c, err := insStore.GetInstancesCount()
	if err != nil {
		t.Fatal(err)
	}

	if c != routeCount*2 {
		t.Fatal("get instance count error")
	}
}

func TestInstanceStore_CheckInstancesExisted(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	m := map[string]bool{
		"insid1":          false,
		"insid2":          false,
		"test-not-exist":  false,
		"test-not-exist1": false,
	}

	mm, err := insStore.CheckInstancesExisted(m)
	if err != nil {
		t.Fatal(err)
	}

	if !mm["insid1"] || !mm["insid2"] || mm["test-not-exist"] || mm["test-not-exist1"] {
		t.Fatal("check instance existed error")
	}
}

func TestInstanceStore_DeleteInstance(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	err = insStore.DeleteInstance("insid1")
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	ins, err := insStore.GetInstance("insid1")
	if err != nil {
		t.Fatal(err)
	}

	if ins != nil {
		t.Fatal(fmt.Sprintf("delete instance error"))
	}
}

func TestInstanceStore_BatchDeleteInstances(t *testing.T) {
	handler, err := NewBoltHandler(&BoltConfig{FileName: "./table.bolt"})
	if nil != err {
		t.Fatal(err)
	}
	defer handler.Close()
	insStore := &instanceStore{handler: handler}

	err = insStore.BatchDeleteInstances([]interface{}{"insid2", "insid3"})
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	ins, err := insStore.GetInstance("insid2")
	if err != nil {
		t.Fatal(err)
	}

	if ins != nil {
		t.Fatal(fmt.Sprintf("delete instance error"))
	}

	ins, err = insStore.GetInstance("insid3")
	if err != nil {
		t.Fatal(err)
	}

	if ins != nil {
		t.Fatal(fmt.Sprintf("delete instance error"))
	}
}
