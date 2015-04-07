package node

func SplitChannel(c <-chan RoutingDecision) (<-chan RoutingDecision, <-chan RoutingDecision) {
	c1, c2 := make(chan RoutingDecision), make(chan RoutingDecision)
	go func() {
		for d := range c {
			c1 <- d
			c2 <- d
		}
	}()
	return c1, c2
}
