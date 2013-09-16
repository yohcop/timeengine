package timeseries

// General algorithm to get datapoints:
// - The request only contains a resolution, start and end points.
// - The response resolution must be the requested one or finer.
//
// The general storage is organized in 2 tiers:
// - the "raw" tier, that contains raw data. This is never modified
//   after it's written, and contains data to higher resolution (ms)
// - an "summarized" tier, that is made of summaries. each summary contains
//   a list of points from the raw tier or a list of points from finer
//   resolution summaries.
//
// It is cheaper to get a bunch of points from the summarized tier than
// from the raw tier. For example, for 1 minute of data, we need 60 or
// more items from the raw tier, or 1 from the summarized tier.
//
// if the resolution less than 1 second, or 'raw' (should be very rare)
// get points from the raw tier. Even the graphs probably don't care
// about that currently.
//
// Otherwise, if the resolution is 1 second or bigger, try to use summaries:
// - GetFromSummary - this is the only exported function
//   - getFromWiderSummary:
//     compute the summary boundaries at the coarser resolution
//     (say 60 seconds, if we try to get 1 second resolution)
//     - try to get those from the datastore.
//     - for those we get, use the list of points they have.
//     - ignore the missing summaries.
//     - call getFromSummaries with the boundaries of the missing summaries.
//   - getFromSummaries:
//     - if the resolution is less than 1 second, get and return
//       points from the raw tier. stop here. otherwise:
//     - compute the summary boundaries, and try to get those summaries. Use a
//       projection query to only retrieve their value, and not all the
//       points. Use the average values from those summaries.
//     - for the missing summaries (this can be done in parallel):
//       - get the summaries of smaller sizes, covered by the summary we
//         are trying to generate, by calling getFromSummaries.
//       - create a missing summary out of these.
//       - store the new summary.
