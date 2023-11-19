package rpc_client

import (
	"net/rpc"
	"time"

	"sequencer-node/types"

	"github.com/Layr-Labs/eigensdk-go/logging"
)

type SequencerRpcClientInterface interface {
	ProposeBlock(proposedBlock types.Block)
	SendSignedProposalResponse(response types.SignedResponse, leaderId uint32)
}

type SequencerRpcClient struct {
	logger                   logging.Logger
	rpcClients               []*rpc.Client
	sequencerIpPortAddresses []string
}

func NewSequencerRpcClient(sequencerIpPortAddresses []string, logger logging.Logger) *SequencerRpcClient {
	rpcClients := make([]*rpc.Client, len(sequencerIpPortAddresses))
	return &SequencerRpcClient{
		rpcClients:               rpcClients,
		logger:                   logger,
		sequencerIpPortAddresses: sequencerIpPortAddresses,
	}
}

func (c *SequencerRpcClient) ProposeBlock(proposedBlock types.Block) {
	for i := 0; i < len(c.rpcClients); i++ {
		go c.sendProposal(uint32(i), proposedBlock)
	}
}

func (c *SequencerRpcClient) SendSignedProposalResponse(response types.SignedResponse, leaderId uint32) {
	go c.sendSignature(leaderId, response)
}

func (c *SequencerRpcClient) SendSignedTimeoutResponse(response types.SignedTimeout, leaderId uint32) {
	go c.sendTimeout(leaderId, response)
}

func (c *SequencerRpcClient) dialSequencer(id uint32) error {
	client, err := rpc.DialHTTP("tcp", c.sequencerIpPortAddresses[id])
	if err != nil {
		return err
	}
	c.rpcClients[id] = client
	return nil
}

func (c *SequencerRpcClient) sendProposal(id uint32, proposedBlock types.Block) {
	if c.rpcClients[id] == nil {
		if err := c.dialSequencer(id); err != nil {
			c.logger.Errorf("error while dialing sequencer with id %d", id)
			return
		}
	}

	var reply bool

	for i := 0; i < 5; i++ {
		if err := c.rpcClients[id].Call("Sequencer.ProcessBlock", proposedBlock, &reply); err == nil {
			c.logger.Infof("successfully sent proposal to sequencer with id %d", id)
			return
		}
		c.logger.Errorf("error while sending block proposal to sequencer with id %d", id)
		c.logger.Infof("Retrying in 2 seconds")
		time.Sleep(2 * time.Second)
	}
}

func (c *SequencerRpcClient) sendSignature(id uint32, response types.SignedResponse) {
	if c.rpcClients[id] == nil {
		if err := c.dialSequencer(id); err != nil {
			c.logger.Errorf("error while dialing sequencer with id %d", id)
			return
		}
	}

	var reply bool

	for i := 0; i < 5; i++ {
		if err := c.rpcClients[id].Call("Sequencer.ProcessResponse", response, &reply); err == nil {
			c.logger.Infof("successfully sent response to sequencer with id %d", id)
			return
		}
		c.logger.Errorf("error while sending response to sequencer with id %d", id)
		c.logger.Infof("Retrying in 2 seconds")
		time.Sleep(2 * time.Second)
	}
}

func (c *SequencerRpcClient) sendTimeout(id uint32, response types.SignedTimeout) {
	if c.rpcClients[id] == nil {
		if err := c.dialSequencer(id); err != nil {
			c.logger.Errorf("error while dialing sequencer with id %d", id)
			return
		}
	}

	var reply bool

	for i := 0; i < 5; i++ {
		if err := c.rpcClients[id].Call("Sequencer.ProcessTimeout", response, &reply); err == nil {
			c.logger.Infof("successfully sent timeout to sequencer with id %d", id)
			return
		}
		c.logger.Errorf("error while sending timeout to sequencer with id %d", id)
		c.logger.Infof("Retrying in 2 seconds")
		time.Sleep(2 * time.Second)
	}
}
