var dateFormatter = new Intl.DateTimeFormat('en-US', { hour: "numeric", minute: "numeric" });

for (svg of document.querySelectorAll("svg")) {
	// Hack to make iOS call mouse move.
	svg.addEventListener("touchstart", ev => {}, {passive: true});

	// Handlers for touch or mouse.
	svg.addEventListener("mousemove", svgMove, {passive: true});
	svg.addEventListener("touchmove", touchMove, {passive: true});
}


function svgMove(ev) {
	let svg = ev.target;
	while (svg.tagName !== "svg") {
		svg = svg.parentElement;
	}

	updateGraph(svg, ev.clientX, ev.clientY);
}

function touchMove(ev) {
	let svg = ev.target;
	while (svg.tagName !== "svg") {
		svg = svg.parentElement;
	}

	updateGraph(svg, ev.touches[0].clientX, ev.touches[0].clientY);
}

function updateGraph(svg, x, y) {
	// The cursor point, translated into svg coordinates
	var pt = svg.createSVGPoint();
	pt.x = x;
	pt.y = y;
	var cursorpt =  pt.matrixTransform(svg.getScreenCTM().inverse());
	x = cursorpt.x;
	y = cursorpt.y;


	let spline = JSON.parse(svg.querySelector(".spline").innerHTML);
	let date = Number(svg.querySelector(".unixtime").innerHTML);
	let abs_t = xToTime(svg, date, x)
	let tideHeight = evalSpline(spline, date, abs_t); 
	let pretty_time = dateFormatter.format(abs_t*1000);

	let tt = gettooltip(svg);
	tt.innerHTML = "tide is " + tideHeight.toFixed(1) + " ft at " + pretty_time;
	tt.hidden = false;

	let dot = getdot(svg);
	dot.setAttribute("cx", x);
	dot.setAttribute("cy", heightToY(svg, tideHeight));
}

function gettooltip(svg) {
	return svg.parentElement.querySelector(".tooltip");
}

function getdot(svg) {
	let dot = svg.querySelector("#dot");
	if (!dot) {
		dot = document.createElementNS("http://www.w3.org/2000/svg", "circle");
		dot.id = "dot";
		dot.setAttribute("class", "dot");
		dot.setAttribute("r", "10");
		dot.setAttribute("fill-opacity", "0");
		svg.appendChild(dot);
	}
	return dot;
}

function evalSpline(spline, date, abs_t) {
	let n = spline.length;
	if (n === 0) {
		return NaN;
	}
	let mid = Math.floor(n/2);
	if (abs_t < spline[mid].start) {
		return evalSpline(spline.slice(0, mid), date, abs_t);
	} else if (abs_t > spline[mid].end) {
		return evalSpline(spline.slice(mid+1), date, abs_t);
	} else {
		return evalCurve(spline[mid], abs_t);
	}
}

function evalCurve(curve, abs_t) {
	let x = xrel(curve.start, abs_t);
	return curve.a*Math.pow(x,3) + curve.b*Math.pow(x, 2) + curve.c*x + curve.d;
}

function xrel(origin, abs_t) {
	return abs_t - origin;
}

function xToTime(svg, date, x) {
	const width = svg.viewBox.baseVal.width;
	let t = (x/width)*(24*60*60);
	let abs_t = date+t;
	return abs_t;
}

function heightToY(svg, tideHeight) {
	const height = svg.viewBox.baseVal.height;
	return height - Math.floor((tideHeight+2)*(height/10))
}
