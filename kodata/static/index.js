var app = new Vue({
	el: '#goodtimes',
	data: { goodtimes: [], no_data: "No good times found." }
});

fetch('api/v2/goodtimes?o=json')
	.then(response => response.json())
	.then(data => {
		for (gt of data) {
			gt.open = false;
		}
		app.goodtimes = data;
	})
	.catch((error) => {
		console.error('Error:', error);
		app.goodtimes = [];
	});

document.querySelector('#beforeload').hidden = true;
document.querySelector('#goodtimes').hidden = false;
