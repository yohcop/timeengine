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

func makeTestSummary(ss, sk int64, p1, p2 int64, v1, v2 float64) *summary {
	return newSummary("foo.bar", SummarySize(ss), SummaryKey(sk),
		[]StatsDataPoint{
			newStats(p1).update(v1, v1, v1, v1),
			newStats(p2).update(v2, v2, v2, v2),
		},
	)
}

func TestBuildSummaries(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p1 := makeTestSummary(60, 3600, 3600, 3610, 1.1, 1.2)
	p2 := makeTestSummary(60, 7260, 7264, 7268, 2.1, 2.2)

	// The limits between which the search is going to happen.
	l1 := summaryKeyName("foo.bar", SummarySize(60), SummaryKey(3600))
	l2 := summaryKeyName("foo.bar", SummarySize(60), SummaryKey(10740))

	mockae := mock_ae.NewMockContext(mockCtrl)
	gomock.InOrder(
		mockae.EXPECT().DsGetBetweenKeys(summaryDatastoreType, l1, l2, -1, gomock.Any()).
			SetArg(4, []*summary{p1, p2}).
			Return([]string{p1.Key(), p2.Key()}, nil),
		mockae.EXPECT().PutMulti(
			summaryDatastoreType, []string{
				summaryKeyName("foo.bar", SummarySize(3600), SummaryKey(3600)),
				summaryKeyName("foo.bar", SummarySize(3600), SummaryKey(7200)),
			}, gomock.Any()).
			Return(nil),
	)

	log.Println("=====")
	summaries, err := BuildSummaries(
		mockae, "foo.bar", NewSpan(SelectSummarySize(3600), 3600, 7200))

	if err != nil {
		t.Fatalf("Error? %s", err.Error())
	}
	if expected, got := 2, len(summaries); expected != got {
		t.Fatalf("Wrong number of points: %d, expected %d", got, expected)
	}
	if expected, got := 1.15, summaries[0].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
	if expected, got := 2.15, summaries[1].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
}

func TestBuildSummariesFirstLevel(t *testing.T) {
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
			summaryDatastoreType, []string{
				summaryKeyName("foo.bar", SummarySize(60), SummaryKey(60)),
			}, gomock.Any()).
			Return(nil),
	)

	log.Println("=====")
	summaries, err := BuildSummaries(
		mockae, "foo.bar", NewSpan(SelectSummarySize(60), 60, 60))

	if err != nil {
		t.Fatalf("Error? %s", err.Error())
	}
	if expected, got := 1, len(summaries); expected != got {
		t.Fatalf("Wrong number of points: %d, expected %d", got, expected)
	}
	if expected, got := 1.15, summaries[0].GetAvg(); math.Abs(expected-got) > epsilon {
		t.Errorf("Wrong value: %f, expected %f", got, expected)
	}
}
