
// Default options. These can be specified in the URL as well.
var opts = {
  graphite_url: '',

  namespace: '',
  dashboard: 'ycoppel1',
  test_metric: '',

  auto_fetch: true,
  sync_graphs: true,
  live_updates: true,
  initial_fetch: '-10s',

  graph_width: '640',
  graph_height: '300',

  // Preloading
  from: 0,
  to: 0,
  full_res_preload: false,

  // Params for dashboard config.
  params: {},
};

// State stuff.
var first = Math.ceil(new Date().valueOf() / 1000);
var last = Math.ceil(new Date().valueOf() / 1000);
var min_freq = 10;
// Dygraph -> object{g, targets, data}
var graphs = [];
// target -> aggregate function. target is of the form
// "target_name@aggregate_fn", e.g. "my.metric@avg"
// This is use for fetching, so we don't fetch the same
// metrics multiple times if they are used in different
// graphs.
var all_targets = {};
var timer = null;
var fetchTimer = null;
var blockRedraw = false;

function drawCallback(me, initial) {
  if (blockRedraw || initial || !opts.sync_graphs) return;
  blockRedraw = true;
  var range = me.xAxisRange();
  for (var gi in graphs) {
    var g = graphs[gi];
    if (g.g == me) continue;
    g.g.updateOptions({
      dateWindow: range,
    });
  }
  blockRedraw = false;
  fetchOnMoveTimer();
}

function replaceParamsInTargets(targets) {
  for (var t in targets) {
    var s = targets[t];
    var match = null;
    while (match = s.match(/\$\{(.+?)\}/)) {
      s = s.replace(match[0], opts.params['$' + match[1]]);
      console.log(s);
    }
    targets[t] = s;
  }
}

function mkgraph(els, targets, expressions, title, dygraphOpts) {
  replaceParamsInTargets(targets);

  var parsedTargets = {};
  var labels = ['x'];
  for (var t in targets) {
    var target = targets[t];
    var aggregate = 'avg';
    var name = target;
    var encodedTarget = target;

    var targetCfg = name.split('@');
    if (targetCfg.length == 2) {
      name = targetCfg[0];
      aggregate = targetCfg[1];
    } else {
      encodedTarget = target + '@' + aggregate;
    }
    labels.push(t);
    parsedTargets[t] = encodedTarget;
    all_targets[encodedTarget] = {name: name, fn: aggregate};
  }
  var parsedExpressions = null;
  if (expressions) {
    parsedExpressions = [];
    labels = ['x'];
    for (var exp_i in expressions) {
      labels.push(exp_i);
      parsedExpressions.push(Parser.parse(expressions[exp_i]));
    }
  }
  dygraphCfg = {
    labels: labels,
    animatedZooms: true,
    panEdgeFraction: 0,
    drawCallback: drawCallback,
    legend: 'always',
    labelsDiv: els.legend,
    title: title,
    width: opts.graph_width,
    height: opts.graph_height,
  };
  $.extend(dygraphCfg, dygraphOpts);
  var g = new Dygraph(els.graph, [], dygraphCfg);
  graphs.push({
    g: g,
    targets: parsedTargets,
    expressions: parsedExpressions,
    data: [],
    data_by_date: {},
  });
}

function extractNameAndAggregateFn(metric) {
}

function pollUrl(from, to, summarize) {
  if (Object.keys(all_targets).length == 0) {
    return null;
  }
  var left = 0;
  if (from) {
    left = from;
  } else {
    left = last - min_freq;
  }
  var maybe_to = '';
  if (to != undefined) {
    maybe_to = '&until=' + to;
  }

  var targets_q = "";
  // We save here the targets we already requested. There may
  // be some duplicates, since all_targets key includes the aggregate
  // function, but when we request full data, we don't use this.
  var done_targets = {};
  for (var k in all_targets) {
    name = all_targets[k].name;
    name = addNamespaceToTarget(name);
    if (!summarize) {
      if (!done_targets[name]) {
        targets_q += "&target=" + name;
        done_targets[name] = true;
      }
    } else {
      fn = all_targets[k].fn;
      targets_q += "&target=" +
        encodeURIComponent(
            "summarize(" + name + ", \"" +
                       summarize + "\", \"" +
                       fn + "\")");
    }
  }
  return opts.graphite_url + '/render/?' +
    'from=' + left + maybe_to +
    targets_q + '&drawNullAsZero=false&noCache=true&format=json&jsonp=?';
}

