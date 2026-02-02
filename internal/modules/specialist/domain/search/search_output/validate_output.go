package searchoutput

func (l *ListSearchOutput) IsEmpty() bool {
	return len(l.Specialists) == 0
}
