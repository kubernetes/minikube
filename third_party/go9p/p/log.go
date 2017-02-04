package p

type Log struct {
	Data  interface{}
	Owner interface{}
	Type  int
}

type Logger struct {
	items   []*Log
	idx     int
	logchan chan *Log
	fltchan chan *flt
	rszchan chan int
}

type flt struct {
	owner   interface{}
	itype   int
	fltchan chan []*Log
}

func NewLogger(sz int) *Logger {
	if sz == 0 {
		return nil
	}

	l := new(Logger)
	l.items = make([]*Log, sz)
	l.logchan = make(chan *Log, 16)
	l.fltchan = make(chan *flt)
	l.rszchan = make(chan int)

	go l.doLog()
	return l
}

func (l *Logger) Resize(sz int) {
	if sz == 0 {
		return
	}

	l.rszchan <- sz
}

func (l *Logger) Log(data, owner interface{}, itype int) {
	l.logchan <- &Log{data, owner, itype}
}

func (l *Logger) Filter(owner interface{}, itype int) []*Log {
	c := make(chan []*Log)
	l.fltchan <- &flt{owner, itype, c}
	return <-c
}

func (l *Logger) doLog() {
	for {
		select {
		case it := <-l.logchan:
			if l.idx >= len(l.items) {
				l.idx = 0
			}

			l.items[l.idx] = it
			l.idx++

		case sz := <-l.rszchan:
			it := make([]*Log, sz)
			for i, j := l.idx, 0; j < len(it); j++ {
				if i >= len(l.items) {
					i = 0
				}

				it[j] = l.items[i]
				i++
				if i == l.idx {
					break
				}
			}

			l.items = it
			l.idx = 0

		case flt := <-l.fltchan:
			n := 0
			// we don't care about the order while counting
			for _, it := range l.items {
				if it == nil {
					continue
				}

				if (flt.owner == nil || it.Owner == flt.owner) &&
					(flt.itype == 0 || it.Type == flt.itype) {
					n++
				}
			}

			its := make([]*Log, n)
			for i, m := l.idx, 0; m < len(its); i++ {
				if i >= len(l.items) {
					i = 0
				}

				it := l.items[i]
				if it != nil && (flt.owner == nil || it.Owner == flt.owner) &&
					(flt.itype == 0 || it.Type == flt.itype) {
					its[m] = it
					m++
				}
			}

			flt.fltchan <- its
		}
	}
}
