package sequencer

import (
	"context"
	"net/http"
	"net/rpc"
	"sync"

	leaderElection "sequencer-node/leader-election"
	rpcClient "sequencer-node/rpc-client"
	"sequencer-node/types"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/pkg/errors"
)

type SequencerInterface interface {
	Start(ctx context.Context) error
	ProcessBlock(proposedBlock types.Block, reply *bool) error
	ProcessResponse(response types.SignedResponse, reply *bool) error
}

type Sequencer struct {
	logger                   logging.Logger
	sequencerId              uint32
	serverIpPortAddr         string
	sequencerIpPortAddresses []string
	sequencerRpcClient       rpcClient.SequencerRpcClientInterface
	leaderElectionService    leaderElection.LeaderElectionServiceInterface

	//	state variables
	carrierBalances      map[string]uint32
	carrierBalancesMutex sync.RWMutex

	latestConfirmedBlock      uint32
	latestConfirmedBlockMutex sync.RWMutex

	qc      map[string][]bool
	qcMutex sync.RWMutex

	tc      map[string][]bool
	tcMutex sync.RWMutex
}

func NewSequencer(sequencerId uint32, serverIpPortAddr string, sequencerIpPortAddresses []string, stakeWeight map[uint32]uint32, logger logging.Logger) *Sequencer {
	return &Sequencer{
		logger:                logger,
		sequencerId:           sequencerId,
		serverIpPortAddr:      serverIpPortAddr,
		sequencerRpcClient:    rpcClient.NewSequencerRpcClient(sequencerIpPortAddresses, logger),
		leaderElectionService: leaderElection.NewLeaderElectionService(stakeWeight),
	}
}

func (s *Sequencer) Start(ctx context.Context) error {
	err := rpc.Register(s)
	if err != nil {
		return errors.Wrapf(err, "while registering sequencer server with addr %s", s.serverIpPortAddr)
	}
	rpc.HandleHTTP()
	err = http.ListenAndServe(s.serverIpPortAddr, nil)
	if err != nil {
		return errors.Wrapf(err, "while starting sequencer server with addr %s", s.serverIpPortAddr)
	}

	return nil
}

func (s *Sequencer) ProcessBlock(proposedBlock *types.Block, reply *bool) error {
	if s.latestConfirmedBlock == proposedBlock.Height-1 && s.validateQC(proposedBlock.QC) {
		//	update the latest confirmed block
		s.latestConfirmedBlockMutex.Lock()
		s.latestConfirmedBlock = proposedBlock.Height
		s.latestConfirmedBlockMutex.Unlock()

		//	update carrier balances
		s.carrierBalancesMutex.Lock()
		for _, transaction := range proposedBlock.Transactions {
			s.carrierBalances[transaction.SenderAddress] -= transaction.GasLimit
		}
		s.carrierBalancesMutex.Unlock()

		s.sequencerRpcClient.SendSignedProposalResponse(
			types.SignedResponse{
				Height:    proposedBlock.Height,
				Signature: true,
			},
			s.leaderElectionService.NextLeader(),
		)
	}
	return nil
}

func (s *Sequencer) ProcessResponse(response *types.SignedResponse, reply *bool) error {
	if s.validateSignature(response.PrevHash, response.Signature) && s.sequencerId == s.leaderElectionService.NextLeader() {
		s.qcMutex.Lock()

		s.qc[response.PrevHash] = append(s.qc[response.PrevHash], response.Signature)
		if s.validateQC(s.qc[response.PrevHash]) {

			s.sequencerRpcClient.ProposeBlock(types.Block{
				Height:          response.Height,
				LeaderSignature: true, // TODO: leader signature
				QC:              s.qc[response.PrevHash],
				//	TODO: Transactions
			})
		}

		s.qcMutex.Unlock()
	}
	return nil
}

func (s *Sequencer) ProcessTimeout(timeout *types.SignedTimeout, reply *bool) error {
	if s.validateSignature(timeout.PrevHash, timeout.Signature) && s.sequencerId == s.leaderElectionService.NextLeader() {
		s.tcMutex.Lock()

		s.tc[timeout.PrevHash] = append(s.tc[timeout.PrevHash], timeout.Signature)
		if s.validateQC(s.qc[timeout.PrevHash]) {

			s.sequencerRpcClient.ProposeBlock(types.Block{
				Height:          timeout.Height,
				LeaderSignature: true, // TODO: leader signature
				TC:              s.tc[timeout.PrevHash],
				//	TODO: Transactions
			})
		}

		s.tcMutex.Unlock()
	}
	return nil
}

func (s *Sequencer) validateQC(qc []bool) bool {
	//	TODO: fix this to sigs
	return len(qc) > len(s.sequencerIpPortAddresses)*2/3
}

func (s *Sequencer) validateSignature(prevHash string, sig bool) bool {
	//	TODO: fix this to sig
	return sig
}
