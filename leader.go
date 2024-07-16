package cloudy

type LeaderElector interface {
	Elect(func(isLeader bool))
	Connect(cfg interface{}) error
}
