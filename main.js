function on_app_ready() {
	loadMainWindow('main');

	bindEvent('btnHello', 'clicked', onHelloClick);
}

function onHelloClick(evt) {
	alert(JSON.stringify(evt));
}