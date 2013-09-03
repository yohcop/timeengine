package points

type StatSelector func(StatsDataPoint) float64

func GetSummarySelector(fn string) StatSelector {
  switch fn {
    case "avg": return avgSelector
    case "sum": return sumSelector
    case "min": return minSelector
    case "max": return maxSelector
  }
  return avgSelector
}

func avgSelector(s StatsDataPoint) float64 {
  return s.GetAvg()
}

func sumSelector(s StatsDataPoint) float64 {
  return s.GetSum()
}

func minSelector(s StatsDataPoint) float64 {
  return s.GetMin()
}

func maxSelector(s StatsDataPoint) float64 {
  return s.GetMax()
}