function update(url) {
  // TODO: wrap update in another function that calls setupUpdates
  // correctly, instead of the hack of looking if url is defined.
  var outOfBand = (url != undefined);
  if (outOfBand) {
    // Only show the loading sign for 'out of band' (i.e. not on timer)
    // updates, otherwise, it's annoying.
    loading(true);
  }
  timer = null;
  var start_update = new Date().valueOf();
  var bye = function() {
    if (outOfBand) {
      loading(false);
      return;
    }
    var end_update = new Date().valueOf();
    setupUpdates(1000 - (end_update - start_update));
  }

  if (!url) url = pollUrl();
  if (!url) {
    bye();
    return;
  }
  $.ajax({
    url: url,
		dataType: 'jsonp',
		success: function(d) {
      var prev_last = last;
      // Extract all the data by date
      var new_data = {};
      for (var i = 0; i < d.length; ++i) {
        var series = d[i];
        var name = series.target;
        var is_summary = name.match(/summarize\(([\w.-]*), .*\)/);
        if (is_summary) {
          name = is_summary[1];
        }
        name = removeNamespaceFromTarget(name);
        var points = series.datapoints;
        var ny = [];
        for (var pi = 0; pi < points.length; ++pi) {
          var ts = points[pi][1];
          var val = points[pi][0];
          ny.push([ts, val]);
          if (ts > last) {
            last = ts;
          }
          if (ts < first) {
            first = ts;
          }
        }
        new_data[name] = ny;
      }

      rebuildGraphs(new_data, prev_last);
      bye();
    },
    error: function(e) {
      console.log("ERROR");
      console.log(e);
      bye();
    },
  });
}

function executeExpr(vars, expr) {
  return expr.evaluate(vars);
}

function processData(targets, rawData, expressions) {
  if (!expressions) {
    return rawData;
  }
  // rawData is an array, where each element is:
  // [date, value_for_target0, ..., value_for_targetN]
  // Expressions reffer to values via v0, v1, ... vN.
  var result = [];
  for (var row_i in rawData) {
    var row = rawData[row_i];
    var vars = {};
    var i = 1;
    for (var target in targets) {
      vars[target] = row[i++];
    }

    var newRow = [row[0]];  // Just the date.
    for (var e in expressions) {
      newRow.push(executeExpr(vars, expressions[e]));
    }
    result.push(newRow);
  }
  return result;
}

// new_data is series_name -> ts -> val
function rebuildGraphs(new_data, append_from) {
  for (var gi in graphs) {
    var obj = graphs[gi];
    var g = obj.g;
    var targets = obj.targets;

    // Update the data by date for this graph.
    // ts -> name -> val
    var data_by_date = {};
    for (var ti in targets) {
      var name = targets[ti];
      var series = new_data[name];
      for (var si in series) {
        var ts = series[si][0];
        var val = series[si][1];
        var x = data_by_date[ts];
        if (!x) { x = {}; }
        x[name] = val;
        data_by_date[ts] = x;
      }
    }

    //console.log('now building final array from', data_by_date);
    // Build the list of points from the dictionary.
    graph_data = [];
    for (var ts in data_by_date) {
      var entry = [ new Date(ts * 1000) ];
      for (var n in targets) {
        entry.push(data_by_date[ts][targets[n]]);
      }
      setInDataArray(graph_data, entry[0], entry);
    }
    if (graph_data.length != 0) {
      var start_replace = findSplitPoint(obj.data, graph_data[0][0]);
      var end_replace = findSplitPoint(obj.data, graph_data[graph_data.length - 1][0]);
      obj.data =
        obj.data.slice(0, start_replace).concat(
            graph_data).concat(obj.data.slice(end_replace + 1));
    }

    // Update the graph.
    var opts = {
      file: processData(obj.targets, obj.data, obj.expressions),
    };
    var win = g.xAxisRange();
    var win_to_last = win[1] - (append_from * 1000);
    var following = g.isZoomed('x') && win_to_last > -5000;//< 1000 && win_to_last > -1000;
    if (following) {
      if (win_to_last < 0) win_to_last = 0;
      // Move the window.
      var head = last * 1000 + win_to_last;
      opts['dateWindow'] = [head - win[1] + win[0], head];
    }
    blockRedraw = true;
    g.updateOptions(opts);
    blockRedraw = false;
  }
}

