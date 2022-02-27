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

function spinner(show, func) {
	let spinner = document.getElementById("spinner");
	if (show) {
		spinner.style.display = "inline-block";
		window.setTimeout(function(){
			spinner.style.opacity = 1;
		},0);
	} else {
		spinner.style.opacity = 0;
		spinner.addEventListener("transitionend", () => {
			spinner.style.display = "none";
			if (func) func();
		});
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
		search.style.transition = "margin-top 1s";
		search.style.marginTop = 0;
	}
	let vids = document.getElementsByClassName("video");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "flex";
	}
	let show = () => { spinner(false, () => {
		for (let i = 0; i < vids.length; i++) {
			vids[i].style.visibility = "visible";
			vids[i].style.opacity = 1;
		}
	})}
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
		spinner(true);
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
		spinner(true);
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
		dl_button.onclick = () => {showVideoInfo(i, json_videos[i].link);};
		document.getElementById("main-view").appendChild(vid);
	}
}

let videoo
async function showVideoInfo(index, link) {
	const resp = await post("/req", {
		request: "video-info",
		link: link
	});
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;
	console.log(json);
	elem = document.getElementsByClassName("video").item(index);
	videoo = elem;
	bg = elem.getElementsByClassName("alt-bg").item(0)
	bg.style.width = "100%";
	dl_button = elem.getElementsByClassName("download-button").item(0);
	dl_button.classList.add("final")
	arrow = dl_button.getElementsByTagName("img").item(0);
	arrow.classList.add("final");
	info = elem.querySelector(".video-info")
	old_h = getComputedStyle(info).height;
	info.innerHTML = document.getElementById("video-info-temp").innerHTML;
	info.style.height = old_h;
	info.style.opacity = 0;

	assign_text = (menu, arr) => {
		item_empty = menu.getElementsByClassName("item").item(0);
		for (let i = 0; i < arr.length; i++) {
			item = item_empty.cloneNode(true);
			item.innerText = arr[i];
			menu.appendChild(item);
		}
		item_empty.remove();
	}
	menus = Array.from(elem.getElementsByClassName("scrollmenu"));
	menu_info = [
		json.video_quality,
		json.audio_quality,
		["mp4", "webm", "mp3"]
	]

	for (let i = 0; i < 3; i++) {
		assign_text(menus[i], menu_info[i]);
	}

	bg.addEventListener("transitionend", () => {
		info.style.opacity = 1;
		// elem.querySelector(".video-info").style.display = "none";
		info.classList.add("final");
		console.log(info)
		info.classList.add("showed");
		window.setTimeout(function() {
			for (let i = 0; i < menus.length; i++) {
				menus[i].style.width = "30%";
			}
		},0);
	});
}
