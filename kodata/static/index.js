var app = new Vue({
	el: '#goodtimes',
	data: { goodtimes: [] }
})

fetch('api/v1/goodtimes?o=json')
  .then(response => response.json())
  .then(data => app.goodtimes = data);
