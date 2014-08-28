
function listNamespaces(cb, errorcb) {
  $.ajax({
    url: "/api/namespace/list",
  	dataType: 'json',
  	success: cb,
    error: errorcb,
  });
}

function listUsers(cb, errorcb) {
  $.ajax({
    url: "/api/user/list",
  	dataType: 'json',
  	success: cb,
    error: errorcb,
  });
}