function findSplitPoint(ar, ts) {
  function findTsInternal(ar, start, end) {
    var pivot = Math.floor(start + (end - start) / 2);
    if (end == start) {
      return pivot;
    }
    if(end-start == 1) {
      if (ar[start] && ar[start][0].valueOf() == ts.valueOf()) {
        return start;
      }
      if (ar[end] && ar[end][0].valueOf() == ts.valueOf()) {
        return end;
      }
      return pivot;
    } else if(ar[pivot][0].valueOf() < ts.valueOf()) {
      return findTsInternal(ar, pivot, end);
    } else {
      return findTsInternal(ar, start, pivot);
    }
  }
  return findTsInternal(ar, 0, ar.length);
}

function setInDataArray(ar, ts, ts_val) {
  // First, shortcuts, as these are 2 very common operations:
  if (ar[ar.length-1] &&
      ar[ar.length-1][0].valueOf() < ts_val[0].valueOf()) {
    ar.push(ts_val);
    return;
  }
  if (ar[0] && ar[0][0].valueOf() > ts_val[0].valueOf()) {
    ar.unshift(ts_val);
    return;
  }

  function cp(from, to) {
    for (var i in from) {
      to[i] = from[i];
    }
  }
  function findTs(ar, start, end) {
    var pivot = Math.floor(start + (end - start) / 2);
    if(end-start <= 1) {
      if (ar[pivot + 1]) {
        if (ar[pivot + 1][0].valueOf() == ts.valueOf()) {
          cp(ts_val, ar[pivot + 1]);
          return ar;
        }
      }
      ar.splice(pivot + 1, 0, ts_val);
      return ar;
    }
    if (ar[pivot][0] == ts) {
      ar[pivot] = ts_val;
      return ar;
    } else if(ar[pivot][0] < ts) {
      return findTs(ar, pivot, end);
    } else {
      return findTs(ar, start, pivot);
    }
  }

  return findTs(ar, 0, ar.length);
}

function toggleUpdates() {
  var c = document.getElementById('ct-update');
  opts.live_updates = c.checked;
  setupUpdates();
}

function togglePullOnPan() {
  var c = document.getElementById('ct-pull-on-pan');
  opts.auto_fetch = c.checked;
}

function setupUpdates(ms) {
  if (opts.live_updates) {
    if (!timer) {
      ms = ms != undefined ? ms : 1000;
      if (ms < 100) {
        // Still use a timout, so we don't grow the call stack
        // constantly, and dont' overload the server.
        ms = 100
      }
      timer = setTimeout(update, ms);
    }
  } else if (timer) {
    console.log("unset");
    clearTimeout(timer);
    timer = null;
  }
}


function fetchOnMoveTimer() {
  if (fetchTimer) {
    clearTimeout(fetchTimer);
  }
  fetchTimer = setTimeout(function() {
    // TODO: check what data we already have here.
    if (opts.auto_fetch) {
      loadFromZoom(false);
    }
  }, 1000);
}

function toggleSync() {
  var c = document.getElementById('ct-sync-graphs');
  opts.sync_graphs = c.checked;
  if (opts.sync_graphs) {
    drawCallback(graphs[0].g, false);
  }
}

function setLast(secs) {
  var now = new Date().valueOf();
  var left = (now - secs*1000);
  setDateWindow(left, now - 1000);
}

// left and right in milliseconds.
function setDateWindow(left, right) {
  // Stop the redraw callback and update them all.
  blockRedraw = true;
  for (var gi in graphs) {
    var obj = graphs[gi];
    var g = obj.g;
    g.updateOptions({'dateWindow': [left, right]});
  }
  blockRedraw = false;
}

function loadMore(secs, res) {
  console.log(first);
  var url = pollUrl(first - secs, first, res);
  console.log(url);
  update(url);
}

// from and to are in milliseconds.
function findGoodSummary(from, to, full_res) {
  var g = graphs[0].g;
  var pixels = g.getArea().w;
  var secs = (to - from) / 1000;
  var sec_per_pixel = Math.floor(secs / pixels);
  var url = "";
  if (sec_per_pixel <= 1 || full_res) {
    url = pollUrl(
        Math.floor(from / 1000),
        Math.floor(to / 1000));
  } else {
    url = pollUrl(
        Math.floor(from / 1000),
        Math.floor(to / 1000),
        sec_per_pixel + 's');
  }
  return url;
}

