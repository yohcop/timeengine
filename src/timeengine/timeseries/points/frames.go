package points

const (
	s = 1
	m = 60
	h = 60 * 60
)

// A FrameSize represent a resolution.
// To build a FrameSize, use SelectFrameSize(int)
type FrameSize int64

// A KeyFrame is a timestamp normalized at a given FrameSize.
// Use FrameSize.KeyFrame to make one.
type KeyFrame int64

type Span struct {
	fs       FrameSize
	from, to KeyFrame
}

var InvalidFrameSize FrameSize = FrameSize(-1)
var RawFrameSize FrameSize = FrameSize(0)
var InvalidKeyFrame KeyFrame = KeyFrame(-1)

// On 500px:
// 1 year: 12h intervals.
// 1 month: ~1.5h intervals.
// 1 week: ~20 minutes intervals.
// 1 day: ~2.5 minutes intervals.
var AvailableFrameSizes = []FrameSize{
	//      // =secs | per day
	0,      // 0.001 | n/a - as many as pushed.
	1 * m,  // 60    | 1440
	1 * h,  // 3600  | 24
	24 * h, // 86400 | 1
}

// === FrameSize ==========================================

func SelectFrameSize(res int64) FrameSize {
	for i, ts := range AvailableFrameSizes[1:] {
		if res == int64(ts) {
			return ts
		}
		if res < int64(ts) {
			return AvailableFrameSizes[i]
		}
	}
	return AvailableFrameSizes[len(AvailableFrameSizes)-1]
}

func (r FrameSize) Secs() int64 {
	return int64(r)
}

func (r FrameSize) KeyFrame(ts int64) KeyFrame {
	if r <= 1 {
		return KeyFrame(ts)
	}
	return KeyFrame(ts - (ts % int64(r)))
}

func (r FrameSize) IsRaw() bool {
	return r == RawFrameSize
}

func (r FrameSize) IsValidNotRaw() bool {
	return r > RawFrameSize
}

func (r FrameSize) higherRes() FrameSize {
	if r == AvailableFrameSizes[0] {
		return r
	}
	for i, ts := range AvailableFrameSizes[1:] {
		if r == ts {
			return AvailableFrameSizes[i]
		}
	}
	return AvailableFrameSizes[0]
}

func (r FrameSize) lowerRes() FrameSize {
	var l = len(AvailableFrameSizes) - 1
	if r == AvailableFrameSizes[l] {
		return r
	}
	for i, ts := range AvailableFrameSizes {
		if r == ts && i < l-1 {
			return AvailableFrameSizes[i+1]
		}
	}
	return AvailableFrameSizes[l]
}

func (r FrameSize) numPoints() int {
	if r <= 1 {
		return 0
	}
	lower := r.lowerRes()
	return int(r / lower)
}

// === KeyFrame ===========================================

func (kf KeyFrame) Ts() int64 {
	return int64(kf)
}

// === Span ===============================================

func NewSpan(fs FrameSize, from, to int64) *Span {
	return &Span{fs, fs.KeyFrame(from), fs.KeyFrame(to)}
}

func (s *Span) HigherRes() *Span {
	l := s.fs.higherRes()
	from := l.KeyFrame(s.from.Ts())
	to := l.KeyFrame(s.to.Ts() + s.fs.Secs() - 1)
	return &Span{l, from, to}
}

func (s *Span) NumFrames() int {
	return 1 + int((s.to.Ts()-s.from.Ts())/s.fs.Secs())
}
