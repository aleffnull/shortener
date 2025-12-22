package pool

func (rs *ObjectWithState) Reset() {
	if rs == nil {
		return
	}

	rs.State = 0
}
