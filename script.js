var wall = {}
var altwall = {}

function addattr(element) {
	element.setAttribute("onmouseover", "mouseover(this)");
	element.setAttribute("onmouseout", "mouseout(this)");
	element.setAttribute("onclick", "mouseclick(this)");
}

function start() {var tb = document.getElementById("board");
	for (i = 0; i < 5; i++) {
		var tr = document.createElement("tr");
		for (j = 0; j < 5; j++) {

			// room
			var td = document.createElement("td");
			td.className = "room";
			td.id = "room" + String(i) + String(j);
			addattr(td);
			tr.appendChild(td);

			if (j == 4) break;

			// vcorr
			var td = document.createElement("td");
			td.className = "vcorr";
			td.id = "corr" + String(i) + String(j) + ":" + String(i) + String(j + 1);
			addattr(td);
			if (i > 0) {
				altwall[td.id] = [
					"corr" + String(i - 1) + String(j) + ":" + String(i - 1) + String(j + 1),
					"pole" + String(i - 1) + String(j),
					td.id
				];
			}
			if (i < 4) {
				wall[td.id] = [
					td.id,
					"pole" + String(i) + String(j),
					"corr" + String(i + 1) + String(j) + ":" + String(i + 1) + String(j + 1)
				];
			}
			tr.appendChild(td);

		}

		tb.appendChild(tr);
		if (i == 4) break;
		var tr = document.createElement("tr");
		for (j = 0; j < 5; j++) {

			// hcorr
			var td = document.createElement("td");
			td.className = "hcorr";
			td.id = "corr" + String(i) + String(j) + ":" + String(i + 1) + String(j);
			addattr(td);
			if (j > 0) {
				altwall[td.id] = [
					"corr" + String(i) + String(j - 1) + ":" + String(i + 1) + String(j - 1),
					"pole" + String(i) + String(j - 1),
					td.id
				];
			}
			if (j < 4) {
				wall[td.id] = [
					td.id,
					"pole" + String(i) + String(j),
					"corr" + String(i) + String(j + 1) + ":" + String(i + 1) + String(j + 1)
				];
			}
			tr.appendChild(td);

			if (j == 4) break;

			// pole
			var td = document.createElement("td");
			td.className = "pole";
			td.id = "pole" + String(i) + String(j);
			addattr(td);
			wall[td.id] = [
				"corr" + String(i) + String(j) + ":" + String(i) + String(j + 1),
				td.id,
				"corr" + String(i + 1) + String(j) + ":" + String(i + 1) + String(j + 1)
			];
			altwall[td.id] = [
				"corr" + String(i) + String(j) + ":" + String(i + 1) + String(j),
				td.id,
				"corr" + String(i) + String(j + 1) + ":" + String(i + 1) + String(j + 1)
			];
			tr.appendChild(td);
		}
		tb.appendChild(tr);
	}
}

function isbuilt(wallparts) {
	for (i = 0; i < wallparts.length; i++) {
		if (document.getElementById(wallparts[i]).classList.contains("built")) {
			return true;
		}
	}
	return false;
}

function mouseclick(element) {
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		wall[element.id].forEach( function(value) { document.getElementById(value).classList.add("built"); });
		drawwall(wall[element.id], "#444444");
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		altwall[element.id].forEach( function(value) { document.getElementById(value).classList.add("built"); });
		drawwall(altwall[element.id], "#444444");
	}
}

function mouseover(element) {
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		drawwall(wall[element.id], "#cccccc");
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		drawwall(altwall[element.id], "#cccccc");
	}
}

function mouseout(element) {
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		drawwall(wall[element.id], "#ffffff");
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		drawwall(altwall[element.id], "#ffffff");
	}
}

function drawwall(wallparts, color) {
	wallparts.forEach( function(value) { setbg(document.getElementById(value), color); });
}

function setbg(element, color) {
	element.style.backgroundColor = color;
}
