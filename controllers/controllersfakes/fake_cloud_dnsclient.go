// Code generated by counterfeiter. DO NOT EDIT.
package controllersfakes

import (
	"context"
	"sync"

	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/dns-operator-gcp/controllers"
)

type FakeCloudDNSClient struct {
	CreateARecordsStub        func(context.Context, *v1beta1.GCPCluster) error
	createARecordsMutex       sync.RWMutex
	createARecordsArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	createARecordsReturns struct {
		result1 error
	}
	createARecordsReturnsOnCall map[int]struct {
		result1 error
	}
	CreateZoneStub        func(context.Context, *v1beta1.GCPCluster) error
	createZoneMutex       sync.RWMutex
	createZoneArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	createZoneReturns struct {
		result1 error
	}
	createZoneReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteARecordsStub        func(context.Context, *v1beta1.GCPCluster) error
	deleteARecordsMutex       sync.RWMutex
	deleteARecordsArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	deleteARecordsReturns struct {
		result1 error
	}
	deleteARecordsReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteZoneStub        func(context.Context, *v1beta1.GCPCluster) error
	deleteZoneMutex       sync.RWMutex
	deleteZoneArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	deleteZoneReturns struct {
		result1 error
	}
	deleteZoneReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCloudDNSClient) CreateARecords(arg1 context.Context, arg2 *v1beta1.GCPCluster) error {
	fake.createARecordsMutex.Lock()
	ret, specificReturn := fake.createARecordsReturnsOnCall[len(fake.createARecordsArgsForCall)]
	fake.createARecordsArgsForCall = append(fake.createARecordsArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.CreateARecordsStub
	fakeReturns := fake.createARecordsReturns
	fake.recordInvocation("CreateARecords", []interface{}{arg1, arg2})
	fake.createARecordsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCloudDNSClient) CreateARecordsCallCount() int {
	fake.createARecordsMutex.RLock()
	defer fake.createARecordsMutex.RUnlock()
	return len(fake.createARecordsArgsForCall)
}

func (fake *FakeCloudDNSClient) CreateARecordsCalls(stub func(context.Context, *v1beta1.GCPCluster) error) {
	fake.createARecordsMutex.Lock()
	defer fake.createARecordsMutex.Unlock()
	fake.CreateARecordsStub = stub
}

func (fake *FakeCloudDNSClient) CreateARecordsArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.createARecordsMutex.RLock()
	defer fake.createARecordsMutex.RUnlock()
	argsForCall := fake.createARecordsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCloudDNSClient) CreateARecordsReturns(result1 error) {
	fake.createARecordsMutex.Lock()
	defer fake.createARecordsMutex.Unlock()
	fake.CreateARecordsStub = nil
	fake.createARecordsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) CreateARecordsReturnsOnCall(i int, result1 error) {
	fake.createARecordsMutex.Lock()
	defer fake.createARecordsMutex.Unlock()
	fake.CreateARecordsStub = nil
	if fake.createARecordsReturnsOnCall == nil {
		fake.createARecordsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createARecordsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) CreateZone(arg1 context.Context, arg2 *v1beta1.GCPCluster) error {
	fake.createZoneMutex.Lock()
	ret, specificReturn := fake.createZoneReturnsOnCall[len(fake.createZoneArgsForCall)]
	fake.createZoneArgsForCall = append(fake.createZoneArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.CreateZoneStub
	fakeReturns := fake.createZoneReturns
	fake.recordInvocation("CreateZone", []interface{}{arg1, arg2})
	fake.createZoneMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCloudDNSClient) CreateZoneCallCount() int {
	fake.createZoneMutex.RLock()
	defer fake.createZoneMutex.RUnlock()
	return len(fake.createZoneArgsForCall)
}

func (fake *FakeCloudDNSClient) CreateZoneCalls(stub func(context.Context, *v1beta1.GCPCluster) error) {
	fake.createZoneMutex.Lock()
	defer fake.createZoneMutex.Unlock()
	fake.CreateZoneStub = stub
}

func (fake *FakeCloudDNSClient) CreateZoneArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.createZoneMutex.RLock()
	defer fake.createZoneMutex.RUnlock()
	argsForCall := fake.createZoneArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCloudDNSClient) CreateZoneReturns(result1 error) {
	fake.createZoneMutex.Lock()
	defer fake.createZoneMutex.Unlock()
	fake.CreateZoneStub = nil
	fake.createZoneReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) CreateZoneReturnsOnCall(i int, result1 error) {
	fake.createZoneMutex.Lock()
	defer fake.createZoneMutex.Unlock()
	fake.CreateZoneStub = nil
	if fake.createZoneReturnsOnCall == nil {
		fake.createZoneReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createZoneReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) DeleteARecords(arg1 context.Context, arg2 *v1beta1.GCPCluster) error {
	fake.deleteARecordsMutex.Lock()
	ret, specificReturn := fake.deleteARecordsReturnsOnCall[len(fake.deleteARecordsArgsForCall)]
	fake.deleteARecordsArgsForCall = append(fake.deleteARecordsArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.DeleteARecordsStub
	fakeReturns := fake.deleteARecordsReturns
	fake.recordInvocation("DeleteARecords", []interface{}{arg1, arg2})
	fake.deleteARecordsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCloudDNSClient) DeleteARecordsCallCount() int {
	fake.deleteARecordsMutex.RLock()
	defer fake.deleteARecordsMutex.RUnlock()
	return len(fake.deleteARecordsArgsForCall)
}

func (fake *FakeCloudDNSClient) DeleteARecordsCalls(stub func(context.Context, *v1beta1.GCPCluster) error) {
	fake.deleteARecordsMutex.Lock()
	defer fake.deleteARecordsMutex.Unlock()
	fake.DeleteARecordsStub = stub
}

func (fake *FakeCloudDNSClient) DeleteARecordsArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.deleteARecordsMutex.RLock()
	defer fake.deleteARecordsMutex.RUnlock()
	argsForCall := fake.deleteARecordsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCloudDNSClient) DeleteARecordsReturns(result1 error) {
	fake.deleteARecordsMutex.Lock()
	defer fake.deleteARecordsMutex.Unlock()
	fake.DeleteARecordsStub = nil
	fake.deleteARecordsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) DeleteARecordsReturnsOnCall(i int, result1 error) {
	fake.deleteARecordsMutex.Lock()
	defer fake.deleteARecordsMutex.Unlock()
	fake.DeleteARecordsStub = nil
	if fake.deleteARecordsReturnsOnCall == nil {
		fake.deleteARecordsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteARecordsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) DeleteZone(arg1 context.Context, arg2 *v1beta1.GCPCluster) error {
	fake.deleteZoneMutex.Lock()
	ret, specificReturn := fake.deleteZoneReturnsOnCall[len(fake.deleteZoneArgsForCall)]
	fake.deleteZoneArgsForCall = append(fake.deleteZoneArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.DeleteZoneStub
	fakeReturns := fake.deleteZoneReturns
	fake.recordInvocation("DeleteZone", []interface{}{arg1, arg2})
	fake.deleteZoneMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCloudDNSClient) DeleteZoneCallCount() int {
	fake.deleteZoneMutex.RLock()
	defer fake.deleteZoneMutex.RUnlock()
	return len(fake.deleteZoneArgsForCall)
}

func (fake *FakeCloudDNSClient) DeleteZoneCalls(stub func(context.Context, *v1beta1.GCPCluster) error) {
	fake.deleteZoneMutex.Lock()
	defer fake.deleteZoneMutex.Unlock()
	fake.DeleteZoneStub = stub
}

func (fake *FakeCloudDNSClient) DeleteZoneArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.deleteZoneMutex.RLock()
	defer fake.deleteZoneMutex.RUnlock()
	argsForCall := fake.deleteZoneArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCloudDNSClient) DeleteZoneReturns(result1 error) {
	fake.deleteZoneMutex.Lock()
	defer fake.deleteZoneMutex.Unlock()
	fake.DeleteZoneStub = nil
	fake.deleteZoneReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) DeleteZoneReturnsOnCall(i int, result1 error) {
	fake.deleteZoneMutex.Lock()
	defer fake.deleteZoneMutex.Unlock()
	fake.DeleteZoneStub = nil
	if fake.deleteZoneReturnsOnCall == nil {
		fake.deleteZoneReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteZoneReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCloudDNSClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createARecordsMutex.RLock()
	defer fake.createARecordsMutex.RUnlock()
	fake.createZoneMutex.RLock()
	defer fake.createZoneMutex.RUnlock()
	fake.deleteARecordsMutex.RLock()
	defer fake.deleteARecordsMutex.RUnlock()
	fake.deleteZoneMutex.RLock()
	defer fake.deleteZoneMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCloudDNSClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ controllers.CloudDNSClient = new(FakeCloudDNSClient)
