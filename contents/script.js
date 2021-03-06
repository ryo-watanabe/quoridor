var dim = 7;
var wall = {};
var altwall = {};
var room = {};
var corr = [];
var you = {"nowon":"42", "move":[]};
var com = {"nowon":"02", "move":[]};
var board = {};
var undoboard = null;
var gameinit = false;
var playable = true;
var manual = false;
var currentplayer = "1";
var currentcom = "2";

function addattr(element) {
	element.setAttribute("onmouseover", "mouseover(this)");
	element.setAttribute("onmouseout", "mouseout(this)");
	element.setAttribute("ontouchstart", "mouseover(this)");
	element.setAttribute("ontouchend", "mouseout(this)");
	element.setAttribute("onclick", "mouseclick(this)");
}

function roomaddattr(element) {
	//element.setAttribute("onmouseover", "roommouseover(this)");
	//element.setAttribute("onmouseout", "roommouseout(this)");
	element.setAttribute("onclick", "roommouseclick(this)");
}

function quoridor_data(res)
{
	console.log(res)
	document.getElementById("status").innerHTML = res.status;
	if (res.message) {
		document.getElementById("message").innerHTML = res.message;
	} else {
		document.getElementById("message").innerHTML = "";
	}
	if (res.evaluation) {
		document.getElementById("eval").innerHTML = "eval:" + res.evaluation.eval + "%";
		// " / player eval:" + res.evaluation.nextPlayerEval +
		// " / route calc:" + res.evaluation.numCases;
		// " / player cases:" + res.evaluation.numNextPlayerCases;
	} else {
		document.getElementById("eval").innerHTML = "";
	}
	document.getElementById("undobtn").disabled = (undoboard == null);
	if (res.status == "NG") return;
	// poles
	if (res.board.poles) {
		for (var i = 0; i < res.board.poles.length; i++) {
			buildpole(res.board.poles[i]);
		}
	} else {
		res.board.poles = [];
	}
	// blockings
	if (res.board.blockings) {
		for (var i = 0; i < res.board.blockings.length; i++) {
			buildblock(res.board.blockings[i]);
		}
	} else {
		res.board.blockings = [];
	}
	// walls left
	if (!manual) {
		document.getElementById("walls").innerHTML = "Walls left com:" + res.board.comWalls + " player:" + res.board.playerWalls;
	}

	if (res.status == "COM") {
		for (j = 0; j < dim; j++) {
			roomid = String(dim - 1) + String(j);
			setbg(document.getElementById(roomid), "#ddaadd");
		}
	}
	if (res.status == "PLY") {
		var goal = 0;
		if (manual) {
			if (currentplayer == "2") goal = dim - 1;
			document.getElementById("message").innerHTML = "Player" + currentplayer + " won";
		}
		for (j = 0; j < dim; j++) {
			roomid = String(goal) + String(j);
			setbg(document.getElementById(roomid), "#aadddd");
		}
	}

	if (manual && res.status == "OK") {
		currentcom = currentplayer;
		if (currentplayer == "1") {
			currentplayer = "2";
			document.getElementById("walls").innerHTML = "Walls left P1:" + res.board.comWalls + " P2:" + res.board.playerWalls;
			document.getElementById("eval").innerHTML = "eval(P1:P2) " + res.evaluation.eval + "%:" + (100 - res.evaluation.eval) + "%";
		} else {
			currentplayer = "1";
			document.getElementById("walls").innerHTML = "Walls left P1:" + res.board.playerWalls + " P2:" + res.board.comWalls;
			document.getElementById("eval").innerHTML = "eval(P1:P2) " + (100 - res.evaluation.eval) + "%:" + res.evaluation.eval + "%";
		}
		document.getElementById("message").innerHTML = "Player" + currentplayer;
	}

	updateyou(String(res.board.playerPos.y) + String(res.board.playerPos.x));
	updatecom(String(res.board.comPos.y) + String(res.board.comPos.x));
	updatemove();
	board = res.board;
	if (res.status == "OK")	playable = true;

	document.getElementById("comfirstbtn").disabled = !gameinit;
	document.getElementById("manualbtn").disabled = !gameinit;
}

function comfirst() {
	if (!gameinit) return;
	gameinit = false;
	request_quoridor_data({action:"Com", board:board});
}

function twoplayer() {
	if (!gameinit) return;
	manual = true;
	currentplayer = "1";
	currentcom = "2";
	document.getElementById("manualbtn").disabled = true;
	document.getElementById("message").innerHTML = "Player1";
	document.getElementById(you["nowon"]).innerHTML = "P1";
	document.getElementById(com["nowon"]).innerHTML = "P2";
	document.getElementById("walls").innerHTML = "Walls left P1:" + board.playerWalls + " P2:" + board.comWalls;
}

