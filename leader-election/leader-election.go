package leader_election

type LeaderElectionServiceInterface interface {
	NextLeader() uint32
	NsLeader() bool
}

type LeaderElectionService struct {
	//	TODO:this should store the stake weight
	sequencerId uint32
	stakeWeight map[uint32]uint32
}

func NewLeaderElectionService(stakeWeight map[uint32]uint32) *LeaderElectionService {
	return &LeaderElectionService{stakeWeight: stakeWeight}
}

func (s *LeaderElectionService) NextLeader() uint32 {
	// TODO: fix this
	return 1
}

func (s *LeaderElectionService) NsLeader() bool {
	//	TODO: fix this
	return s.sequencerId == 1
}
