var app = new Vue({
	el: '#goodtimes',
	data: { goodtimes: [], no_data: "No good times found." }
});

fetch('api/v1/goodtimes?o=json')
	.then(response => response.json())
	.then(data => app.goodtimes = data)
	.catch((error) => {
		console.error('Error:', error);
		app.goodtimes = [];
	});

document.querySelector('#beforeload').hidden = true;
document.querySelector('#goodtimes').hidden = false;
