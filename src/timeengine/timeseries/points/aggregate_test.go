package points

import (
	"code.google.com/p/gomock/gomock"
	"log"
	"math"
	"testing"
	"timeengine/mock_ae"
)

var _ = log.Println

const epsilon = 0.0000001

func makeTestFrame(fs, kf int64, p1, p2 int64, v1, v2 float64) *frame {
	return newFrame("foo.bar", FrameSize(fs), KeyFrame(kf),
		[]StatsDataPoint{
			newStats(p1).update(v1, v1, v1, v1),
			newStats(p2).update(v2, v2, v2, v2),
		},
	)
}

func TestBuildFrames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := makeTestFrame(60, 3600, 3600, 3610, 1.1, 1.2)
	p2 := makeTestFrame(60, 7260, 7264, 7268, 2.1, 2.2)

	// The limits between which the search is going to happen.
	l1 := aggregateKeyName("foo.bar", FrameSize(60), KeyFrame(3600))
	l2 := aggregateKeyName("foo.bar", FrameSize(60), KeyFrame(10740))

	mockae := mock_ae.NewMockContext(mockCtrl)
	gomock.InOrder(
		mockae.EXPECT().DsGetBetweenKeys("F", l1, l2, -1, gomock.Any()).
			SetArg(4, []*frame{p1, p2}).
			Return([]string{p1.Key(), p2.Key()}, nil),
		mockae.EXPECT().PutMulti(
			"F", []string{
				aggregateKeyName("foo.bar", FrameSize(3600), KeyFrame(3600)),
				aggregateKeyName("foo.bar", FrameSize(3600), KeyFrame(7200)),
			}, gomock.Any()).
			Return(nil),
	)

	log.Println("=====")
	frames, err := BuildFrames(
		mockae, "foo.bar", NewSpan(SelectFrameSize(3600), 3600, 7200))

	if err != nil {
		t.Fatalf("Error? %s", err.Error())
	}
	if expected, got := 2, len(frames); expected != got {
		t.Fatalf("Wrong number of points: %d, expected %d", got, expected)
	}
	if expected, got := 1.15, frames[0].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
	if expected, got := 2.15, frames[1].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
}

func TestBuildFramesFirstLevel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := NewP(1.1, 60, "foo.bar")
	p2 := NewP(1.2, 63, "foo.bar")

	// The limits between which the search is going to happen.
	l1 := keyAt("foo.bar", 60)
	l2 := keyAt("foo.bar", 119)

	mockae := mock_ae.NewMockContext(mockCtrl)
	gomock.InOrder(
		mockae.EXPECT().DsGetBetweenKeys("P", l1, l2, -1, gomock.Any()).
			SetArg(4, []*P{p1, p2}).
			Return([]string{p1.Key(), p2.Key()}, nil),
		mockae.EXPECT().PutMulti(
			"F", []string{
				aggregateKeyName("foo.bar", FrameSize(60), KeyFrame(60)),
			}, gomock.Any()).
			Return(nil),
	)

	log.Println("=====")
	frames, err := BuildFrames(
		mockae, "foo.bar", NewSpan(SelectFrameSize(60), 60, 60))

	if err != nil {
		t.Fatalf("Error? %s", err.Error())
	}
	if expected, got := 1, len(frames); expected != got {
		t.Fatalf("Wrong number of points: %d, expected %d", got, expected)
	}
	if expected, got := 1.15, frames[0].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
}

