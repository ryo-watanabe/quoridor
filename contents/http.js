function request_quoridor_data(req)
{
	http = new XMLHttpRequest();
	http.open('POST', '/api/', true);
	http.setRequestHeader('Content-Type', 'application/json');
	http.setRequestHeader('Authorization', 'Basic ' + '[base64 user:pass]');
	http.onreadystatechange = function() {
		if (http.readyState == 4) {
			console.log("Status:" + http.status);
			//console.log(http.responseText);
			if (http.status == 200) {
				quoridor_data(JSON.parse(http.responseText))
			}
		}
	}
	http.send(JSON.stringify(req));
}