function undocopy() {
	var str = JSON.stringify(board);
	undoboard = JSON.parse(str);
}

function undo() {

	if (undoboard == null) return;

	draw_board();
	// poles
	for (var i = 0; i < undoboard.poles.length; i++) {
		buildpole(undoboard.poles[i]);
	}
	// blockings
	for (var i = 0; i < undoboard.blockings.length; i++) {
		buildblock(undoboard.blockings[i]);
	}
	// walls left
	document.getElementById("walls").innerHTML = "Walls left com:" + undoboard.comWalls + " player:" + undoboard.playerWalls;

	if (manual) {
		if (playable) {
			currentcom = currentplayer;
			if (currentplayer == "1") {
				currentplayer = "2";
				document.getElementById("walls").innerHTML = "Walls left P1:" + undoboard.comWalls + " P2:" + undoboard.playerWalls;
			} else {
				currentplayer = "1";
				document.getElementById("walls").innerHTML = "Walls left P1:" + undoboard.playerWalls + " P2:" + undoboard.comWalls;
			}
		}
		document.getElementById("message").innerHTML = "Player" + currentplayer;
	}
	updateyou(String(undoboard.playerPos.y) + String(undoboard.playerPos.x));
	updatecom(String(undoboard.comPos.y) + String(undoboard.comPos.x));
	updatemove();
	playable = true;
	board = undoboard;
	undoboard = null;
	document.getElementById("undobtn").disabled = (undoboard == null);
}

function newgame(dimension) {
	dim = dimension;
	start();
}

function start() {
	gameinit = true;
	manual = false;
	draw_board();
	request_quoridor_data({action:"Init", board:{dimension:dim}});
}

function draw_board() {
	var tb = document.getElementById("board");

	// clear game board
	while (tb.firstChild) tb.removeChild(tb.firstChild);
	wall = {};
	altwall = {};
	room = {};
	corr = [];
	you = {"nowon":"42", "move":[]};
	com = {"nowon":"02", "move":[]};

	for (i = 0; i < dim; i++) {
		var tr = document.createElement("tr");
		for (j = 0; j < dim; j++) {

			// room
			var td = document.createElement("td");
			td.className = "room";
			td.id = String(i) + String(j);
			roomaddattr(td);
			room[td.id] = []
			if (i > 0) {
				room[td.id].push(String(i - 1) + String(j));
			}
			if (j > 0) {
				room[td.id].push(String(i) + String(j - 1));
			}
			if (i < dim - 1) {
				room[td.id].push(String(i + 1) + String(j));
			}
			if (j < dim - 1) {
				room[td.id].push(String(i) + String(j + 1));
			}
			tr.appendChild(td);

			if (j == dim - 1) break;

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
			if (i < dim - 1) {
				wall[td.id] = [
					td.id,
					"pole" + String(i) + String(j),
					"corr" + String(i + 1) + String(j) + ":" + String(i + 1) + String(j + 1)
				];
			}
			corr.push(td.id);
			tr.appendChild(td);

		}

		tb.appendChild(tr);

		if (i == dim - 1) break;

		var tr = document.createElement("tr");
		for (j = 0; j < dim; j++) {

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
			if (j < dim - 1) {
				wall[td.id] = [
					td.id,
					"pole" + String(i) + String(j),
					"corr" + String(i) + String(j + 1) + ":" + String(i + 1) + String(j + 1)
				];
			}
			corr.push(td.id);
			tr.appendChild(td);

			if (j == dim - 1) break;

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
	//updatemove()
}

function setbg(element, color) {
	element.style.backgroundColor = color;
}

// Wall UI functions

function isbuilt(wallparts) {
	for (i = 0; i < wallparts.length; i++) {
		if (document.getElementById(wallparts[i]).classList.contains("built")) {
			return true;
		}
	}
	return false;
}

function mouseclick(element) {
	if (!playable) return;
	if (board.playerWalls <= 0) return;
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		wall[element.id].forEach( function(value) { document.getElementById(value).classList.add("built"); });
		drawwall(wall[element.id], "#446644");
		writewall(wall[element.id]);
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		altwall[element.id].forEach( function(value) { document.getElementById(value).classList.add("built"); });
		drawwall(altwall[element.id], "#446644");
		writewall(altwall[element.id]);
	}
}

function writewall(wallparts) {
	undocopy();
	gameinit = false;
	wallparts.forEach( function(val) {
		if (val.startsWith("pole")) {
			board.poles.push({y:parseInt(val.substr(4, 1)), x:parseInt(val.substr(5,1))});
		}
		if (val.startsWith("corr")) {
			board.blockings.push([{y:parseInt(val.substr(4, 1)), x:parseInt(val.substr(5,1))}, {y:parseInt(val.substr(7, 1)), x:parseInt(val.substr(8,1))}]);
		}
	});
	board.playerWalls--;
	playable = false;
	document.getElementById("undobtn").disabled = true;

	var action = "Com";
	if (manual) action = "Man";
	var calc = parseInt(document.getElementById("calc").value)
	request_quoridor_data({action:action, board:board, calc:calc});
}

function mouseover(element) {
	if (!playable) return;
	if (board.playerWalls <= 0) return
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		drawwall(wall[element.id], "#aaccaa");
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		drawwall(altwall[element.id], "#aaccaa");
	}
}

