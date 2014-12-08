package auctionrunner

import (
	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type Cell struct {
	client auctiontypes.CellRep
	state  auctiontypes.CellState

	workToCommit auctiontypes.Work
}

func NewCell(client auctiontypes.CellRep, state auctiontypes.CellState) *Cell {
	return &Cell{
		client: client,
		state:  state,
	}
}

func (c *Cell) ScoreForLRPStartAuction(lrpStartAuction models.LRPStartAuction) (float64, error) {
	err := c.canHandleLRPStartAuction(lrpStartAuction)
	if err != nil {
		return 0, err
	}

	numberOfInstancesWithMatchingProcessGuid := 0
	for _, lrp := range c.state.LRPs {
		if lrp.ProcessGuid == lrpStartAuction.DesiredLRP.ProcessGuid {
			numberOfInstancesWithMatchingProcessGuid++
		}
	}

	remainingResources := c.state.AvailableResources
	remainingResources.MemoryMB -= lrpStartAuction.DesiredLRP.MemoryMB
	remainingResources.DiskMB -= lrpStartAuction.DesiredLRP.DiskMB
	remainingResources.Containers -= 1

	resourceScore := c.computeScore(remainingResources, numberOfInstancesWithMatchingProcessGuid)

	return resourceScore, nil
}

func (c *Cell) ScoreForTask(task models.Task) (float64, error) {
	err := c.canHandleTask(task)
	if err != nil {
		return 0, err
	}

	remainingResources := c.state.AvailableResources
	remainingResources.MemoryMB -= task.MemoryMB
	remainingResources.DiskMB -= task.DiskMB
	remainingResources.Containers -= 1

	resourceScore := c.computeTaskScore(remainingResources)

	return resourceScore, nil
}

func (c *Cell) ScoreForLRPStopAuction(lrpStopAuction models.LRPStopAuction) (float64, []string, error) {
	matchingLRPs := []auctiontypes.LRP{}
	numberOfInstancesWithMatchingProcessGuidButDifferentIndex := 0
	for _, lrp := range c.state.LRPs {
		if lrp.ProcessGuid == lrpStopAuction.ProcessGuid {
			if lrp.Index == lrpStopAuction.Index {
				matchingLRPs = append(matchingLRPs, lrp)
			} else {
				numberOfInstancesWithMatchingProcessGuidButDifferentIndex++
			}
		}
	}

	if len(matchingLRPs) == 0 {
		return 0, nil, auctiontypes.ErrorNothingToStop
	}

	remainingResources := c.state.AvailableResources
	instanceGuids := make([]string, len(matchingLRPs))

	for i, lrp := range matchingLRPs {
		instanceGuids[i] = lrp.InstanceGuid
		remainingResources.MemoryMB += lrp.MemoryMB
		remainingResources.DiskMB += lrp.DiskMB
		remainingResources.Containers += 1
	}

	resourceScore := c.computeScore(remainingResources, numberOfInstancesWithMatchingProcessGuidButDifferentIndex)

	return resourceScore, instanceGuids, nil
}

func (c *Cell) StartLRP(lrpStartAuction models.LRPStartAuction) error {
	err := c.canHandleLRPStartAuction(lrpStartAuction)
	if err != nil {
		return err
	}

	c.state.LRPs = append(c.state.LRPs, auctiontypes.LRP{
		ProcessGuid:  lrpStartAuction.DesiredLRP.ProcessGuid,
		InstanceGuid: lrpStartAuction.InstanceGuid,
		Index:        lrpStartAuction.Index,
		MemoryMB:     lrpStartAuction.DesiredLRP.MemoryMB,
		DiskMB:       lrpStartAuction.DesiredLRP.DiskMB,
	})

	c.state.AvailableResources.MemoryMB -= lrpStartAuction.DesiredLRP.MemoryMB
	c.state.AvailableResources.DiskMB -= lrpStartAuction.DesiredLRP.DiskMB
	c.state.AvailableResources.Containers -= 1

	c.workToCommit.LRPStarts = append(c.workToCommit.LRPStarts, lrpStartAuction)

	return nil
}

func (c *Cell) StartTask(task models.Task) error {
	err := c.canHandleTask(task)
	if err != nil {
		return err
	}

	c.state.Tasks = append(c.state.Tasks, auctiontypes.Task{
		TaskGuid: task.TaskGuid,
		MemoryMB: task.MemoryMB,
		DiskMB:   task.DiskMB,
	})

	c.state.AvailableResources.MemoryMB -= task.MemoryMB
	c.state.AvailableResources.DiskMB -= task.DiskMB
	c.state.AvailableResources.Containers -= 1

	c.workToCommit.Tasks = append(c.workToCommit.Tasks, task)

	return nil
}

func (c *Cell) StopLRP(stop models.ActualLRP) error {
	indexToDelete := -1
	for i, lrp := range c.state.LRPs {
		if lrp.ProcessGuid != stop.ProcessGuid {
			continue
		}
		if lrp.InstanceGuid != stop.InstanceGuid {
			continue
		}
		if lrp.Index != stop.Index {
			continue
		}
		indexToDelete = i
		break
	}

	if indexToDelete == -1 {
		return auctiontypes.ErrorNothingToStop
	}

	c.state.AvailableResources.MemoryMB += c.state.LRPs[indexToDelete].MemoryMB
	c.state.AvailableResources.DiskMB += c.state.LRPs[indexToDelete].DiskMB
	c.state.AvailableResources.Containers += 1

	c.state.LRPs = append(c.state.LRPs[0:indexToDelete], c.state.LRPs[indexToDelete+1:]...)
	c.workToCommit.LRPStops = append(c.workToCommit.LRPStops, stop)

	return nil
}

func (c *Cell) Commit() auctiontypes.Work {
	if len(c.workToCommit.LRPStarts) == 0 && len(c.workToCommit.LRPStops) == 0 && len(c.workToCommit.Tasks) == 0 {
		return auctiontypes.Work{}
	}

	failedWork, err := c.client.Perform(c.workToCommit)
	if err != nil {
		//an error may indicate partial failure
		//in this case we don't reschedule work in order to make sure we don't
		//create duplicates of things -- we'll let the converger figure things out for us later
		return auctiontypes.Work{}
	}
	return failedWork
}

func (c *Cell) canHandleLRPStartAuction(lrpStartAuction models.LRPStartAuction) error {
	if c.state.Stack != lrpStartAuction.DesiredLRP.Stack {
		return auctiontypes.ErrorStackMismatch
	}
	if c.state.AvailableResources.MemoryMB < lrpStartAuction.DesiredLRP.MemoryMB {
		return auctiontypes.ErrorInsufficientResources
	}
	if c.state.AvailableResources.DiskMB < lrpStartAuction.DesiredLRP.DiskMB {
		return auctiontypes.ErrorInsufficientResources
	}
	if c.state.AvailableResources.Containers < 1 {
		return auctiontypes.ErrorInsufficientResources
	}

	return nil
}

func (c *Cell) canHandleTask(task models.Task) error {
	if c.state.Stack != task.Stack {
		return auctiontypes.ErrorStackMismatch
	}
	if c.state.AvailableResources.MemoryMB < task.MemoryMB {
		return auctiontypes.ErrorInsufficientResources
	}
	if c.state.AvailableResources.DiskMB < task.DiskMB {
		return auctiontypes.ErrorInsufficientResources
	}
	if c.state.AvailableResources.Containers < 1 {
		return auctiontypes.ErrorInsufficientResources
	}

	return nil
}

func (c *Cell) computeScore(remainingResources auctiontypes.Resources, numInstances int) float64 {
	fractionUsedMemory := 1.0 - float64(remainingResources.MemoryMB)/float64(c.state.TotalResources.MemoryMB)
	fractionUsedDisk := 1.0 - float64(remainingResources.DiskMB)/float64(c.state.TotalResources.DiskMB)
	fractionUsedContainers := 1.0 - float64(remainingResources.Containers)/float64(c.state.TotalResources.Containers)

	resourceScore := (fractionUsedMemory + fractionUsedDisk + fractionUsedContainers) / 3.0
	resourceScore += float64(numInstances)

	return resourceScore
}

func (c *Cell) computeTaskScore(remainingResources auctiontypes.Resources) float64 {
	fractionUsedMemory := 1.0 - float64(remainingResources.MemoryMB)/float64(c.state.TotalResources.MemoryMB)
	fractionUsedDisk := 1.0 - float64(remainingResources.DiskMB)/float64(c.state.TotalResources.DiskMB)
	fractionUsedContainers := 1.0 - float64(remainingResources.Containers)/float64(c.state.TotalResources.Containers)

	resourceScore := (fractionUsedMemory + fractionUsedDisk + fractionUsedContainers) / 3.0

	return resourceScore
}
