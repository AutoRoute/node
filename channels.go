package node

func splitChannel(c <-chan routingDecision) (<-chan routingDecision, <-chan routingDecision) {
	c1, c2 := make(chan routingDecision), make(chan routingDecision)
	go func() {
		for d := range c {
			c1 <- d
			c2 <- d
		}
	}()
	return c1, c2
}