function mouseout(element) {
	if (!playable) return;
	if (board.playerWalls <= 0) return
	if (wall[element.id] && !isbuilt(wall[element.id])) {
		drawwall(wall[element.id], "#ffffff00");
	} else if (altwall[element.id] && !isbuilt(altwall[element.id])) {
		drawwall(altwall[element.id], "#ffffff00");
	}
}

function drawwall(wallparts, color) {
	wallparts.forEach( function(value) { setbg(document.getElementById(value), color); });
}

function buildpole(pole) {
	poleid = "pole" + String(pole.y) + String(pole.x);
	document.getElementById(poleid).classList.add("built");
	setbg(document.getElementById(poleid), "#446644");
}

function buildblock(block) {
	blockid = "corr" + String(block[0].y) + String(block[0].x) + ":" + String(block[1].y) + String(block[1].x);
	document.getElementById(blockid).classList.add("built");
	setbg(document.getElementById(blockid), "#446644");
}

function unbuildpole(pole) {
	poleid = "pole" + String(pole.y) + String(pole.x);
	document.getElementById(poleid).classList.remove("built");
	setbg(document.getElementById(poleid), "#ffffff00");
}

function unbuildblock(block) {
	blockid = "corr" + String(block[0].y) + String(block[0].x) + ":" + String(block[1].y) + String(block[1].x);
	document.getElementById(blockid).classList.remove("built");
	setbg(document.getElementById(blockid), "#ffffff00");
}

// Room UI functions

function updateyou(roomid) {
	document.getElementById(you["nowon"]).innerHTML = "";
	document.getElementById(you["nowon"]).classList.remove("player");
	document.getElementById(roomid).innerHTML = "Ply";
	if (manual) {
		document.getElementById(roomid).innerHTML = "P" + currentplayer;
	}
	document.getElementById(roomid).classList.add("player");
	you["nowon"] = roomid;
}

function updatecom(roomid) {
	if (!manual) {
		document.getElementById(com["nowon"]).innerHTML = "";
	}
	document.getElementById(com["nowon"]).classList.remove("com");
	document.getElementById(roomid).innerHTML = "Com";
	if (manual) {
		document.getElementById(roomid).innerHTML = "P" + currentcom;
	}
	document.getElementById(roomid).classList.add("com");
	com["nowon"] = roomid;
}

function updatemove() {
	// update you
	for (i = 0; i < you["move"].length; i++) {
		document.getElementById(you["move"][i]).classList.remove("move");
	}
	you["move"] = [];
	for (i = 0; i < room[you["nowon"]].length; i++) {
		neigh = room[you["nowon"]][i];
		corrid = "corr" + neigh + ":" + you["nowon"];
		if (corr.indexOf(corrid) < 0) {
			corrid = "corr" + you["nowon"] + ":" + neigh;
		}
		if (corr.indexOf(corrid) < 0) {
			return
		}
		if (!document.getElementById(corrid).classList.contains("built")) {
			if (com["nowon"] == neigh) {
				for (j = 0; j < room[com["nowon"]].length; j++) {
					comneigh = room[com["nowon"]][j];
					if (comneigh == you["nowon"]) {
						continue;
					}
					comcorrid = "corr" + comneigh + ":" + com["nowon"];
					if (corr.indexOf(comcorrid) < 0) {
						comcorrid = "corr" + com["nowon"] + ":" + comneigh;
					}
					if (corr.indexOf(comcorrid) < 0) {
						return
					}
					if (!document.getElementById(comcorrid).classList.contains("built")) {
						you["move"].push(comneigh);
						document.getElementById(comneigh).classList.add("move");
					}
				}
			} else {
				you["move"].push(neigh);
				document.getElementById(neigh).classList.add("move");
			}
		}
	}
}

function roommouseclick(element) {
	if (!playable) return;
	for (i = 0; i < you["move"].length; i++) {
		if (element.id == you["move"][i]) {
			updateyou(element.id);
			undocopy();
			gameinit = false;
			board.playerPos = {y:parseInt(element.id.substr(0, 1)), x:parseInt(element.id.substr(1, 1))}
			playable = false;
			document.getElementById("undobtn").disabled = true;
			var action = "Com";
			if (manual) action = "Man";
			var calc = parseInt(document.getElementById("calc").value)
			request_quoridor_data({action:action, board:board, calc:calc});
			return;
		}
	}
}
