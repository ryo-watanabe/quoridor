function start() {
        var tb = document.getElementById("board");
        for (i = 0; i < 5; i++) {
                var tr = document.createElement("tr");
                for (j = 0; j < 5; j++) {
                        var td = document.createElement("td");
                        td.className = "room";
                        tr.appendChild(td)
                        if (j == 4) break;
                        var td = document.createElement("td");
                        td.className = "vcorr";
                        tr.appendChild(td)
                }
                tb.appendChild(tr);
                if (i == 4) break;
                var tr = document.createElement("tr");
                for (j = 0; j < 5; j++) {
                        var td = document.createElement("td");
                        td.className = "hcorr";
                        tr.appendChild(td)
                        if (j == 4) break;
                        var td = document.createElement("td");
                        td.className = "pole";
                        tr.appendChild(td)
                }
                tb.appendChild(tr);
        }
}

const wall = {
        "p00":["00:01", "p00", "10:11"],
        "p10":["10:11", "p10", "20:21"],
        "p20":["20:21", "p20", "30:31"],
        "p30":["30:31", "p30", "40:41"],
        "p01":["01:02", "p01", "11:12"],
        "p11":["11:12", "p11", "21:22"],
        "p21":["21:22", "p21", "31:32"],
        "p31":["31:32", "p31", "41:42"],
        "p02":["02:03", "p02", "12:13"],
        "p12":["12:13", "p12", "22:23"],
        "p22":["22:23", "p22", "32:33"],
        "p32":["32:33", "p32", "42:43"],
        "p03":["03:04", "p03", "13:14"],
        "p13":["13:14", "p13", "23:24"],
        "p23":["23:24", "p23", "33:34"],
        "p33":["33:34", "p33", "43:44"],
};

function mouseover(element) {
        drawwall(element, "#aaaaaa");
}

function mouseout(element) {
        drawwall(element, "#ffffff");
}

function drawwall(element, color) {
        wall[element.id].forEach( function(value) { setbg(document.getElementById(value), color); });
}

function setbg(element, color) {
        element.style.backgroundColor = color;
}
