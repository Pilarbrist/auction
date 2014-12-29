// This file was generated by counterfeiter
package fakes

import (
	"os"
	"sync"

	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type FakeAuctionRunner struct {
	RunStub        func(signals <-chan os.Signal, ready chan<- struct{}) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		signals <-chan os.Signal
		ready   chan<- struct{}
	}
	runReturns struct {
		result1 error
	}
	AddLRPStartForAuctionStub        func(models.LRPStart)
	addLRPStartForAuctionMutex       sync.RWMutex
	addLRPStartForAuctionArgsForCall []struct {
		arg1 models.LRPStart
	}
	ScheduleTasksForAuctionsStub        func([]models.Task)
	scheduleTasksForAuctionsMutex       sync.RWMutex
	scheduleTasksForAuctionsArgsForCall []struct {
		arg1 []models.Task
	}
}

func (fake *FakeAuctionRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	fake.runMutex.Lock()
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		signals <-chan os.Signal
		ready   chan<- struct{}
	}{signals, ready})
	fake.runMutex.Unlock()
	if fake.RunStub != nil {
		return fake.RunStub(signals, ready)
	} else {
		return fake.runReturns.result1
	}
}

func (fake *FakeAuctionRunner) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeAuctionRunner) RunArgsForCall(i int) (<-chan os.Signal, chan<- struct{}) {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return fake.runArgsForCall[i].signals, fake.runArgsForCall[i].ready
}

func (fake *FakeAuctionRunner) RunReturns(result1 error) {
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAuctionRunner) AddLRPStartForAuction(arg1 models.LRPStart) {
	fake.addLRPStartForAuctionMutex.Lock()
	fake.addLRPStartForAuctionArgsForCall = append(fake.addLRPStartForAuctionArgsForCall, struct {
		arg1 models.LRPStart
	}{arg1})
	fake.addLRPStartForAuctionMutex.Unlock()
	if fake.AddLRPStartForAuctionStub != nil {
		fake.AddLRPStartForAuctionStub(arg1)
	}
}

func (fake *FakeAuctionRunner) AddLRPStartForAuctionCallCount() int {
	fake.addLRPStartForAuctionMutex.RLock()
	defer fake.addLRPStartForAuctionMutex.RUnlock()
	return len(fake.addLRPStartForAuctionArgsForCall)
}

func (fake *FakeAuctionRunner) AddLRPStartForAuctionArgsForCall(i int) models.LRPStart {
	fake.addLRPStartForAuctionMutex.RLock()
	defer fake.addLRPStartForAuctionMutex.RUnlock()
	return fake.addLRPStartForAuctionArgsForCall[i].arg1
}

func (fake *FakeAuctionRunner) ScheduleTasksForAuctions(arg1 []models.Task) {
	fake.scheduleTasksForAuctionsMutex.Lock()
	fake.scheduleTasksForAuctionsArgsForCall = append(fake.scheduleTasksForAuctionsArgsForCall, struct {
		arg1 []models.Task
	}{arg1})
	fake.scheduleTasksForAuctionsMutex.Unlock()
	if fake.ScheduleTasksForAuctionsStub != nil {
		fake.ScheduleTasksForAuctionsStub(arg1)
	}
}

func (fake *FakeAuctionRunner) ScheduleTasksForAuctionsCallCount() int {
	fake.scheduleTasksForAuctionsMutex.RLock()
	defer fake.scheduleTasksForAuctionsMutex.RUnlock()
	return len(fake.scheduleTasksForAuctionsArgsForCall)
}

func (fake *FakeAuctionRunner) ScheduleTasksForAuctionsArgsForCall(i int) []models.Task {
	fake.scheduleTasksForAuctionsMutex.RLock()
	defer fake.scheduleTasksForAuctionsMutex.RUnlock()
	return fake.scheduleTasksForAuctionsArgsForCall[i].arg1
}

var _ auctiontypes.AuctionRunner = new(FakeAuctionRunner)
