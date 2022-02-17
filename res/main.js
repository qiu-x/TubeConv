function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

function post(url, data) {
	return fetch(url, {method: "POST", redirect: "follow", body: JSON.stringify(data)})
}

function log(text) {
	var caller = arguments.callee.caller
	if (caller != null) {
		console.log(caller.name.toString() + ": " + text)
	} else {
		console.log("main: " + text)
	}
}

async function searchFunc() {
	await clearSearch();
	txt = document.getElementById("search-input").value;
	await checkLink(txt);
	await moveSearch();
}

async function moveSearch() {
	search = document.getElementById("search");
	search.style.transform = 'translateY(-100%)';
	vids = document.getElementsByClassName("video");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "inline-block";
	}
	await sleep(700);
	for (let i = 0; i < vids.length; i++) {
		vids[i].classList.add("showed");
	}
}

async function clearSearch() {
	vids = document.getElementsByClassName("video");
	main_view = document.getElementById("main-view");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.opacity = 0;
	}
	await sleep(700);
	Array.from(vids).map(x => x.remove());
}

async function checkLink(txt) {
	const resp = await post("/req", {
		request: "check-link",
		link: txt
	});
	const json = await resp.json();
	switch(json.type) {
	case "title":
		await queryVideo(txt);
		break;
	case "link":
		// showVideoInfo(txt);
		break;
	default:
		log("bad response")
		break;
	}
}

async function queryVideo(title) {
	const resp = await post("/req", {
		request: "query",
		text: title
	});
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;
	console.log(json);
	for (let i = 0; i < json_videos.length; i++) {
		vid = template.cloneNode(true);
		vid.id = "video";
		vid.classList.remove("template");
		vid.classList.add("video");
		var san_element = document.createElement('div');
		san_element.innerText = json_videos[i].title;
		vid.getElementsByClassName("video-info").item(0).innerHTML = san_element.innerHTML;
		vid.getElementsByClassName("video-info").item(0).innerHTML += '</br>';
		vid.getElementsByClassName("video-info").item(0).innerHTML += '</br>';
		san_element.innerText = json_videos[i].author;
		vid.getElementsByClassName("video-info").item(0).innerHTML += san_element.innerHTML;
		vid.getElementsByClassName("thumb").item(0).src = json_videos[i].thumbnail;
		dl_button = vid.getElementsByClassName("download-button").item(0);
		// dl_button.onclick = () => {showVideoInfo(json_videos[i].link);};
		document.getElementById("main-view").appendChild(vid);
		console.log(vid);
	}
}
