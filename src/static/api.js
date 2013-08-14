
function listNamespaces(cb, errorcb) {
  $.ajax({
    url: "/api/namespace/list/",
  	dataType: 'json',
  	success: cb,
    error: errorcb,
  });
}
