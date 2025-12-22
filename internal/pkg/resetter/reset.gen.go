package resetter

func (rs *ResetableStruct) Reset() {
	if rs == nil {
		return
	}

	rs.i = 0
	rs.str = ""
	if rs.strP != nil {
		*rs.strP = ""
	}
	rs.s = rs.s[:0]
	clear(rs.m)
	if resetter, ok := any(rs.child).(interface{ Reset() }); ok && rs.child != nil {
		resetter.Reset()
	}
}
