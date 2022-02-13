function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

function moveSearch() {
	search = document.getElementById("search");
	search.style.transform = 'translateY(-100%)';
	vids = document.getElementsByClassName("video")
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "inline-block";
		sleep(i*500).then(() => {
			vids[i].classList.add("showed");
		})
	}
}

