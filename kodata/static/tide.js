var dateFormatter = new Intl.DateTimeFormat('en-US', { hour: "numeric", minute: "numeric" });

for (hoverEl of document.querySelectorAll(".goodtime_glance")) {
	// iOS touchmove doesn't work unless touchstart is also set..
	hoverEl.addEventListener("touchstart", touchMove, {capture: true, passive: true});

	// Handlers for touch or mouse.
	hoverEl.addEventListener("mousemove", svgMove, {capture: true, passive: true});
	hoverEl.addEventListener("touchmove", touchMove, {capture: true, passive: true});
}


function svgMove(ev) {
	let svg = findSVG(ev.target);

	updateGraph(svg, ev.clientX, ev.clientY);
}

function touchMove(ev) {
	let svg = findSVG(ev.target);

	updateGraph(svg, ev.touches[0].clientX, ev.touches[0].clientY);
}

// findSVG looks for the closest svg in the DOM to SVG.
function findSVG(el) {
	let svg = el.querySelector("svg");
	if (!svg) {
		return findSVG(el.parentElement);
	}
	return svg;
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
	tt.innerText = "tide is " + tideHeight.toFixed(1) + " ft at " + pretty_time;

	let dot = getdot(svg);
	let doty = svgTideY(svg, x);
	// let doty = heightToY(svg, tideHeight);
	dot.setAttribute("cx", x);
	dot.setAttribute("cy", doty);
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

function svgTideY(svg, x) {
	let allCurves = Array.from(svg.querySelectorAll(".tide")).map(tideControlPoints);
	for (curve of allCurves) {
		if (curve.points[0].x <= x && x <= curve.points[3].x) {
			return bezierYFromX(curve, x);
		}
	}
	return NaN;
}

function tideControlPoints(path) {
	let d = path.getAttribute("d");
	let flatPoints = Array.from(d.matchAll(/-?\d+/g), x => Number(x[0])).slice(0, 8);
	return {
		path: path,
		points: [
			{x: flatPoints[0], y: flatPoints[1]},
			{x: flatPoints[2], y: flatPoints[3]},
			{x: flatPoints[4], y: flatPoints[5]},
			{x: flatPoints[6], y: flatPoints[7]},
		],
	};
}

function bezierYFromX(curve, x) {
	let adjusted = curve.points.map((p, i) => ({x: i/3, y: p.x-x}));
	let t = utils.getRoots(adjusted)[0];
	const imgRect = curve.path.parentElement.viewBox.baseVal
	let length = (curve.path.getTotalLength()
		- (imgRect.height - curve.points[0].y) // Left edge.
		- (imgRect.height - curve.points[3].y) // Right edge.
		- (curve.points[3].x - curve.points[0].x)); // Bottom edge.
	t *= length;
	return curve.path.getPointAtLength(t).y;
}

const pi = Math.PI,
  tau = 2 * pi,
  quart = pi / 2,
  // float precision significant decimal
  epsilon = 0.000001

// math-inlining.
const { abs, cos, sin, acos, atan2, sqrt, pow  } = Math;

// cube root function yielding real roots
function crt(v) {
	return v < 0 ? -pow(-v, 1 / 3) : pow(v, 1 / 3);
}

// See https://github.com/Pomax/BezierInfo-2.
// This object is lifted from docs/js/graphics-element/lib/bezierjs/utils.js.
let utils = {
	// This function is "roots" in the original text.
	getRoots: function(points, line) {
		line = line || { p1: { x: 0, y: 0 }, p2: { x: 1, y: 0 } };

		const order = points.length - 1;
		const aligned = utils.align(points, line);
		const reduce = function (t) {
			return 0 <= t && t <= 1;
		};

		if (order === 2) {
			const a = aligned[0].y,
				b = aligned[1].y,
				c = aligned[2].y,
				d = a - 2 * b + c;
			if (d !== 0) {
				const m1 = -sqrt(b * b - a * c),
					m2 = -a + b,
					v1 = -(m1 + m2) / d,
					v2 = -(-m1 + m2) / d;
				return [v1, v2].filter(reduce);
			} else if (b !== c && d === 0) {
				return [(2 * b - c) / (2 * b - 2 * c)].filter(reduce);
			}
			return [];
		}

		// see http://www.trans4mind.com/personal_development/mathematics/polynomials/cubicAlgebra.htm
		const pa = aligned[0].y,
			pb = aligned[1].y,
			pc = aligned[2].y,
			pd = aligned[3].y;

		let d = -pa + 3 * pb - 3 * pc + pd,
			a = 3 * pa - 6 * pb + 3 * pc,
			b = -3 * pa + 3 * pb,
			c = pa;

		if (utils.approximately(d, 0)) {
			// this is not a cubic curve.
			if (utils.approximately(a, 0)) {
				// in fact, this is not a quadratic curve either.
				if (utils.approximately(b, 0)) {
					// in fact in fact, there are no solutions.
					return [];
				}
				// linear solution:
				return [-c / b].filter(reduce);
			}
			// quadratic solution:
			const q = sqrt(b * b - 4 * a * c),
				a2 = 2 * a;
			return [(q - b) / a2, (-b - q) / a2].filter(reduce);
		}

		// at this point, we know we need a cubic solution:

		a /= d;
		b /= d;
		c /= d;

		const p = (3 * b - a * a) / 3,
			p3 = p / 3,
			q = (2 * a * a * a - 9 * a * b + 27 * c) / 27,
			q2 = q / 2,
			discriminant = q2 * q2 + p3 * p3 * p3;

		let u1, v1, x1, x2, x3;
		if (discriminant < 0) {
			const mp3 = -p / 3,
				mp33 = mp3 * mp3 * mp3,
				r = sqrt(mp33),
				t = -q / (2 * r),
				cosphi = t < -1 ? -1 : t > 1 ? 1 : t,
				phi = acos(cosphi),
				crtr = crt(r),
				t1 = 2 * crtr;
			x1 = t1 * cos(phi / 3) - a / 3;
			x2 = t1 * cos((phi + tau) / 3) - a / 3;
			x3 = t1 * cos((phi + 2 * tau) / 3) - a / 3;
			return [x1, x2, x3].filter(reduce);
		} else if (discriminant === 0) {
			u1 = q2 < 0 ? crt(-q2) : -crt(q2);
			x1 = 2 * u1 - a / 3;
			x2 = -u1 - a / 3;
			return [x1, x2].filter(reduce);
		} else {
			const sd = sqrt(discriminant);
			u1 = crt(-q2 + sd);
			v1 = crt(q2 + sd);
			return [u1 - v1 - a / 3].filter(reduce);
		}
	},

	approximately: function(a, b, precision) {
		return abs(a - b) <= (precision || epsilon);
	},

	align: function (points, line) {
		const tx = line.p1.x,
			ty = line.p1.y,
			a = -atan2(line.p2.y - ty, line.p2.x - tx),
			d = function (v) {
				return {
					x: (v.x - tx) * cos(a) - (v.y - ty) * sin(a),
					y: (v.x - tx) * sin(a) + (v.y - ty) * cos(a),

				};
			};
		return points.map(d);
	},
}
