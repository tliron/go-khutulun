package conductor

// memberlist.Delegate interface
func (self *Server) NodeMeta(limit int) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Server) NotifyMsg(bytes []byte) {
}

// memberlist.Delegate interface
func (self *Server) GetBroadcasts(overhead int, limit int) [][]byte {
	return self.clusterQueue.GetBroadcasts(overhead, limit)
}

// memberlist.Delegate interface
func (self *Server) LocalState(join bool) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Server) MergeRemoteState(buf []byte, join bool) {
}