/*
func TestGetFromLowerRes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := &pointAggregate{
		kf:      KeyFrame(60),
		ptReady: true,
		points:  []*P{&P{V: 1.0}, &P{V: 1.1}},
	}
	p2 := &pointAggregate{
		kf:      KeyFrame(120),
		ptReady: true,
		points:  []*P{&P{V: 2.0}, &P{V: 2.1}},
	}
	mockae := mock_ae.NewMockContext(mockCtrl)
	mockae.EXPECT().DsGetBetweenKeys(
		"F", "foo.bar@60@60", "foo.bar@60@120", gomock.Any()).
		SetArg(3, []*pointAggregate{p1, p2}).
		Return([]string{"foo.bar@60@60", "foo.bar@60@120"}, nil)

	p, missingFrom, missingTo, err := getFromLowerRes(
		mockae, "foo.bar",
    FrameSize(1), KeyFrame(60), KeyFrame(130),
    FrameSize(60), KeyFrame(60), KeyFrame(120))

	if err != nil {
		t.Errorf("Error? %s", err.Error())
	}
	if missingFrom != -1 {
		t.Errorf("We should have gotten all the points (missingFrom=%d)",
			missingFrom)
	}
	if missingTo != -1 {
		t.Errorf("We should have gotten all the points (missingTo=%d)",
			missingTo)
	}
	if len(p) != 4 {
		t.Errorf("Wrong number of points: %d, expected 4", len(p))
	}
}

func TestFromLowerResMissingSecond(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := &pointAggregate{
		kf:      KeyFrame(60),
		ptReady: true,
		points:  []*P{&P{V: 1.0}, &P{V: 1.1}},
	}
	mockae := mock_ae.NewMockContext(mockCtrl)
	mockae.EXPECT().DsGetBetweenKeys(
		"F", "foo.bar@60@60", "foo.bar@60@180", gomock.Any()).
		SetArg(3, []*pointAggregate{p1}).
		Return([]string{"foo.bar@60@60"}, nil)

	p, missingFrom, missingTo, err := getFromLowerRes(
		mockae, "foo.bar",
    FrameSize(1), KeyFrame(60), KeyFrame(195),
    FrameSize(60), KeyFrame(60), KeyFrame(180))

	if err != nil {
		t.Errorf("Error? %s", err.Error())
	}
	if missingFrom != 120 {
		t.Errorf("Wrong missing start block (missingFrom=%d)", missingFrom)
	}
	if missingTo != 195 {
		t.Errorf("Wrong missing end block (missingTo=%d)", missingTo)
	}
	if len(p) != 2 {
		t.Errorf("Wrong number of points: %d, expected 2", len(p))
	}
}

func TestGetFromLowerResMissingFirst(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p2 := &pointAggregate{
		kf:      KeyFrame(120),
		ptReady: true,
		points:  []*P{&P{V: 2.0}, &P{V: 2.1}},
	}
	mockae := mock_ae.NewMockContext(mockCtrl)
	mockae.EXPECT().DsGetBetweenKeys(
		"F", "foo.bar@60@60", "foo.bar@60@120", gomock.Any()).
		SetArg(3, []*pointAggregate{p2}).
		Return([]string{"foo.bar@60@120"}, nil)

	p, missingFrom, missingTo, err := getFromLowerRes(
		mockae, "foo.bar",
    FrameSize(1), KeyFrame(60), KeyFrame(130),
    FrameSize(60), KeyFrame(60), KeyFrame(120))

	if err != nil {
		t.Errorf("Error? %s", err.Error())
	}
	if missingFrom != 60 {
		t.Errorf("We should have gotten all the points (missingFrom=%d)",
			missingFrom)
	}
	if missingTo != 60 {
		t.Errorf("We should have gotten all the points (missingTo=%d)",
			missingTo)
	}
	if len(p) != 2 {
		t.Errorf("Wrong number of points: %d, expected 2", len(p))
	}
}

func TestGenFramesAllPresent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := &pointAggregate{kf: KeyFrame(60), Avg: 43.81}
	p2 := &pointAggregate{kf: KeyFrame(120), Avg: 12.95}
	mockae := mock_ae.NewMockContext(mockCtrl)
	mockae.EXPECT().DsGetBetweenKeys(
		"F", "foo.bar@60@60", "foo.bar@60@120", gomock.Any()).
		SetArg(3, []*pointAggregate{p1, p2}).
		Return([]string{"foo.bar@60@60", "foo.bar@60@120"}, nil)

	p, err := genFrames(
		mockae, "foo.bar", FrameSize(60), KeyFrame(60), KeyFrame(120))

	if err != nil {
		t.Errorf("Did not expect an error: %s", err.Error())
	}
	if len(p) != 2 {
		t.Errorf("Expected 2 points, got %d", len(p))
	}
	if p[0].V != 43.81 {
		t.Errorf("Wrong value for p0: got %f", p[0].V)
	}
	if p[1].V != 12.95 {
		t.Errorf("Wrong value for p1: got %f", p[1].V)
	}
}

func TestGenFramesManyRes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

  // 60                  120                 180
  // [      43.81         ]
  //                             70    90
  //                       [        80        ]
	// Call made to get points at the expected resolution.
	p1 := &pointAggregate{kf: KeyFrame(60), Avg: 43.81}
  // Call made to get raw points.
  p20 := &P{V:70}
  p21 := &P{V:90}

	mockae := mock_ae.NewMockContext(mockCtrl)
	gomock.InOrder(
		mockae.EXPECT().DsGetBetweenKeys(
			"F", "foo.bar@60@60", "foo.bar@60@120", gomock.Any()).
			SetArg(3, []*pointAggregate{p1}).
			Return([]string{"foo.bar@60@60"}, nil),
		mockae.EXPECT().DsGetBetweenKeys(
			"P", "foo.bar@120", "foo.bar@179", gomock.Any()).
			SetArg(3, []*P{p20, p21}).
			Return([]string{"foo.bar@150", "foo.bar@165"}, nil),
		mockae.EXPECT().PutMulti(
			"F", []string{"foo.bar@60@120"}, gomock.Any()).
			Return(nil),
	)

	p, err := genFrames(
		mockae, "foo.bar", FrameSize(60), KeyFrame(60), KeyFrame(120))

	if err != nil {
		t.Errorf("Did not expect an error: %s", err.Error())
	}
	if len(p) != 2 {
		t.Errorf("Expected 2 points, got %d", len(p))
	}
	if p[0].V != 43.81 {
		t.Errorf("Wrong value for p0: got %f", p[0].V)
	}
	if p[1].V != 80 {
		t.Errorf("Wrong value for p1: got %f", p[1].V)
	}
}

func TestGetFromFrames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

  // 60                  120                 180
  // [      43.81         ]
  //                              70    90
  //                       [         80        ]
	// Call made to get points at the expected resolution.
	p1 := &pointAggregate{kf: KeyFrame(60), Avg: 43.81}
  // Call made to get raw points.
  p20 := &P{V:70}
  p21 := &P{V:90}

	mockae := mock_ae.NewMockContext(mockCtrl)
	gomock.InOrder(
    // This is the call at a higher resolution
		mockae.EXPECT().DsGetBetweenKeys(
			"F", "foo.bar@3600@0", "foo.bar@3600@0", gomock.Any()).
			Return([]string{}, nil),
    // Then, we call at the expected resolution, on a range that
    // covers our times.
		mockae.EXPECT().DsGetBetweenKeys(
			"F", "foo.bar@60@60", "foo.bar@60@120", gomock.Any()).
			SetArg(3, []*pointAggregate{p1}).
			Return([]string{"foo.bar@60@60"}, nil),
		mockae.EXPECT().DsGetBetweenKeys(
			"P", "foo.bar@120", "foo.bar@179", gomock.Any()).
			SetArg(3, []*P{p20, p21}).
			Return([]string{"foo.bar@150", "foo.bar@165"}, nil),
		mockae.EXPECT().PutMulti(
			"F", []string{"foo.bar@60@120"}, gomock.Any()).
			Return(nil),
	)

	p, err := getFromFrames(
		mockae, "foo.bar", FrameSize(60), 65, 164)

	if err != nil {
		t.Errorf("Did not expect an error: %s", err.Error())
	}
	if len(p) != 2 {
		t.Errorf("Expected 2 points, got %d", len(p))
	}
	if p[0].V != 43.81 {
		t.Errorf("Wrong value for p0: got %f", p[0].V)
	}
	if p[1].V != 80 {
		t.Errorf("Wrong value for p1: got %f", p[1].V)
	}
}
*/
