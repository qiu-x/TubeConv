function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

function post(url, data) {
	return fetch(url, {method: "POST", redirect: "follow", body: JSON.stringify(data)})
}

function log(text) {
	let caller = arguments.callee.caller
	if (caller != null) {
		console.log(caller.name.toString() + ": " + text)
	} else {
		console.log("main: " + text)
	}
}

async function searchFunc() {
	await clearResoult();
	txt = document.getElementById("search-input").value;
	await checkLink(txt);
	await showResoult();
}

async function showResoult() {
	let search = document.getElementById("search");
	if (!showResoult.transformed) {
		search.style.transform = 'translateY(-100%)';
	}
	let vids = document.getElementsByClassName("video");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "inline-block";
	}
	let show = () => {
		for (let i = 0; i < vids.length; i++) {
			vids[i].style.visibility = "visible";
			vids[i].style.opacity = 1;

		}
	}
	if (!showResoult.transformed) {
		search.addEventListener("transitionend", show);
	} else {
		vids.item(0).addEventListener("transitionend", show);
	}
	showResoult.transformed = true;
}

async function clearResoult() {
	let vids = document.getElementsByClassName("video");
	if (!vids.item(0)) {
		return;
	}
	let main_view = document.getElementById("main-view");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.opacity = 0;
	}
	let arr = Array.from(vids);
	var now = Date.now();
	vids.item(0).addEventListener("transitionend", () => {
		arr.map(x => x.remove())
		console.log(arr)
		console.log(Date.now() - now);
	});
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
		let san_element = document.createElement('div');
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
	}
}
