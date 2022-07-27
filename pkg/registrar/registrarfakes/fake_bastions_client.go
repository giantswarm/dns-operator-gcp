// Code generated by counterfeiter. DO NOT EDIT.
package registrarfakes

import (
	"context"
	"sync"

	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/dns-operator-gcp/v2/pkg/registrar"
)

type FakeBastionsClient struct {
	GetBastionIPListStub        func(context.Context, *v1beta1.GCPCluster) ([]string, error)
	getBastionIPListMutex       sync.RWMutex
	getBastionIPListArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	getBastionIPListReturns struct {
		result1 []string
		result2 error
	}
	getBastionIPListReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBastionsClient) GetBastionIPList(arg1 context.Context, arg2 *v1beta1.GCPCluster) ([]string, error) {
	fake.getBastionIPListMutex.Lock()
	ret, specificReturn := fake.getBastionIPListReturnsOnCall[len(fake.getBastionIPListArgsForCall)]
	fake.getBastionIPListArgsForCall = append(fake.getBastionIPListArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.GetBastionIPListStub
	fakeReturns := fake.getBastionIPListReturns
	fake.recordInvocation("GetBastionIPList", []interface{}{arg1, arg2})
	fake.getBastionIPListMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBastionsClient) GetBastionIPListCallCount() int {
	fake.getBastionIPListMutex.RLock()
	defer fake.getBastionIPListMutex.RUnlock()
	return len(fake.getBastionIPListArgsForCall)
}

func (fake *FakeBastionsClient) GetBastionIPListCalls(stub func(context.Context, *v1beta1.GCPCluster) ([]string, error)) {
	fake.getBastionIPListMutex.Lock()
	defer fake.getBastionIPListMutex.Unlock()
	fake.GetBastionIPListStub = stub
}

func (fake *FakeBastionsClient) GetBastionIPListArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.getBastionIPListMutex.RLock()
	defer fake.getBastionIPListMutex.RUnlock()
	argsForCall := fake.getBastionIPListArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBastionsClient) GetBastionIPListReturns(result1 []string, result2 error) {
	fake.getBastionIPListMutex.Lock()
	defer fake.getBastionIPListMutex.Unlock()
	fake.GetBastionIPListStub = nil
	fake.getBastionIPListReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeBastionsClient) GetBastionIPListReturnsOnCall(i int, result1 []string, result2 error) {
	fake.getBastionIPListMutex.Lock()
	defer fake.getBastionIPListMutex.Unlock()
	fake.GetBastionIPListStub = nil
	if fake.getBastionIPListReturnsOnCall == nil {
		fake.getBastionIPListReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.getBastionIPListReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeBastionsClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getBastionIPListMutex.RLock()
	defer fake.getBastionIPListMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBastionsClient) recordInvocation(key string, args []interface{}) {
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

var _ registrar.BastionsClient = new(FakeBastionsClient)
