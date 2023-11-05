function request_quoridor_data(req)
{
	document.getElementById("status").innerHTML = "Wait";
	document.getElementById("message").innerHTML = "Sending request... <img src='loading_grn.gif'>";
	http = new XMLHttpRequest();
	http.open('POST', '/api/', true);
	http.setRequestHeader('Content-Type', 'application/json');
	http.setRequestHeader('Authorization', 'Basic ' + '[set base64 user:pass]');
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