function loadFromZoom(full_res) {
  var g = graphs[0].g;
  var range = g.xAxisRange();
  var url = findGoodSummary(
      range[0], range[1], full_res);
  update(url);
}

function createChartEl() {
  var container = document.createElement('div');
  var graph_el = document.createElement('div');
  var legend = document.createElement('div');
  container.className = 'graph-container';
  graph_el.className = 'graph';
  legend.className = 'legend';
  legend.style.width=""+opts.graph_width+"px";
  container.appendChild(graph_el);
  container.appendChild(legend);
  document.getElementById('graphs').appendChild(container);
  return {container: container, graph: graph_el, legend: legend};
}

function parseOpts() {
  var path = location.hash.slice(1);
  var dash_opts = path.split('&');
  opts.dashboard = dash_opts[0];
  for (var i = 1; i < dash_opts.length; ++i) {
    var kv = dash_opts[i].split('=');
    var key = kv[0];
    var val = true;
    if (kv.length > 0) {
      val = kv.slice(1).join('=');
      if (val == 'true') val = true;
      else if (val == 'false') val = false;
    }
    if (key[0] == '$') {
      opts.params[key] = val;
    } else {
      opts[key] = val;
    }
  }
  console.log(opts);

  document.getElementById('ct-update').checked = opts.live_updates;
  setupUpdates();
  document.getElementById('ct-sync-graphs').checked = opts.sync_graphs;
  document.getElementById('ct-pull-on-pan').checked = opts.auto_fetch;
}

function setupTestDashboard() {
  document.getElementById('title').textContent = '#TEST#';
  var targets = [
      opts.test_metric,
  ];
  var els = createChartEl();
  mkgraph(els, targets, null, "test", {});
  finishSetup();
}

function setupDashboard() {
  document.getElementById('title').textContent = '#' + opts.dashboard;
  $.ajax({
    url: opts.graphite_url + "/api/dashboard/get?dashboard=" + opts.dashboard,
  	dataType: 'json',
  	success: function(d) {
      for (var gi = 0; gi < d.graphs.length; ++gi) {
        //if (gi != 1) continue;
        var cfg = d.graphs[gi];
        var els = createChartEl();
        mkgraph(els, cfg.targets, cfg.expressions, cfg.title,
          cfg.dygraphOpts);
      }
      finishSetup();
    },
    error: function(e) {
      console.log("error", e);
    }
  });
}

function addNamespaceToTarget(t) {
  if (opts.namespace.length > 0) {
    return opts.namespace + '*' + t;
  }
  return t;
}

function removeNamespaceFromTarget(t) {
  if (opts.namespace.length > 0) {
    return t.slice(opts.namespace.length + 1);
  }
  return t;
}

// from and to in milliseconds
function loadDates(from, to, full_res) {
  var url = findGoodSummary(from, to, full_res);
  console.log(from, new Date(from));
  console.log(to, new Date(to));

  setDateWindow(from, to);
  update(url);
}

function finishSetup() {
  if (opts.from && opts.to) {
    loadDates(
        new Date(opts.from).valueOf(),
        new Date(opts.to).valueOf(),
        opts.full_res_preload);
  } else {
    update();
  }
  toggleUpdates();
}

function shareView() {
  var url = shareUrl();
  document.location.hash = url;
  alert('Copy the URL. By default, this will load the timestamp ' +
      'at full resolution. If your time range is too wide, you ' +
      'may consider editing the URL before sharing it.');
}

function shareUrl() {
  var g = graphs[0].g;
  var range = g.xAxisRange();
  console.log(range);
  var from = new Date(range[0]).toISOString();
  var to = new Date(range[1]).toISOString();
  console.log(from, to);
  var url = '#' + opts.dashboard;
  for (var k in opts) {
    if (k == 'dashboard' || k == 'params') {
      continue;
    }
    url += '&' + k + '=' + opts[k];
  }
  for (var p in opts.params) {
    url += '&' + p + '=' + opts.params[p];
  }
  return url;
}

function loading(isLoading) {
  console.log(isLoading);
  if (isLoading) {
    $('#loading').show();
  } else {
    $('#loading').hide();
  }
}
