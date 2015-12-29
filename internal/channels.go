package node

func splitChannel(c <-chan routingDecision) (<-chan routingDecision, <-chan routingDecision, chan bool) {
	c1, c2 := make(chan routingDecision), make(chan routingDecision)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case d := <-c:
				c1 <- d
				c2 <- d
			case <-quit:
				return
			}
		}
	}()
	return c1, c2, quit
}
